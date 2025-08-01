package converter

import (
	"fmt"
	
	"github.com/xuri/excelize/v2"
)

// ExcelToJSONFile converts Excel to chunked JSON files
func (c *converter) ExcelToJSONFile(inputPath, outputPath string, options ConvertOptions) error {
	// First, convert Excel to in-memory document
	doc, err := c.ExcelToJSON(inputPath, options)
	if err != nil {
		return fmt.Errorf("failed to convert Excel to JSON: %w", err)
	}

	// Always use chunking strategy for better git performance
	chunks, err := c.chunkingStrategy.WriteChunks(doc, outputPath, options)
	if err != nil {
		return fmt.Errorf("failed to write chunks: %w", err)
	}

	c.logger.Infof("Successfully wrote %d chunk files for %s", len(chunks), inputPath)
	return nil
}

// JSONFileToExcel converts chunked JSON files back to Excel
func (c *converter) JSONFileToExcel(inputPath, outputPath string, options ConvertOptions) error {
	// Read chunks
	doc, err := c.chunkingStrategy.ReadChunks(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read chunks: %w", err)
	}
	c.logger.Infof("Successfully read chunked JSON files from %s", inputPath)

	// Convert to Excel
	if err := c.JSONToExcel(doc, outputPath, options); err != nil {
		return fmt.Errorf("failed to convert JSON to Excel: %w", err)
	}

	return nil
}

// GetExcelSheetNames returns the sheet names from an Excel file without processing the data
func (c *converter) GetExcelSheetNames(filePath string) ([]string, error) {
	// This method is implemented in excel_to_json.go, but since Go doesn't allow forward declarations,
	// we'll implement it here to avoid circular calls
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
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
