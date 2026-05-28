package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// newSilentConverter returns a converter wired to a quiet logger so test
// output stays focused on assertions, not warning noise.
func newSilentConverter() Converter {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	return NewConverter(logger)
}

// roundtrip writes the given excelize file under tempDir, runs the full
// Excel → chunks → Excel pipeline, and returns the reconstructed file
// handle. The caller closes it.
func roundtrip(t *testing.T, tempDir string, build func(f *excelize.File)) *excelize.File {
	t.Helper()
	conv := newSilentConverter()

	src := filepath.Join(tempDir, "src.xlsx")
	f := excelize.NewFile()
	build(f)
	require.NoError(t, f.SaveAs(src))
	require.NoError(t, f.Close())

	require.NoError(t, conv.ExcelToJSONFile(src, src, ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
		ChunkingStrategy: "sheet-based",
	}))

	out := filepath.Join(tempDir, "out.xlsx")
	require.NoError(t, conv.JSONFileToExcel(src, out, ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
	}))

	rebuilt, err := excelize.OpenFile(out)
	require.NoError(t, err)
	return rebuilt
}

func TestRoundTrip_UnicodeSheetNames(t *testing.T) {
	tempDir := t.TempDir()

	unicodeSheets := []string{"日本語", "Ω-data", "naïve résumé"}

	rebuilt := roundtrip(t, tempDir, func(f *excelize.File) {
		// excelize seeds Sheet1; create the unicode-named ones and remove
		// the default placeholder so the sheet list matches exactly.
		for i, name := range unicodeSheets {
			idx, err := f.NewSheet(name)
			require.NoError(t, err)
			if i == 0 {
				f.SetActiveSheet(idx)
			}
			require.NoError(t, f.SetCellValue(name, "A1", "hello "+name))
		}
		require.NoError(t, f.DeleteSheet("Sheet1"))
	})
	defer rebuilt.Close()

	got := rebuilt.GetSheetList()
	sort.Strings(got)
	want := append([]string(nil), unicodeSheets...)
	sort.Strings(want)
	assert.Equal(t, want, got, "unicode sheet names must survive round-trip")

	for _, name := range unicodeSheets {
		val, err := rebuilt.GetCellValue(name, "A1")
		assert.NoError(t, err)
		assert.Equal(t, "hello "+name, val)
	}
}

func TestRoundTrip_FormulaErrorValues(t *testing.T) {
	tempDir := t.TempDir()

	// Formulas whose evaluation produces an Excel error value. We don't
	// need to evaluate them — we need to confirm the formula text itself
	// survives the JSON round-trip so reopening in Excel re-evaluates to
	// the same error.
	formulas := map[string]string{
		"A1": "=1/0",
		"A2": "=NA()",
		"A3": "=VLOOKUP(\"missing\",B1:B2,1,FALSE)",
		"A4": "=#REF!+1",
	}

	rebuilt := roundtrip(t, tempDir, func(f *excelize.File) {
		for cell, formula := range formulas {
			require.NoError(t, f.SetCellFormula("Sheet1", cell, formula))
		}
	})
	defer rebuilt.Close()

	for cell, want := range formulas {
		got, err := rebuilt.GetCellFormula("Sheet1", cell)
		assert.NoError(t, err)
		// excelize sometimes strips the leading "=" on read; compare
		// post-strip so either form passes.
		gotTrim := got
		if len(gotTrim) > 0 && gotTrim[0] == '=' {
			gotTrim = gotTrim[1:]
		}
		wantTrim := want
		if len(wantTrim) > 0 && wantTrim[0] == '=' {
			wantTrim = wantTrim[1:]
		}
		assert.Equal(t, wantTrim, gotTrim, "formula at %s should survive round-trip", cell)
	}
}

func TestRoundTrip_EmptyWorkbook(t *testing.T) {
	tempDir := t.TempDir()

	rebuilt := roundtrip(t, tempDir, func(f *excelize.File) {
		// excelize seeds Sheet1 with no cells — that's the empty case.
	})
	defer rebuilt.Close()

	sheets := rebuilt.GetSheetList()
	require.NotEmpty(t, sheets, "rebuilt empty workbook should still have at least one sheet")

	// Reading any cell on the surviving sheet must not error out.
	val, err := rebuilt.GetCellValue(sheets[0], "A1")
	assert.NoError(t, err)
	assert.Equal(t, "", val)

	// The chunk dir on disk should contain metadata + workbook + one
	// sheet chunk (or the empty-sheets edge case where only metadata
	// and workbook exist). Verify the metadata file exists at minimum.
	src := filepath.Join(tempDir, "src.xlsx")
	chunkDir := resolveChunkDir(src)
	_, err = os.Stat(filepath.Join(chunkDir, ".gitcells_chunks.json"))
	assert.NoError(t, err, "empty workbook should still produce chunk metadata")
}

func TestRoundTrip_DeterministicJSON(t *testing.T) {
	// Converting the same .xlsx twice must produce byte-identical
	// JSON chunk files. This is what makes diffs across machines (and
	// across the same machine on different days) actually useful —
	// without determinism, every commit shows spurious churn from map
	// iteration order, time.Now() fingerprints, or unsorted defined
	// names. Build a fixture exercising the historic offenders.
	tempDirA := t.TempDir()
	tempDirB := t.TempDir()
	conv := newSilentConverter()

	build := func() string {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.xlsx")
		f := excelize.NewFile()
		// Defined names — previously a map[string]string, randomized
		// per run. Include several with non-sorted keys to catch any
		// future regression to map iteration.
		for _, dn := range []struct{ name, refs string }{
			{"ZetaTotal", "Sheet1!A1"},
			{"AlphaCount", "Sheet1!B1"},
			{"MiddleBand", "Sheet1!C1"},
		} {
			require.NoError(t, f.SetDefinedName(&excelize.DefinedName{
				Name: dn.name, RefersTo: dn.refs, Scope: "Workbook",
			}))
		}
		// A couple of sheets to ensure sheet order is stable.
		for _, name := range []string{"Sheet1", "Beta", "Alpha"} {
			if name != "Sheet1" {
				_, err := f.NewSheet(name)
				require.NoError(t, err)
			}
			require.NoError(t, f.SetCellValue(name, "A1", "v-"+name))
		}
		require.NoError(t, f.SaveAs(src))
		require.NoError(t, f.Close())
		return src
	}

	// Build the same fixture twice — but reuse a single source file
	// across the two conversions so file mtime (which we now embed in
	// metadata) matches and the two outputs can be compared verbatim.
	src := build()

	convertInto := func(dst string) {
		t.Helper()
		require.NoError(t, os.MkdirAll(dst, 0o750))
		dstSrc := filepath.Join(dst, "src.xlsx")
		// Use a hard copy that preserves the source mtime — same byte
		// content, same mtime → same Metadata.Modified/Created, so any
		// remaining JSON difference is purely an order/encoding bug.
		raw, err := os.ReadFile(src) // #nosec G304 - test fixture
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dstSrc, raw, 0o644))
		srcInfo, err := os.Stat(src)
		require.NoError(t, err)
		require.NoError(t, os.Chtimes(dstSrc, srcInfo.ModTime(), srcInfo.ModTime()))
		require.NoError(t, conv.ExcelToJSONFile(dstSrc, dstSrc, ConvertOptions{
			PreserveFormulas: true,
			PreserveStyles:   true,
			ChunkingStrategy: "sheet-based",
		}))
	}

	convertInto(tempDirA)
	convertInto(tempDirB)

	collectChunks := func(root string) map[string][]byte {
		t.Helper()
		out := make(map[string][]byte)
		chunkRoot := filepath.Join(root, ".gitcells", "data")
		require.NoError(t, filepath.Walk(chunkRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(chunkRoot, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path) // #nosec G304 - under our tempdir
			if err != nil {
				return err
			}
			out[filepath.ToSlash(rel)] = data
			return nil
		}))
		return out
	}

	a := collectChunks(tempDirA)
	b := collectChunks(tempDirB)

	require.Equal(t, len(a), len(b), "same fixture must produce the same number of chunk files")
	for relPath, dataA := range a {
		dataB, ok := b[relPath]
		require.True(t, ok, "file %s present in run A but not B", relPath)
		assert.Equal(t, string(dataA), string(dataB),
			"file %s differs between two conversions of the same fixture — determinism regression", relPath)
	}

	// And one more concrete byte-stability assertion: chunk JSON files
	// must end with a newline so git diffs aren't perpetually noisy.
	for relPath, data := range a {
		if !strings.HasSuffix(relPath, ".json") {
			continue
		}
		require.NotEmpty(t, data, "%s is empty", relPath)
		assert.Equal(t, byte('\n'), data[len(data)-1],
			"chunk file %s must end with a newline", relPath)
	}
}

func TestRoundTrip_DateCellPreservesSerial(t *testing.T) {
	// A date-formatted cell must round-trip as a usable date, not as
	// the formatted display string. Before the date-detection fix, A1
	// came back as the string "1/1/20" after round-trip and =MONTH(A1)
	// evaluated to #VALUE!. The serial value (43831) and the date
	// number-format style must both survive.
	tempDir := t.TempDir()
	conv := newSilentConverter()

	src := filepath.Join(tempDir, "src.xlsx")
	f := excelize.NewFile()

	// 2020-01-01 = Excel serial 43831 in the 1900-epoch system.
	const serial = 43831
	require.NoError(t, f.SetCellValue("Sheet1", "A1", serial))

	// Apply a built-in date format (ID 14 = "mm-dd-yy") so date
	// detection fires via the built-in-ID branch.
	styleID, err := f.NewStyle(&excelize.Style{NumFmt: 14})
	require.NoError(t, err)
	require.NoError(t, f.SetCellStyle("Sheet1", "A1", "A1", styleID))

	require.NoError(t, f.SaveAs(src))
	require.NoError(t, f.Close())

	require.NoError(t, conv.ExcelToJSONFile(src, src, ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		ChunkingStrategy: "sheet-based",
	}))

	// Inspect the on-disk JSON to confirm the raw serial — not the
	// formatted display string — was stored, and the cell type is
	// "date".
	chunkDir := resolveChunkDir(src)
	sheetBytes, err := os.ReadFile(filepath.Join(chunkDir, "sheet_Sheet1.json"))
	require.NoError(t, err)
	var chunk SheetChunk
	require.NoError(t, json.Unmarshal(sheetBytes, &chunk))
	a1, ok := chunk.Sheet.Cells["A1"]
	require.True(t, ok, "A1 should be present in chunk")
	assert.Equal(t, models.CellTypeDate, a1.Type, "A1 must round-trip as a date type")
	asFloat, ok := a1.Value.(float64)
	require.True(t, ok, "A1.Value should be numeric in JSON, got %T (%v)", a1.Value, a1.Value)
	assert.InDelta(t, float64(serial), asFloat, 0.0001, "A1 must preserve the raw serial")

	// Round-trip back to xlsx and verify the cell is still numeric (so
	// downstream =MONTH(A1) etc. would evaluate correctly).
	out := filepath.Join(tempDir, "out.xlsx")
	require.NoError(t, conv.JSONFileToExcel(src, out, ConvertOptions{
		PreserveStyles: true,
	}))
	rebuilt, err := excelize.OpenFile(out)
	require.NoError(t, err)
	defer rebuilt.Close()
	rawAfter, err := rebuilt.GetCellValue("Sheet1", "A1", excelize.Options{RawCellValue: true})
	require.NoError(t, err)
	assert.Equal(t, "43831", rawAfter, "raw serial must survive .xlsx round-trip")
}

func TestIsDateFormat(t *testing.T) {
	// Built-in date format IDs (14-22 and 45-47) are unconditionally
	// dates.
	for _, id := range []int{14, 15, 18, 22, 45, 47} {
		assert.True(t, isDateFormat(id, ""), "built-in id %d must be flagged as date", id)
	}

	// Non-date built-in IDs should not be flagged.
	for _, id := range []int{0, 1, 2, 4, 9, 49} {
		assert.False(t, isDateFormat(id, ""), "built-in id %d must NOT be flagged as date", id)
	}

	// Custom formats with date/time tokens outside literal quotes.
	for _, fmt := range []string{
		"mm/dd/yyyy",
		"[$-409]m/d/yy h:mm AM/PM;@",
		"[h]:mm:ss",
		"yyyy-mm-dd",
		"d-mmm-yy",
	} {
		assert.True(t, isDateFormat(0, fmt), "custom format %q must be flagged as date", fmt)
	}

	// Custom formats whose date-token-shaped characters all live inside
	// quoted literals or brackets — must NOT be flagged.
	for _, fmt := range []string{
		`"meters"0.00`,
		`0"d"`,
		`#,##0.00`,
		`0.00E+00`,
		`"$"#,##0_);("$"#,##0)`,
	} {
		assert.False(t, isDateFormat(0, fmt), "custom format %q must NOT be flagged as date", fmt)
	}
}

func TestReadChunks_MissingSheetFileIsFatal(t *testing.T) {
	// A chunk file listed in metadata but absent on disk used to be
	// logged as a warning and skipped, producing a silently-truncated
	// workbook on round-trip. Confirm the new behavior is a loud error.
	tempDir := t.TempDir()
	conv := newSilentConverter()

	src := filepath.Join(tempDir, "src.xlsx")
	f := excelize.NewFile()
	idx, err := f.NewSheet("KeepMe")
	require.NoError(t, err)
	f.SetActiveSheet(idx)
	require.NoError(t, f.SetCellValue("KeepMe", "A1", "still here"))
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "also here"))
	require.NoError(t, f.SaveAs(src))
	require.NoError(t, f.Close())

	require.NoError(t, conv.ExcelToJSONFile(src, src, ConvertOptions{
		PreserveFormulas: true,
		ChunkingStrategy: "sheet-based",
	}))

	chunkDir := resolveChunkDir(src)

	// Remove one sheet file behind the converter's back. Metadata still
	// claims it exists. ReadChunks must refuse to proceed.
	require.NoError(t, os.Remove(filepath.Join(chunkDir, "sheet_KeepMe.json")))

	out := filepath.Join(tempDir, "out.xlsx")
	err = conv.JSONFileToExcel(src, out, ConvertOptions{})
	require.Error(t, err, "missing chunk file must surface as a fatal error")
	assert.Contains(t, err.Error(), "missing")
}

func TestSheetChunkFilename_SanitizationCollision(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	s := &SheetBasedChunking{logger: logger}

	a := s.sheetChunkFilename("A B")
	b := s.sheetChunkFilename("A_B")
	c := s.sheetChunkFilename("A_B") // identical input → identical output

	assert.NotEqual(t, a, b, `sanitized "A B" must not collide with literal "A_B"`)
	assert.Equal(t, b, c, "identical sheet names must produce identical chunk filenames")

	// The non-sanitized name takes the simple form; the sanitized name
	// carries a short content-hash suffix per chunking.go:402–409.
	assert.Equal(t, "sheet_A_B.json", b)
	assert.NotEqual(t, "sheet_A_B.json", a)
}
