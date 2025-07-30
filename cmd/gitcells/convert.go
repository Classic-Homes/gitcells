package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
			isExcelToJSON := ext == ".xlsx" || ext == ".xls" || ext == ".xlsm"

			if outputFile == "" {
				// Auto-generate output filename
				if isExcelToJSON {
					outputFile = inputFile + ".json"
				} else if ext == ".json" {
					outputFile = strings.TrimSuffix(inputFile, ".json")
					if !strings.HasSuffix(outputFile, ".xlsx") {
						outputFile += ".xlsx"
					}
				} else {
					return fmt.Errorf("unsupported file type: %s", ext)
				}
			}

			if isExcelToJSON {
				logger.Infof("Converting Excel to JSON: %s -> %s", inputFile, outputFile)
				// TODO: Implement Excel to JSON conversion
			} else {
				logger.Infof("Converting JSON to Excel: %s -> %s", inputFile, outputFile)
				// TODO: Implement JSON to Excel conversion
			}

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
