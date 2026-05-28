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

// ConvertOptions controls Excel↔JSON conversion. Only features that
// round-trip cleanly are exposed. Charts, pivot tables, conditional
// formatting, tables and rich-text-per-run are intentionally not
// captured — the converter logs a per-sheet warning when source workbooks
// contain them so users know what wasn't preserved.
type ConvertOptions struct {
	PreserveFormulas bool
	PreserveStyles   bool
	PreserveComments bool
	CompactJSON      bool
	IgnoreEmptyCells bool
	MaxCellsPerSheet int    // Prevent memory issues with huge files
	ChunkingStrategy string // currently always "sheet-based"

	ProgressCallback func(stage string, current, total int)
	ShowProgressBar  bool

	// Sheet selection options. All zero-value = convert every sheet.
	SheetsToConvert []string // Specific sheet names to convert (empty = convert all)
	ExcludeSheets   []string // Sheet names to exclude from conversion
	SheetIndices    []int    // Specific sheet indices to convert (0-based, empty = convert all)

	// SkipBadSheets, when true, downgrades a sheet that fails to process
	// from a hard error to a warn-and-continue. Default (false) aborts the
	// whole conversion on the first failure, which is what callers want
	// when round-trip integrity matters more than partial output.
	SkipBadSheets bool
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
