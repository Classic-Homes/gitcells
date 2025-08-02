package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/sirupsen/logrus"
)

// ConverterAdapter bridges the TUI with the converter package
type ConverterAdapter struct {
	converter converter.Converter
	options   converter.ConvertOptions
}

func NewConverterAdapter() *ConverterAdapter {
	logger := logrus.New()
	return &ConverterAdapter{
		converter: converter.NewConverter(logger),
		options: converter.ConvertOptions{
			PreserveFormulas: true,
			PreserveStyles:   true,
			PreserveComments: true,
			CompactJSON:      false,
			IgnoreEmptyCells: true,
		},
	}
}

// ConvertFile converts a single Excel file to JSON
func (ca *ConverterAdapter) ConvertFile(excelPath string) (*ConversionResult, error) {
	jsonPath := GetJSONPath(excelPath)

	err := ca.converter.ExcelToJSONFile(excelPath, jsonPath, ca.options)
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		ExcelPath: excelPath,
		JSONPath:  jsonPath,
		Success:   true,
	}, nil
}

// ConvertFileWithSheetOptions converts a single Excel file to JSON with sheet selection options
func (ca *ConverterAdapter) ConvertFileWithSheetOptions(excelPath string, sheetOptions SheetSelectionOptions) (*ConversionResult, error) {
	jsonPath := GetJSONPath(excelPath)

	// Create options with sheet selection
	options := ca.options
	options.SheetsToConvert = sheetOptions.SheetsToConvert
	options.ExcludeSheets = sheetOptions.ExcludeSheets
	options.SheetIndices = sheetOptions.SheetIndices

	err := ca.converter.ExcelToJSONFile(excelPath, jsonPath, options)
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		ExcelPath: excelPath,
		JSONPath:  jsonPath,
		Success:   true,
	}, nil
}

// ConvertJSONToExcel converts a JSON file back to Excel
func (ca *ConverterAdapter) ConvertJSONToExcel(jsonPath string) (*ConversionResult, error) {
	excelPath := GetExcelPath(jsonPath)

	err := ca.converter.JSONFileToExcel(jsonPath, excelPath, ca.options)
	if err != nil {
		return nil, err
	}

	return &ConversionResult{
		ExcelPath: excelPath,
		JSONPath:  jsonPath,
		Success:   true,
	}, nil
}

// GetPendingConversions returns list of Excel files that need conversion
func (ca *ConverterAdapter) GetPendingConversions(directory string, pattern string) ([]string, error) {
	// Find Excel files matching pattern
	files, err := filepath.Glob(filepath.Join(directory, pattern))
	if err != nil {
		return nil, err
	}

	pending := []string{}
	for _, file := range files {
		// Check if JSON version exists and is up to date
		jsonPath := GetJSONPath(file)
		if !IsUpToDate(file, jsonPath) {
			pending = append(pending, file)
		}
	}

	return pending, nil
}

// GetConversionStats returns statistics about conversions
func (ca *ConverterAdapter) GetConversionStats(directory string) (*ConversionStats, error) {
	// This would analyze the directory for conversion statistics
	// For now, return mock data
	return &ConversionStats{
		TotalExcelFiles:    15,
		ConvertedFiles:     12,
		PendingConversions: 3,
		FailedConversions:  0,
		TotalJSONSize:      1024 * 1024 * 5, // 5MB
	}, nil
}

// ConversionResult contains the result of a conversion operation
type ConversionResult struct {
	ExcelPath string
	JSONPath  string
	Success   bool
	Error     error
}

// ConversionStats contains statistics about conversions
type ConversionStats struct {
	TotalExcelFiles    int
	ConvertedFiles     int
	PendingConversions int
	FailedConversions  int
	TotalJSONSize      int64
}

// SheetSelectionOptions contains options for selecting specific sheets
type SheetSelectionOptions struct {
	SheetsToConvert []string // Specific sheet names to convert
	ExcludeSheets   []string // Sheet names to exclude
	SheetIndices    []int    // Specific sheet indices to convert (0-based)
}

// SheetInfo contains information about a sheet in an Excel file
type SheetInfo struct {
	Name  string
	Index int
}

// GetExcelSheets returns a list of sheet names and indices from an Excel file
func (ca *ConverterAdapter) GetExcelSheets(excelPath string) ([]SheetInfo, error) {
	// Use the optimized method to get sheet names without processing data
	sheetNames, err := ca.converter.GetExcelSheetNames(excelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Excel file: %w", err)
	}

	sheets := make([]SheetInfo, len(sheetNames))
	for i, name := range sheetNames {
		sheets[i] = SheetInfo{
			Name:  name,
			Index: i,
		}
	}

	return sheets, nil
}

// ValidatePattern checks if a file pattern is valid
func ValidatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Test the pattern
	_, err := filepath.Match(pattern, "test.xlsx")
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	return nil
}

// GetJSONPath returns the JSON output path for an Excel file
func GetJSONPath(excelPath string) string {
	// For conversion, we use the Excel path as the base
	// The converter will handle creating the proper chunk directory
	return excelPath
}

// GetExcelPath returns the Excel output path for a JSON chunk directory
func GetExcelPath(jsonChunkPath string) string {
	// If it's a chunk directory path, extract the original filename
	if strings.Contains(jsonChunkPath, ".gitcells/data/") && strings.HasSuffix(jsonChunkPath, "_chunks") {
		// Extract the original filename from chunk directory name
		base := filepath.Base(jsonChunkPath)
		originalName := strings.TrimSuffix(base, "_chunks")

		// For simplicity, place the Excel file in the current working directory
		return originalName
	}

	// Fallback for legacy paths
	dir := filepath.Dir(jsonChunkPath)
	base := filepath.Base(jsonChunkPath)
	nameWithoutExt := strings.TrimSuffix(base, ".json")
	return filepath.Join(dir, nameWithoutExt+".xlsx")
}

// IsUpToDate checks if the JSON chunks are up to date with the Excel file
func IsUpToDate(excelPath, jsonPath string) bool {
	excelInfo, err := os.Stat(excelPath)
	if err != nil {
		return false
	}

	// Check for chunk directory in .gitcells/data
	cwd, _ := os.Getwd()
	gitRoot := findGitRoot(cwd)
	relPath, _ := filepath.Rel(gitRoot, filepath.Dir(excelPath))
	chunkDir := filepath.Join(gitRoot, ".gitcells", "data", relPath, filepath.Base(excelPath)+"_chunks")

	// Check if chunk metadata exists
	metadataPath := filepath.Join(chunkDir, ".gitcells_chunks.json")
	metadataInfo, err := os.Stat(metadataPath)
	if err != nil {
		return false
	}

	return metadataInfo.ModTime().After(excelInfo.ModTime())
}

// findGitRoot finds the git repository root, or returns the current directory
func findGitRoot(startDir string) string {
	dir := startDir
	for {
		// Check if .git directory exists
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir
		}

		// Check if we've reached the root
		parent := filepath.Dir(dir)
		if parent == dir {
			// Return the original directory if no git root found
			return startDir
		}
		dir = parent
	}
}
