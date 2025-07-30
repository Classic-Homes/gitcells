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
	PreserveFormulas   bool
	PreserveStyles     bool
	PreserveComments   bool
	CompactJSON        bool
	IgnoreEmptyCells   bool
	MaxCellsPerSheet   int // Prevent memory issues with huge files
}

type converter struct {
	logger *logrus.Logger
}

func NewConverter(logger *logrus.Logger) Converter {
	return &converter{logger: logger}
}