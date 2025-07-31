package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/converter"
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

			// Determine conversion direction based on file extension
			ext := strings.ToLower(filepath.Ext(inputFile))
			isExcelToJSON := ext == extXLSX || ext == ".xls" || ext == ".xlsm"

			if outputFile == "" {
				// Auto-generate output filename
				if isExcelToJSON {
					outputFile = inputFile + extJSON
				} else if ext == extJSON {
					outputFile = strings.TrimSuffix(inputFile, extJSON)
					if !strings.HasSuffix(outputFile, extXLSX) {
						outputFile += extXLSX
					}
				} else {
					return fmt.Errorf("unsupported file type: %s", ext)
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

			if isExcelToJSON {
				logger.Infof("Converting Excel to JSON: %s -> %s", inputFile, outputFile)
				if err := conv.ExcelToJSONFile(inputFile, outputFile, opts); err != nil {
					return fmt.Errorf("conversion failed: %w", err)
				}
			} else {
				logger.Infof("Converting JSON to Excel: %s -> %s", inputFile, outputFile)
				if err := conv.JSONFileToExcel(inputFile, outputFile, opts); err != nil {
					return fmt.Errorf("conversion failed: %w", err)
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

	return cmd
}

// getBoolFlag safely retrieves a boolean flag value
func getBoolFlag(cmd *cobra.Command, name string) bool {
	val, _ := cmd.Flags().GetBool(name)
	return val
}
