package converter

import (
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
)

// ExcelToJSONFile converts Excel to chunked JSON files
func (c *converter) ExcelToJSONFile(inputPath, outputPath string, options ConvertOptions) error {
	// First, convert Excel to in-memory document
	doc, err := c.ExcelToJSON(inputPath, options)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "ExcelToJSONFile", inputPath, "failed to convert Excel to JSON")
	}

	// Always use chunking strategy for better git performance
	chunks, err := c.chunkingStrategy.WriteChunks(doc, outputPath, options)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "ExcelToJSONFile", outputPath, "failed to write chunks")
	}

	c.logger.Infof("Successfully wrote %d chunk files for %s", len(chunks), inputPath)
	return nil
}

// ReadStoredDocument hydrates the in-memory ExcelDocument from the chunk
// tree at basePath. It's the read half of the JSON→Excel pipeline,
// exposed so callers that only need the document (diff, status, future
// inspection commands) can skip the .xlsx synthesis step in
// JSONFileToExcel.
func (c *converter) ReadStoredDocument(basePath string) (*models.ExcelDocument, error) {
	return c.chunkingStrategy.ReadChunks(basePath)
}

// JSONFileToExcel converts chunked JSON files back to Excel
func (c *converter) JSONFileToExcel(inputPath, outputPath string, options ConvertOptions) error {
	// Read chunks
	doc, err := c.chunkingStrategy.ReadChunks(inputPath)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "JSONFileToExcel", inputPath, "failed to read chunks")
	}
	c.logger.Infof("Successfully read chunked JSON files from %s", inputPath)

	// Convert to Excel
	if err := c.JSONToExcel(doc, outputPath, options); err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "JSONFileToExcel", outputPath, "failed to convert JSON to Excel")
	}

	return nil
}

// GetExcelSheetNames returns the sheet names from an Excel file without processing the data
func (c *converter) GetExcelSheetNames(filePath string) ([]string, error) {
	f, err := openExcelWithRetry(filePath, excelOpenTimeout)
	if err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "GetExcelSheetNames", filePath, "failed to open Excel file")
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			c.logger.Warnf("Failed to close Excel file: %v", closeErr)
		}
	}()

	// Get sheet list
	sheetList := f.GetSheetList()
	return sheetList, nil
}

// GetChunkPaths returns the paths to the JSON chunk files for a given Excel file
func (c *converter) GetChunkPaths(basePath string) ([]string, error) {
	return c.chunkingStrategy.GetChunkPaths(basePath)
}
