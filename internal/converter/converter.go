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
}

type ConvertOptions struct {
	PreserveFormulas    bool
	PreserveStyles      bool
	PreserveComments    bool
	PreserveCharts      bool // New: Extract chart information
	PreservePivotTables bool // New: Extract pivot table structure
	CompactJSON         bool
	IgnoreEmptyCells    bool
	MaxCellsPerSheet    int                                    // Prevent memory issues with huge files
	ProgressCallback    func(stage string, current, total int) // Progress reporting
	ShowProgressBar     bool                                   // Enable built-in progress display
	ChunkingStrategy    string                                 // "sheet-based" or "hybrid" (future), defaults to "sheet-based"
}

type converter struct {
	logger           *logrus.Logger
	chunkingStrategy ChunkingStrategy
}

func NewConverter(logger *logrus.Logger) Converter {
	return &converter{
		logger:           logger,
		chunkingStrategy: NewSheetBasedChunking(logger),
	}
}
