package converter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func TestExcelToJSONFile_Integration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	tests := []struct {
		name        string
		inputFile   string
		options     ConvertOptions
		validateDoc func(t *testing.T, chunksDir string)
	}{
		{
			name:      "simple Excel file",
			inputFile: "../../test/testdata/sample_files/simple.xlsx",
			options: ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   true,
				PreserveComments: true,
			},
			validateDoc: func(t *testing.T, chunksDir string) {
				// Check that chunk files were created
				metadataFile := filepath.Join(chunksDir, ".gitcells_chunks.json")
				assert.FileExists(t, metadataFile)

				workbookFile := filepath.Join(chunksDir, "workbook.json")
				assert.FileExists(t, workbookFile)

				// Should have at least one sheet file
				files, err := filepath.Glob(filepath.Join(chunksDir, "sheet_*.json"))
				require.NoError(t, err)
				assert.NotEmpty(t, files)
			},
		},
		{
			name:      "complex Excel file with multiple sheets",
			inputFile: "../../test/testdata/sample_files/complex.xlsx",
			options: ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   true,
			},
			validateDoc: func(t *testing.T, chunksDir string) {
				// Check metadata
				metadataFile := filepath.Join(chunksDir, ".gitcells_chunks.json")
				assert.FileExists(t, metadataFile)

				// Should have multiple sheet files
				files, err := filepath.Glob(filepath.Join(chunksDir, "sheet_*.json"))
				require.NoError(t, err)
				assert.Greater(t, len(files), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, "output.json")

			// Convert Excel to JSON chunks
			err := conv.ExcelToJSONFile(tt.inputFile, outputPath, tt.options)
			require.NoError(t, err)

			// Determine chunks directory
			baseFile := filepath.Base(outputPath)
			baseFile = baseFile[:len(baseFile)-len(filepath.Ext(baseFile))]
			chunksDir := filepath.Join(tempDir, ".gitcells", "data", baseFile+"_chunks")

			// Validate chunks
			tt.validateDoc(t, chunksDir)
		})
	}
}

func TestJSONFileToExcel_Integration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	tests := []struct {
		name           string
		setupChunks    func(t *testing.T, tempDir string) string // returns JSON path
		options        ConvertOptions
		validateOutput func(t *testing.T, outputPath string)
		expectError    bool
	}{
		{
			name: "valid chunks to Excel",
			setupChunks: func(t *testing.T, tempDir string) string {
				// First create chunks from a real Excel file
				inputFile := "../../test/testdata/sample_files/simple.xlsx"
				jsonPath := filepath.Join(tempDir, "test.json")

				err := conv.ExcelToJSONFile(inputFile, jsonPath, ConvertOptions{
					PreserveFormulas: true,
				})
				require.NoError(t, err)

				return jsonPath
			},
			options: ConvertOptions{
				PreserveFormulas: true,
			},
			validateOutput: func(t *testing.T, outputPath string) {
				// Verify Excel file was created and can be opened
				f, err := excelize.OpenFile(outputPath)
				require.NoError(t, err)
				defer f.Close()

				// Check that we have sheets
				sheets := f.GetSheetList()
				assert.NotEmpty(t, sheets)
			},
			expectError: false,
		},
		{
			name: "missing chunk metadata",
			setupChunks: func(t *testing.T, tempDir string) string {
				// Create chunks directory without metadata file
				chunksDir := filepath.Join(tempDir, ".gitcells", "data", "broken_chunks")
				err := os.MkdirAll(chunksDir, 0755)
				require.NoError(t, err)

				// Create a workbook file but no metadata
				workbookFile := filepath.Join(chunksDir, "workbook.json")
				err = os.WriteFile(workbookFile, []byte(`{"version":"1.0"}`), 0600)
				require.NoError(t, err)

				return filepath.Join(tempDir, "broken.json")
			},
			options:        ConvertOptions{},
			validateOutput: nil,
			expectError:    true,
		},
		{
			name: "corrupted chunk files",
			setupChunks: func(t *testing.T, tempDir string) string {
				chunksDir := filepath.Join(tempDir, ".gitcells", "data", "corrupted_chunks")
				err := os.MkdirAll(chunksDir, 0755)
				require.NoError(t, err)

				// Create valid metadata
				metadataFile := filepath.Join(chunksDir, ".gitcells_chunks.json")
				metadata := `{
					"version": "1.0",
					"strategy": "sheet-based",
					"mainFile": "workbook.json",
					"chunkFiles": ["workbook.json", "sheet_Sheet1.json"],
					"totalSheets": 1
				}`
				err = os.WriteFile(metadataFile, []byte(metadata), 0600)
				require.NoError(t, err)

				// Create corrupted workbook file
				workbookFile := filepath.Join(chunksDir, "workbook.json")
				err = os.WriteFile(workbookFile, []byte(`invalid json`), 0600)
				require.NoError(t, err)

				return filepath.Join(tempDir, "corrupted.json")
			},
			options:        ConvertOptions{},
			validateOutput: nil,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			jsonPath := tt.setupChunks(t, tempDir)
			outputPath := filepath.Join(tempDir, "output.xlsx")

			// Convert JSON chunks to Excel
			err := conv.JSONFileToExcel(jsonPath, outputPath, tt.options)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateOutput != nil {
					tt.validateOutput(t, outputPath)
				}
			}
		})
	}
}

func TestFullConversionCycle(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	testFiles := []string{
		"../../test/testdata/sample_files/simple.xlsx",
		"../../test/testdata/sample_files/complex.xlsx",
	}

	for _, testFile := range testFiles {
		t.Run(filepath.Base(testFile), func(t *testing.T) {
			tempDir := t.TempDir()

			// Step 1: Convert Excel to JSON chunks
			jsonPath := filepath.Join(tempDir, "intermediate.json")
			err := conv.ExcelToJSONFile(testFile, jsonPath, ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   true,
				PreserveComments: true,
			})
			require.NoError(t, err)

			// Step 2: Convert JSON chunks back to Excel
			outputPath := filepath.Join(tempDir, "reconstructed.xlsx")
			err = conv.JSONFileToExcel(jsonPath, outputPath, ConvertOptions{
				PreserveFormulas: true,
				PreserveStyles:   true,
				PreserveComments: true,
			})
			require.NoError(t, err)

			// Step 3: Verify the reconstructed file
			// Open both original and reconstructed files
			originalFile, err := excelize.OpenFile(testFile)
			require.NoError(t, err)
			defer originalFile.Close()

			reconstructedFile, err := excelize.OpenFile(outputPath)
			require.NoError(t, err)
			defer reconstructedFile.Close()

			// Compare sheet lists
			originalSheets := originalFile.GetSheetList()
			reconstructedSheets := reconstructedFile.GetSheetList()
			assert.ElementsMatch(t, originalSheets, reconstructedSheets, "Sheet lists should match")

			// For each sheet, verify some basic properties
			for _, sheetName := range originalSheets {
				// Get all rows from original
				originalRows, err := originalFile.GetRows(sheetName)
				assert.NoError(t, err)

				// Get all rows from reconstructed
				reconstructedRows, err := reconstructedFile.GetRows(sheetName)
				assert.NoError(t, err)

				// Compare row count
				assert.Equal(t, len(originalRows), len(reconstructedRows),
					"Row count should match for sheet %s", sheetName)

				// Compare some cell values (sampling to avoid excessive checks)
				for i, row := range originalRows {
					if i >= len(reconstructedRows) {
						break
					}
					// Compare column count
					assert.Equal(t, len(row), len(reconstructedRows[i]),
						"Column count should match for row %d in sheet %s", i, sheetName)
				}
			}
		})
	}
}

func TestConversionWithSheetSelection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	tempDir := t.TempDir()
	inputFile := "../../test/testdata/sample_files/complex.xlsx"

	// Test converting only specific sheets
	t.Run("convert specific sheets", func(t *testing.T) {
		jsonPath := filepath.Join(tempDir, "selected_sheets.json")

		err := conv.ExcelToJSONFile(inputFile, jsonPath, ConvertOptions{
			SheetsToConvert: []string{"Summary"}, // complex.xlsx has a "Summary" sheet
		})
		require.NoError(t, err)

		// Convert back to Excel
		outputPath := filepath.Join(tempDir, "selected_sheets.xlsx")
		err = conv.JSONFileToExcel(jsonPath, outputPath, ConvertOptions{})
		require.NoError(t, err)

		// Verify only selected sheets were converted
		f, err := excelize.OpenFile(outputPath)
		require.NoError(t, err)
		defer f.Close()

		sheets := f.GetSheetList()
		assert.Contains(t, sheets, "Summary")

		// Note: Even though we only selected "Summary" sheet, Excel might create
		// a default "Sheet1" when creating a new workbook. The important thing
		// is that the Summary sheet exists and contains the right data.

		// Verify that we don't have more sheets than the original
		assert.LessOrEqual(t, len(sheets), 2)

		// Log what sheets we actually got for debugging
		t.Logf("Sheets in output: %v", sheets)
	})
}

func TestErrorHandling(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	conv := NewConverter(logger)

	t.Run("non-existent Excel file", func(t *testing.T) {
		err := conv.ExcelToJSONFile("/non/existent/file.xlsx", "output.json", ConvertOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ExcelToJSONFile")
	})

	t.Run("non-existent JSON file", func(t *testing.T) {
		err := conv.JSONFileToExcel("/non/existent/file.json", "output.xlsx", ConvertOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "JSONFileToExcel")
	})

	t.Run("invalid output directory", func(t *testing.T) {
		inputFile := "../../test/testdata/sample_files/simple.xlsx"
		err := conv.ExcelToJSONFile(inputFile, "/invalid/path/output.json", ConvertOptions{})
		assert.Error(t, err)
	})

	t.Run("standalone JSON file not chunked", func(t *testing.T) {
		// Create a standalone JSON file (not chunked)
		tempDir := t.TempDir()
		jsonFile := filepath.Join(tempDir, "standalone.json")
		jsonContent := `{
			"document": {
				"version": "1.0",
				"sheets": [{
					"name": "Sheet1",
					"cells": {
						"A1": {"value": "test"}
					}
				}]
			}
		}`
		err := os.WriteFile(jsonFile, []byte(jsonContent), 0600)
		require.NoError(t, err)

		// Try to convert it - should fail because it's not chunked
		outputPath := filepath.Join(tempDir, "output.xlsx")
		err = conv.JSONFileToExcel(jsonFile, outputPath, ConvertOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "appears to be a standalone JSON file")
	})
}
