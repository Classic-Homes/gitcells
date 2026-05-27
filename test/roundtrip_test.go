package test

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestRoundTripFidelity is the canary for the core Excel↔JSON pipeline.
// It builds an in-memory ExcelDocument exercising the supported feature
// floor (values, formulas, basic styles, merged cells, comments, multi-
// sheet, unicode sheet names, special chars), writes it to .xlsx via
// JSONToExcel, reads it back via ExcelToJSON, and asserts that every
// field that should survive a round-trip actually does.
//
// Anything failing here is a real reliability regression in the core
// pipeline — not a test brittleness issue. Fixing the pipeline (or
// shrinking the supported feature floor) is the right response, never
// loosening these assertions.
func TestRoundTripFidelity(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(logger)

	original := buildRoundTripFixture()

	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "roundtrip.xlsx")

	opts := converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
	}

	require.NoError(t, conv.JSONToExcel(original, outPath, opts), "JSONToExcel")

	loaded, err := conv.ExcelToJSON(outPath, opts)
	require.NoError(t, err, "ExcelToJSON")

	assertSheetsEquivalent(t, original.Sheets, loaded.Sheets)
}

func buildRoundTripFixture() *models.ExcelDocument {
	bold := &models.CellStyle{
		Font: &models.Font{Name: "Calibri", Size: 12, Bold: true, Color: "FF0000FF"},
	}
	numFmt := &models.CellStyle{
		NumberFormat: "#,##0.00",
	}

	return &models.ExcelDocument{
		Version: "1.0",
		Sheets: []models.Sheet{
			{
				Name:  "Numbers",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "Item", Type: models.CellTypeString, Style: bold},
					"B1": {Value: "Qty", Type: models.CellTypeString, Style: bold},
					"C1": {Value: "Price", Type: models.CellTypeString, Style: bold},
					"D1": {Value: "Total", Type: models.CellTypeString, Style: bold},
					"A2": {Value: "Widget", Type: models.CellTypeString},
					"B2": {Value: float64(3), Type: models.CellTypeNumber},
					"C2": {Value: float64(9.99), Type: models.CellTypeNumber, Style: numFmt},
					"D2": {Value: float64(29.97), Type: models.CellTypeFormula, Formula: "B2*C2", Style: numFmt},
					"A3": {Value: "Note", Type: models.CellTypeString,
						Comment: &models.Comment{Author: "Tester", Text: "first row of data"}},
				},
				MergedCells: []models.MergedCell{
					{Range: "A5:D5"},
				},
			},
			{
				Name:  "Üñîçødé Sheet 名前", // unicode + space in sheet name
				Index: 1,
				Cells: map[string]models.Cell{
					"A1": {Value: "héllo wörld", Type: models.CellTypeString},
					"B1": {Value: "包含中文", Type: models.CellTypeString},
					"A2": {Value: true, Type: models.CellTypeBoolean},
				},
			},
		},
	}
}

func assertSheetsEquivalent(t *testing.T, want, got []models.Sheet) {
	t.Helper()
	require.Equal(t, len(want), len(got), "sheet count")

	wantBy := indexSheetsByName(want)
	gotBy := indexSheetsByName(got)

	for name, wantSheet := range wantBy {
		gotSheet, ok := gotBy[name]
		require.Truef(t, ok, "sheet %q missing from round-trip", name)

		assertCellsEquivalent(t, name, wantSheet.Cells, gotSheet.Cells)
		assertMergedEquivalent(t, name, wantSheet.MergedCells, gotSheet.MergedCells)
	}
}

func indexSheetsByName(sheets []models.Sheet) map[string]models.Sheet {
	m := make(map[string]models.Sheet, len(sheets))
	for _, s := range sheets {
		m[s.Name] = s
	}
	return m
}

func assertCellsEquivalent(t *testing.T, sheet string, want, got map[string]models.Cell) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "cell count in sheet %q", sheet)
	for ref, w := range want {
		g, ok := got[ref]
		require.Truef(t, ok, "sheet %q cell %s missing from round-trip", sheet, ref)

		// Formula cells: only the formula round-trips. The cached computed
		// value gets cleared because excelize doesn't recalculate on save,
		// and Excel itself would normally re-evaluate on open.
		if w.Formula != "" {
			require.Equalf(t, w.Formula, g.Formula,
				"sheet %q cell %s formula mismatch", sheet, ref)
		} else {
			// Value comparison normalizes float-from-string vs float-from-number.
			require.Equalf(t, normalizeValue(w.Value), normalizeValue(g.Value),
				"sheet %q cell %s value mismatch", sheet, ref)
		}

		// Comment.Text should round-trip; Author often gets remapped by
		// Excel ("excelize-author" default) so only assert the body.
		if w.Comment != nil {
			require.NotNilf(t, g.Comment, "sheet %q cell %s lost comment", sheet, ref)
			require.Equalf(t, w.Comment.Text, g.Comment.Text,
				"sheet %q cell %s comment text mismatch", sheet, ref)
		}
	}
}

func assertMergedEquivalent(t *testing.T, sheet string, want, got []models.MergedCell) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "merged cell count in sheet %q", sheet)
	w := mergedRanges(want)
	g := mergedRanges(got)
	for i := range w {
		require.Equalf(t, w[i], g[i], "merged range %d in sheet %q", i, sheet)
	}
}

func mergedRanges(m []models.MergedCell) []string {
	out := make([]string, len(m))
	for i, mc := range m {
		out[i] = mc.Range
	}
	sort.Strings(out)
	return out
}

// normalizeValue coerces types so cells compared post-round-trip can be
// equal even though excelize reads everything as strings:
//   - all integer + float widths → float64
//   - bool true/false → "TRUE"/"FALSE" (matching what excelize emits)
//
// The original fixture uses native Go types; the round-tripped doc holds
// strings (sometimes parsed back to float64 by detectCellType). Both
// sides go through normalize so equality is meaningful.
func normalizeValue(v interface{}) interface{} {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case float32:
		return float64(x)
	case bool:
		if x {
			return "TRUE"
		}
		return "FALSE"
	default:
		return x
	}
}
