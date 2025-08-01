// Package converter handles Excel to JSON conversion and vice versa.
package converter

import (
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
)

type Converter interface {
	// In-memory operations (used internally)
	ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error)
	JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error

	// File-based operations with automatic chunking
	ExcelToJSONFile(inputPath, outputPath string, options ConvertOptions) error
	JSONFileToExcel(inputPath, outputPath string, options ConvertOptions) error

	// Utility operations
	GetExcelSheetNames(filePath string) ([]string, error)
	GetChunkPaths(basePath string) ([]string, error)
}

type ConvertOptions struct {
	PreserveFormulas           bool
	PreserveStyles             bool
	PreserveComments           bool
	PreserveCharts             bool // New: Extract chart information
	PreservePivotTables        bool // New: Extract pivot table structure
	PreserveDataValidation     bool // Extract data validation rules
	PreserveConditionalFormats bool // Extract conditional formatting
	PreserveRichText           bool // Extract rich text formatting
	PreserveTables             bool // Extract Excel tables
	CompactJSON                bool
	IgnoreEmptyCells           bool
	MaxCellsPerSheet           int                                    // Prevent memory issues with huge files
	ProgressCallback           func(stage string, current, total int) // Progress reporting
	ShowProgressBar            bool                                   // Enable built-in progress display
	ChunkingStrategy           string                                 // "sheet-based" or "hybrid" (future), defaults to "sheet-based"

	// Sheet selection options
	SheetsToConvert []string // Specific sheet names to convert (empty = convert all)
	ExcludeSheets   []string // Sheet names to exclude from conversion
	SheetIndices    []int    // Specific sheet indices to convert (0-based, empty = convert all)
}

type converter struct {
	logger           *logrus.Logger
	chunkingStrategy ChunkingStrategy
}

func NewConverter(logger *logrus.Logger) Converter {
	c := &converter{
		logger:           logger,
		chunkingStrategy: nil,
	}
	c.chunkingStrategy = NewSheetBasedChunking(c.logger)
	return c
}
