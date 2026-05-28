package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	// minDiffArgs is the minimum number of arguments for diff command
	minDiffArgs = 1
	// maxDiffArgs is the maximum number of arguments for diff command
	maxDiffArgs = 2
)

func newDiffCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <file1> [file2]",
		Short: "Show differences between Excel files or versions",
		Long: `Compare two Excel files, or compare one Excel file against its
stored JSON chunks (the snapshot from the last 'gitcells convert' or
'gitcells sync').

Examples:
  gitcells diff file1.xlsx file2.xlsx    # Compare two Excel files
  gitcells diff file.xlsx                # Compare with stored chunks`,
		Args: cobra.RangeArgs(minDiffArgs, maxDiffArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd, args, logger)
		},
	}

	// Add flags. --no-color is inherited from the root persistent flag;
	// don't redefine it here or cobra will refuse to register both.
	cmd.Flags().Bool("summary", false, "Show only summary of changes")
	cmd.Flags().String("format", "text", "Output format: text, json")
	cmd.Flags().StringSlice("sheets", []string{}, "limit comparison to these sheets (repeatable or comma-separated)")
	cmd.Flags().Bool("ignore-formatting", false, "Ignore cell formatting differences")
	cmd.Flags().Bool("ignore-empty", false, "Ignore empty cell differences")

	return cmd
}

func runDiff(cmd *cobra.Command, args []string, logger *logrus.Logger) error {
	// Get flags
	summaryOnly, _ := cmd.Flags().GetBool("summary")
	format, _ := cmd.Flags().GetString("format")
	sheets, _ := cmd.Flags().GetStringSlice("sheets")
	ignoreFormatting, _ := cmd.Flags().GetBool("ignore-formatting")
	ignoreEmpty, _ := cmd.Flags().GetBool("ignore-empty")

	// PersistentPreRun already computed colorsDisabled from the root
	// --no-color flag, NO_COLOR env, and TTY detection. Reuse it so diff
	// output stays in sync with logger output.
	useColor := !colorsDisabled

	file1 := args[0]

	var doc1, doc2 *models.ExcelDocument
	var err error

	if len(args) == maxDiffArgs {
		// Two-arg form: compare two Excel files directly.
		file2 := args[1]
		logger.Debugf("Comparing %s with %s", file1, file2)

		doc1, err = loadDocument(file1, false, ignoreFormatting, logger)
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_document", file1, "failed to load first document")
		}
		doc2, err = loadDocument(file2, false, ignoreFormatting, logger)
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_document", file2, "failed to load second document")
		}
	} else {
		// One-arg form: compare the live xlsx against its stored chunk
		// representation under `.gitcells/data/...`. This is what
		// `gitcells diff sample.xlsx` advertises in --help and what users
		// reach for when they want to know "what changed since the last
		// commit". Pre-fix this path returned "not yet implemented".
		logger.Debugf("Comparing %s against stored chunks", file1)

		// Confirm chunks exist before doing any conversion work so the
		// error message reflects user intent ("no stored version") rather
		// than ReadChunks' generic "JSON file chunks not found".
		if err := requireStoredChunks(file1); err != nil {
			return err
		}

		doc1, err = loadDocument(file1, false, ignoreFormatting, logger)
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_document", file1, "failed to load live document")
		}

		conv := converter.NewConverter(logger)
		doc2, err = conv.ReadStoredDocument(file1)
		if err != nil {
			return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_stored_document", file1, "failed to load stored document")
		}
	}

	// Filter sheets if specified
	if len(sheets) > 0 {
		doc1 = filterSheets(doc1, sheets)
		doc2 = filterSheets(doc2, sheets)
	}

	// Compute diff
	diff := models.ComputeDiff(doc1, doc2)

	// Filter empty cell changes if requested
	if ignoreEmpty {
		diff = filterEmptyChanges(diff)
	}

	switch format {
	case "json":
		return outputDiffJSON(diff)
	default:
		return outputDiffText(diff, summaryOnly, useColor)
	}
}

// requireStoredChunks returns a clean user-facing error when no chunk
// tree exists for excelPath under `.gitcells/data/...`. Mirrors the
// resolution logic in status.go and chunking.go so the diagnostic
// matches what convert/sync would actually write.
func requireStoredChunks(excelPath string) error {
	absExcel, err := filepath.Abs(excelPath)
	if err != nil {
		absExcel = excelPath
	}
	excelDir := filepath.Dir(absExcel)
	gitRoot, gerr := git.FindRepositoryRoot(excelDir)
	if gerr != nil {
		gitRoot = excelDir
	}
	relPath, rerr := filepath.Rel(gitRoot, absExcel)
	if rerr != nil {
		relPath = absExcel
	}
	baseName := strings.TrimSuffix(filepath.Base(absExcel), filepath.Ext(absExcel))
	chunkDir := filepath.Join(gitRoot, constants.GitCellsDataDir, filepath.Dir(relPath), baseName+constants.ChunksDirSuffix)
	metaFile := filepath.Join(chunkDir, constants.ChunkMetadataFile)
	if _, err := os.Stat(metaFile); err != nil {
		if os.IsNotExist(err) {
			return utils.NewError(utils.ErrorTypeValidation, "diff",
				fmt.Sprintf("no stored version found for %s; run 'gitcells convert' or 'gitcells sync' first", excelPath))
		}
		return utils.WrapFileError(err, utils.ErrorTypeFileSystem, "diff", metaFile, "failed to stat chunk metadata")
	}
	return nil
}

func loadDocument(filePath string, jsonMode, ignoreFormatting bool, logger *logrus.Logger) (*models.ExcelDocument, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Only support Excel files
	if ext != ".xlsx" && ext != ".xls" && ext != ".xlsm" {
		return nil, utils.NewError(utils.ErrorTypeValidation, "loadDocument", fmt.Sprintf("unsupported file type: %s", ext))
	}

	// Load Excel file and convert
	conv := converter.NewConverter(logger)
	options := converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   !ignoreFormatting,
		PreserveComments: true,
		IgnoreEmptyCells: false,
	}

	return conv.ExcelToJSON(filePath, options)
}

func filterSheets(doc *models.ExcelDocument, sheetNames []string) *models.ExcelDocument {
	sheetMap := make(map[string]bool)
	for _, name := range sheetNames {
		sheetMap[name] = true
	}

	filtered := &models.ExcelDocument{
		Version:      doc.Version,
		Metadata:     doc.Metadata,
		Sheets:       []models.Sheet{},
		DefinedNames: doc.DefinedNames,
		Properties:   doc.Properties,
	}

	for _, sheet := range doc.Sheets {
		if sheetMap[sheet.Name] {
			filtered.Sheets = append(filtered.Sheets, sheet)
		}
	}

	return filtered
}

func filterEmptyChanges(diff *models.ExcelDiff) *models.ExcelDiff {
	filtered := &models.ExcelDiff{
		Timestamp:  diff.Timestamp,
		Summary:    diff.Summary,
		SheetDiffs: []models.SheetDiff{},
	}

	for _, sheetDiff := range diff.SheetDiffs {
		filteredSheet := models.SheetDiff{
			SheetName: sheetDiff.SheetName,
			Action:    sheetDiff.Action,
			Changes:   []models.CellChange{},
		}

		for _, change := range sheetDiff.Changes {
			// Skip changes where both old and new values are empty
			if isEmptyValue(change.OldValue) && isEmptyValue(change.NewValue) {
				continue
			}
			filteredSheet.Changes = append(filteredSheet.Changes, change)
		}

		if len(filteredSheet.Changes) > 0 || filteredSheet.Action != "" {
			filtered.SheetDiffs = append(filtered.SheetDiffs, filteredSheet)
		}
	}

	// Recalculate summary
	filtered.Summary.CellChanges = 0
	for _, sheetDiff := range filtered.SheetDiffs {
		filtered.Summary.CellChanges += len(sheetDiff.Changes)
	}

	return filtered
}

func isEmptyValue(value interface{}) bool {
	if value == nil {
		return true
	}

	if str, ok := value.(string); ok {
		return strings.TrimSpace(str) == ""
	}

	return false
}

func outputDiffJSON(diff *models.ExcelDiff) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(diff)
}

func outputDiffText(diff *models.ExcelDiff, summaryOnly, useColor bool) error {
	if !diff.HasChanges() {
		fmt.Println("No differences found")
		return nil
	}

	// Color codes
	var (
		green  = ""
		red    = ""
		yellow = ""
		blue   = ""
		reset  = ""
	)

	if useColor {
		green = "\033[32m"
		red = "\033[31m"
		yellow = "\033[33m"
		blue = "\033[34m"
		reset = "\033[0m"
	}

	// Print summary
	fmt.Printf("%s=== Diff Summary ===%s\n", blue, reset)
	fmt.Printf("Timestamp: %s\n", diff.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("Changes: %s\n", diff.String())
	fmt.Println()

	if summaryOnly {
		return nil
	}

	// Print detailed changes
	for _, sheetDiff := range diff.SheetDiffs {
		fmt.Printf("%s=== Sheet: %s ===%s\n", blue, sheetDiff.SheetName, reset)

		if sheetDiff.Action != "" {
			actionColor := green
			if sheetDiff.Action == models.ChangeTypeDelete {
				actionColor = red
			}
			fmt.Printf("Action: %s%s%s\n", actionColor, strings.ToUpper(string(sheetDiff.Action)), reset)
		}

		if len(sheetDiff.Changes) == 0 {
			fmt.Println("No cell changes")
		} else {
			fmt.Printf("Cell changes (%d):\n", len(sheetDiff.Changes))

			for _, change := range sheetDiff.Changes {
				var changeColor string
				var symbol string

				switch change.Type {
				case models.ChangeTypeAdd:
					changeColor = green
					symbol = "+"
				case models.ChangeTypeDelete:
					changeColor = red
					symbol = "-"
				case models.ChangeTypeModify:
					changeColor = yellow
					symbol = "~"
				}

				fmt.Printf("  %s%s %s%s", changeColor, symbol, change.Cell, reset)

				if change.Description != "" {
					fmt.Printf(": %s", change.Description)
				}

				// Show value changes
				switch change.Type {
				case models.ChangeTypeModify:
					if change.OldValue != nil || change.NewValue != nil {
						fmt.Printf(" (%s%v%s → %s%v%s)",
							red, change.OldValue, reset,
							green, change.NewValue, reset)
					}
				case models.ChangeTypeAdd:
					if change.NewValue != nil {
						fmt.Printf(" (%s%v%s)", green, change.NewValue, reset)
					}
				case models.ChangeTypeDelete:
					if change.OldValue != nil {
						fmt.Printf(" (%s%v%s)", red, change.OldValue, reset)
					}
				}

				fmt.Println()
			}
		}
		fmt.Println()
	}

	return nil
}
