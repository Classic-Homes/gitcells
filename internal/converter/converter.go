package converter

import (
	"github.com/Classic-Homes/sheetsync/pkg/models"
	"github.com/sirupsen/logrus"
)

type Converter interface {
	ExcelToJSON(filePath string, options ConvertOptions) (*models.ExcelDocument, error)
	JSONToExcel(doc *models.ExcelDocument, outputPath string, options ConvertOptions) error
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
}

type converter struct {
	logger *logrus.Logger
}

func NewConverter(logger *logrus.Logger) Converter {
	return &converter{logger: logger}
}
