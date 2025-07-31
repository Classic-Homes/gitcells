package converter

import (
	"fmt"
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
