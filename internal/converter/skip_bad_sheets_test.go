package converter

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// buildTwoSheetFixture creates a real xlsx with two readable sheets and
// returns its on-disk path. Used as the input to corruptSheetXML.
func buildTwoSheetFixture(t *testing.T) string {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Sheet1 (default) has cells.
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "good"))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", 42))

	// Sheet2 also has cells.
	_, err := f.NewSheet("Sheet2")
	require.NoError(t, err)
	require.NoError(t, f.SetCellValue("Sheet2", "A1", "will-be-broken"))

	path := filepath.Join(t.TempDir(), "two_sheet.xlsx")
	require.NoError(t, f.SaveAs(path))
	return path
}

// removeZipEntry rewrites the xlsx zip in place with the named entry
// dropped. workbook.xml still references the sheet, but its XML payload
// is gone — GetRows("Sheet2") then errors out, exercising the
// processSheet failure branch in ExcelToJSON.
func removeZipEntry(t *testing.T, xlsxPath, entryName string) {
	t.Helper()

	src, err := zip.OpenReader(xlsxPath)
	require.NoError(t, err)

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	matched := false
	for _, f := range src.File {
		if f.Name == entryName {
			matched = true
			continue
		}
		hdr := f.FileHeader
		out, err := w.CreateHeader(&hdr)
		require.NoError(t, err)

		rc, err := f.Open()
		require.NoError(t, err)
		_, err = io.Copy(out, rc)
		require.NoError(t, err)
		_ = rc.Close()
	}
	require.NoError(t, w.Close())
	require.NoError(t, src.Close())
	require.True(t, matched, "expected to find entry %q in xlsx", entryName)

	require.NoError(t, os.WriteFile(xlsxPath, buf.Bytes(), 0o644))
}

func TestExcelToJSON_BadSheetPolicy(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	t.Run("default aborts when a sheet fails to read", func(t *testing.T) {
		path := buildTwoSheetFixture(t)
		removeZipEntry(t, path, "xl/worksheets/sheet2.xml")

		doc, err := conv.ExcelToJSON(path, ConvertOptions{})
		assert.Error(t, err)
		assert.Nil(t, doc)
		assert.Contains(t, err.Error(), "Sheet2")
	})

	t.Run("SkipBadSheets keeps the good sheet and drops the bad one", func(t *testing.T) {
		path := buildTwoSheetFixture(t)
		removeZipEntry(t, path, "xl/worksheets/sheet2.xml")

		doc, err := conv.ExcelToJSON(path, ConvertOptions{SkipBadSheets: true})
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Len(t, doc.Sheets, 1, "only Sheet1 should survive")
		assert.Equal(t, "Sheet1", doc.Sheets[0].Name)
	})

	t.Run("happy path is unaffected when SkipBadSheets is true", func(t *testing.T) {
		path := buildTwoSheetFixture(t)

		doc, err := conv.ExcelToJSON(path, ConvertOptions{SkipBadSheets: true})
		require.NoError(t, err)
		require.NotNil(t, doc)
		require.Len(t, doc.Sheets, 2)
	})
}
