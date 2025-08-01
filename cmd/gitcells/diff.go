package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
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
		Long: `Compare two Excel files or show changes in a file.

Examples:
  gitcells diff file1.xlsx file2.xlsx    # Compare two Excel files
  gitcells diff file.xlsx                # Compare with stored version`,
		Args: cobra.RangeArgs(minDiffArgs, maxDiffArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiff(cmd, args, logger)
		},
	}

	// Add flags
	cmd.Flags().Bool("summary", false, "Show only summary of changes")
	cmd.Flags().Bool("no-color", false, "Disable colored output")
	cmd.Flags().Bool("tui", false, "Launch interactive TUI diff viewer")
	cmd.Flags().String("format", "text", "Output format: text, json")
	cmd.Flags().StringSlice("sheets", []string{}, "Only compare specific sheets")
	cmd.Flags().Bool("ignore-formatting", false, "Ignore cell formatting differences")
	cmd.Flags().Bool("ignore-empty", false, "Ignore empty cell differences")

	return cmd
}

func runDiff(cmd *cobra.Command, args []string, logger *logrus.Logger) error {
	// Get flags
	summaryOnly, _ := cmd.Flags().GetBool("summary")
	noColor, _ := cmd.Flags().GetBool("no-color")
	tuiMode, _ := cmd.Flags().GetBool("tui")
	format, _ := cmd.Flags().GetString("format")
	sheets, _ := cmd.Flags().GetStringSlice("sheets")
	ignoreFormatting, _ := cmd.Flags().GetBool("ignore-formatting")
	ignoreEmpty, _ := cmd.Flags().GetBool("ignore-empty")

	var file1, file2 string
	file1 = args[0]

	// Determine second file
	if len(args) == maxDiffArgs {
		file2 = args[1]
	} else {
		// For single file diff, we would need to load from chunks
		// This is not yet implemented
		return utils.NewError(utils.ErrorTypeValidation, "diff", "comparing with stored version not yet implemented - please specify two Excel files")
	}

	logger.Debugf("Comparing %s with %s", file1, file2)

	// Load documents
	doc1, err := loadDocument(file1, false, ignoreFormatting, logger)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_document", file1, "failed to load first document")
	}

	doc2, err := loadDocument(file2, false, ignoreFormatting, logger)
	if err != nil {
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "load_document", file2, "failed to load second document")
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

	// Output results
	if tuiMode {
		return runDiffTUI(diff)
	}

	switch format {
	case "json":
		return outputDiffJSON(diff)
	default:
		return outputDiffText(diff, summaryOnly, !noColor)
	}
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

// diffTUIModel wraps the DiffViewer for the TUI application
type diffTUIModel struct {
	viewer *components.DiffViewer
	width  int
	height int
}

func (m diffTUIModel) Init() tea.Cmd {
	return nil
}

func (m *diffTUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewer.SetDimensions(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "tab":
			m.viewer.NextMode()
			return m, nil
		case "d":
			m.viewer.ToggleDetails()
			return m, nil
		case "up", "k":
			m.viewer.ScrollUp()
			return m, nil
		case "down", "j":
			m.viewer.ScrollDown()
			return m, nil
		case "left", "h":
			m.viewer.SelectPrev()
			return m, nil
		case "right", "l":
			m.viewer.SelectNext()
			return m, nil
		}
	}
	return m, nil
}

func (m *diffTUIModel) View() string {
	if m.viewer == nil {
		return "Loading diff viewer..."
	}
	return m.viewer.View()
}

// runDiffTUI launches the TUI diff viewer
func runDiffTUI(diff *models.ExcelDiff) error {
	viewer := components.NewDiffViewer(diff)

	model := &diffTUIModel{
		viewer: &viewer,
		width:  80,
		height: 24,
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
