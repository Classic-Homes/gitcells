package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	extXLSX = ".xlsx"
	extJSON = ".json"
)

func newConvertCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "convert <file>",
		Short: "Convert between Excel and JSON formats",
		Long:  "Convert Excel files to JSON or JSON files back to Excel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputFile := args[0]
			outputFile, _ := cmd.Flags().GetString("output")

			// Determine conversion direction
			ext := strings.ToLower(filepath.Ext(inputFile))
			isExcelToJSON := ext == extXLSX || ext == ".xls" || ext == ".xlsm"

			// Check if input is a chunk directory
			isChunkDir := false
			if !isExcelToJSON {
				// Check if it's a chunk directory
				if strings.HasSuffix(inputFile, "_chunks") {
					isChunkDir = true
				} else if ext == extJSON {
					// Legacy single JSON file - not supported
					return utils.NewError(utils.ErrorTypeValidation, "convert", "single JSON files are no longer supported, please use chunk directories")
				}
			}

			if outputFile == "" {
				// Auto-generate output filename
				switch {
				case isExcelToJSON:
					// For Excel to JSON, we'll use the input path as the base for chunking
					outputFile = inputFile
				case isChunkDir:
					// Extract Excel filename from chunk directory name
					baseName := strings.TrimSuffix(filepath.Base(inputFile), "_chunks")
					outputFile = filepath.Join(filepath.Dir(inputFile), baseName)
					if !strings.HasSuffix(outputFile, extXLSX) {
						outputFile += extXLSX
					}
				default:
					return utils.NewError(utils.ErrorTypeValidation, "convert", fmt.Sprintf("unsupported input: %s", inputFile))
				}
			}

			// Create converter
			conv := converter.NewConverter(logger)

			// Build conversion options
			opts := converter.ConvertOptions{
				PreserveFormulas: getBoolFlag(cmd, "preserve-formulas"),
				PreserveStyles:   getBoolFlag(cmd, "preserve-styles"),
				PreserveComments: getBoolFlag(cmd, "preserve-comments"),
				CompactJSON:      getBoolFlag(cmd, "compact"),
				ChunkingStrategy: "sheet-based",
			}

			// Add sheet selection options for Excel to JSON conversion
			if isExcelToJSON {
				if sheetsToConvert, _ := cmd.Flags().GetStringSlice("sheets"); len(sheetsToConvert) > 0 {
					opts.SheetsToConvert = sheetsToConvert
				}
				if excludeSheets, _ := cmd.Flags().GetStringSlice("exclude-sheets"); len(excludeSheets) > 0 {
					opts.ExcludeSheets = excludeSheets
				}
				if sheetIndices, _ := cmd.Flags().GetIntSlice("sheet-indices"); len(sheetIndices) > 0 {
					opts.SheetIndices = sheetIndices
				}
			}

			if isExcelToJSON {
				logger.Infof("Converting Excel to JSON chunks in .gitcells/data: %s", inputFile)
				if err := conv.ExcelToJSONFile(inputFile, outputFile, opts); err != nil {
					return utils.WrapFileError(err, utils.ErrorTypeConverter, "convert", inputFile, "conversion failed")
				}
			} else {
				logger.Infof("Converting JSON to Excel: %s -> %s", inputFile, outputFile)
				if err := conv.JSONFileToExcel(inputFile, outputFile, opts); err != nil {
					return utils.WrapFileError(err, utils.ErrorTypeConverter, "convert", inputFile, "conversion failed")
				}
			}

			logger.Info("Conversion completed successfully")
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "output file path (auto-generated if not specified)")
	cmd.Flags().Bool("preserve-formulas", true, "preserve Excel formulas")
	cmd.Flags().Bool("preserve-styles", true, "preserve cell styles")
	cmd.Flags().Bool("preserve-comments", true, "preserve cell comments")
	cmd.Flags().Bool("compact", false, "output compact JSON")

	// Sheet selection flags (only applicable for Excel to JSON conversion)
	cmd.Flags().StringSlice("sheets", []string{}, "comma-separated list of sheet names to convert (default: all sheets)")
	cmd.Flags().StringSlice("exclude-sheets", []string{}, "comma-separated list of sheet names to exclude from conversion")
	cmd.Flags().IntSlice("sheet-indices", []int{}, "comma-separated list of sheet indices to convert (0-based, default: all sheets)")

	return cmd
}

// getBoolFlag safely retrieves a boolean flag value
func getBoolFlag(cmd *cobra.Command, name string) bool {
	val, _ := cmd.Flags().GetBool(name)
	return val
}
