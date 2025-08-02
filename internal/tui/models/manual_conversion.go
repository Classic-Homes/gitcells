package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConversionState int

const (
	ConversionStateDirectionSelection ConversionState = iota
	ConversionStateFileSelection
	ConversionStateModeSelection
	ConversionStateSheetSelection
	ConversionStateOptions
	ConversionStateProcessing
	ConversionStateResults
	ConversionStateError
)

type ConversionDirection int

const (
	ConversionDirectionExcelToJSON ConversionDirection = iota
	ConversionDirectionJSONToExcel
)

type ConversionMode int

const (
	ConversionModeAll ConversionMode = iota
	ConversionModeSelectSheets
)

type ManualConversionModel struct {
	state               ConversionState
	width               int
	height              int
	cursor              int
	files               []string
	jsonFiles           []string
	jsonFilePaths       map[string]string // Maps display name to actual path
	selectedFile        string
	conversionDirection ConversionDirection
	conversionMode      ConversionMode
	modeCursor          int
	directionCursor     int
	availableSheets     []adapter.SheetInfo
	selectedSheets      map[string]bool
	sheetCursor         int
	options             ConversionOptions
	optionCursor        int
	result              *adapter.ConversionResult
	errorMsg            string
	showHelp            bool
	converterAdapter    *adapter.ConverterAdapter
}

type ConversionOptions struct {
	PreserveFormulas    bool
	PreserveStyles      bool
	PreserveComments    bool
	PreserveCharts      bool
	PreservePivotTables bool
	CompactJSON         bool
	IgnoreEmptyCells    bool
}

func NewManualConversionModel() ManualConversionModel {
	return ManualConversionModel{
		state: ConversionStateDirectionSelection,
		options: ConversionOptions{
			PreserveFormulas:    true,
			PreserveStyles:      true,
			PreserveComments:    true,
			PreserveCharts:      false,
			PreservePivotTables: false,
			CompactJSON:         false,
			IgnoreEmptyCells:    true,
		},
		selectedSheets:      make(map[string]bool),
		jsonFilePaths:       make(map[string]string),
		showHelp:            true,
		converterAdapter:    adapter.NewConverterAdapter(),
		conversionDirection: ConversionDirectionExcelToJSON,
	}
}

func (m ManualConversionModel) Init() tea.Cmd {
	return m.loadFiles
}

func (m ManualConversionModel) loadFiles() tea.Msg {
	var excelFiles []string
	var jsonFiles []string
	jsonFilePaths := make(map[string]string)

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return errMsg{err}
	}

	// Walk through current directory and subdirectories looking for Excel files
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			// Skip .git and other hidden directories (except .gitcells)
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." && info.Name() != ".gitcells" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		// Make path relative to current directory
		relPath, err := filepath.Rel(cwd, path)
		if err != nil {
			relPath = path
		}

		if ext == ".xlsx" || ext == ".xls" || ext == ".xlsm" {
			// Skip temporary Excel files
			if !strings.HasPrefix(info.Name(), "~$") {
				excelFiles = append(excelFiles, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return errMsg{err}
	}

	// Look for GitCells JSON files in .gitcells/data directory
	gitCellsDataDir := filepath.Join(cwd, ".gitcells", "data")
	if _, err := os.Stat(gitCellsDataDir); err == nil {
		// Walk through .gitcells/data looking for chunk directories
		err = filepath.Walk(gitCellsDataDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}

			// Look for directories ending with _chunks
			if info.IsDir() && strings.HasSuffix(info.Name(), "_chunks") {
				// Check if this is a valid GitCells chunk directory by looking for metadata
				metadataFile := filepath.Join(path, ".gitcells_chunks.json")
				if _, err := os.Stat(metadataFile); err == nil {
					// Extract the original filename from the chunk directory name
					dirName := info.Name()
					originalName := strings.TrimSuffix(dirName, "_chunks")

					// Store the actual chunk directory path for conversion
					jsonFilePaths[originalName] = path

					// Add to JSON files list
					jsonFiles = append(jsonFiles, originalName)
				}
			}
			return nil
		})

		if err != nil {
			return errMsg{err}
		}
	}

	return conversionFilesLoadedMsg{excelFiles: excelFiles, jsonFiles: jsonFiles, jsonFilePaths: jsonFilePaths}
}

func (m ManualConversionModel) loadSheetsForFile(filename string) tea.Cmd {
	return func() tea.Msg {
		sheets, err := m.converterAdapter.GetExcelSheets(filename)
		if err != nil {
			return errMsg{fmt.Errorf("failed to load sheets from %s: %w", filename, err)}
		}
		return sheetsLoadedMsg{sheets}
	}
}

type conversionFilesLoadedMsg struct {
	excelFiles    []string
	jsonFiles     []string
	jsonFilePaths map[string]string
}

type sheetsLoadedMsg struct {
	sheets []adapter.SheetInfo
}

func (m ManualConversionModel) performConversion() tea.Cmd {
	return func() tea.Msg {
		var result *adapter.ConversionResult
		var err error

		if m.conversionDirection == ConversionDirectionExcelToJSON {
			// Excel to JSON conversion
			if m.conversionMode == ConversionModeAll {
				// Use standard conversion without sheet selection
				result, err = m.converterAdapter.ConvertFile(m.selectedFile)
			} else {
				// Convert selected sheets to slice
				var sheetsToConvert []string
				for sheetName, selected := range m.selectedSheets {
					if selected {
						sheetsToConvert = append(sheetsToConvert, sheetName)
					}
				}

				// Prepare sheet selection options
				sheetOptions := adapter.SheetSelectionOptions{
					SheetsToConvert: sheetsToConvert,
				}

				// If no sheets selected, convert all
				if len(sheetsToConvert) == 0 {
					sheetOptions = adapter.SheetSelectionOptions{}
				}

				result, err = m.converterAdapter.ConvertFileWithSheetOptions(m.selectedFile, sheetOptions)
			}
		} else {
			// JSON to Excel conversion
			result, err = m.converterAdapter.ConvertJSONToExcel(m.selectedFile)
		}

		if err != nil {
			return errMsg{fmt.Errorf("conversion failed: %w", err)}
		}

		return conversionCompleteMsg{result}
	}
}

type conversionCompleteMsg struct {
	result *adapter.ConversionResult
}

func (m *ManualConversionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case conversionFilesLoadedMsg:
		m.files = msg.excelFiles
		m.jsonFiles = msg.jsonFiles
		m.jsonFilePaths = msg.jsonFilePaths
		if len(m.files) == 0 && len(m.jsonFiles) == 0 {
			m.state = ConversionStateError
			m.errorMsg = "No Excel or JSON files found in current directory"
		}
		return m, nil

	case sheetsLoadedMsg:
		m.availableSheets = msg.sheets
		m.selectedSheets = make(map[string]bool)
		m.state = ConversionStateSheetSelection
		return m, nil

	case conversionCompleteMsg:
		m.result = msg.result
		m.state = ConversionStateResults
		return m, nil

	case errMsg:
		m.state = ConversionStateError
		m.errorMsg = msg.err.Error()
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m *ManualConversionModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ConversionStateDirectionSelection:
		return m.handleDirectionSelectionKeys(msg)
	case ConversionStateFileSelection:
		return m.handleFileSelectionKeys(msg)
	case ConversionStateModeSelection:
		return m.handleModeSelectionKeys(msg)
	case ConversionStateSheetSelection:
		return m.handleSheetSelectionKeys(msg)
	case ConversionStateOptions:
		return m.handleOptionsKeys(msg)
	case ConversionStateProcessing:
		return m.handleProcessingKeys(msg)
	case ConversionStateResults:
		return m.handleResultsKeys(msg)
	case ConversionStateError:
		return m.handleErrorKeys(msg)
	}
	return m, nil
}

func (m *ManualConversionModel) handleDirectionSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.resetState()
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "up", "k":
		if m.directionCursor > 0 {
			m.directionCursor--
		}
	case "down", "j":
		if m.directionCursor < 1 {
			m.directionCursor++
		}
	case "enter", " ":
		m.conversionDirection = ConversionDirection(m.directionCursor)
		m.state = ConversionStateFileSelection
		return m, nil
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *ManualConversionModel) handleFileSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Get appropriate file list based on conversion direction
	var fileList []string
	if m.conversionDirection == ConversionDirectionExcelToJSON {
		fileList = m.files
	} else {
		fileList = m.jsonFiles
	}

	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.state = ConversionStateDirectionSelection
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(fileList)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(fileList) > 0 && m.cursor < len(fileList) {
			selectedName := fileList[m.cursor]

			// For JSON files, get the actual path from the map
			if m.conversionDirection == ConversionDirectionJSONToExcel {
				if actualPath, ok := m.jsonFilePaths[selectedName]; ok {
					m.selectedFile = actualPath
				} else {
					m.selectedFile = selectedName // Fallback
				}
				m.state = ConversionStateOptions
				m.optionCursor = 0
			} else {
				m.selectedFile = selectedName
				m.state = ConversionStateModeSelection
				m.modeCursor = 0
			}
			return m, nil
		}
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *ManualConversionModel) handleModeSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.state = ConversionStateFileSelection
		m.selectedFile = ""
		m.cursor = 0
		return m, nil
	case "up", "k":
		if m.modeCursor > 0 {
			m.modeCursor--
		}
	case "down", "j":
		if m.modeCursor < 1 {
			m.modeCursor++
		}
	case "enter", " ":
		m.conversionMode = ConversionMode(m.modeCursor)
		if m.conversionMode == ConversionModeAll {
			// Skip sheet selection and go directly to options
			m.state = ConversionStateOptions
			m.optionCursor = 0
		} else {
			// Load sheets for selection
			return m, m.loadSheetsForFile(m.selectedFile)
		}
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *ManualConversionModel) handleSheetSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.state = ConversionStateModeSelection
		m.availableSheets = nil
		m.selectedSheets = make(map[string]bool)
		return m, nil
	case "up", "k":
		if m.sheetCursor > 0 {
			m.sheetCursor--
		}
	case "down", "j":
		if m.sheetCursor < len(m.availableSheets)-1 {
			m.sheetCursor++
		}
	case " ":
		if len(m.availableSheets) > 0 && m.sheetCursor < len(m.availableSheets) {
			sheetName := m.availableSheets[m.sheetCursor].Name
			m.selectedSheets[sheetName] = !m.selectedSheets[sheetName]
		}
	case "a":
		// Select all sheets
		for _, sheet := range m.availableSheets {
			m.selectedSheets[sheet.Name] = true
		}
	case "n":
		// Select none
		m.selectedSheets = make(map[string]bool)
	case "enter":
		// Proceed to options after sheet selection
		m.state = ConversionStateOptions
		m.optionCursor = 0
	case "c":
		// Convert with current selection
		m.state = ConversionStateProcessing
		return m, m.performConversion()
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *ManualConversionModel) handleOptionsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		// Navigate back based on conversion direction and mode
		switch {
		case m.conversionDirection == ConversionDirectionJSONToExcel:
			m.state = ConversionStateFileSelection
		case m.conversionMode == ConversionModeAll:
			m.state = ConversionStateModeSelection
		default:
			m.state = ConversionStateSheetSelection
		}
		return m, nil
	case "up", "k":
		if m.optionCursor > 0 {
			m.optionCursor--
		}
	case "down", "j":
		if m.optionCursor < 6 {
			m.optionCursor++
		}
	case "enter", " ":
		m.toggleOption(m.optionCursor)
	case "tab":
		// Tab is not used for navigation in options - removed to avoid confusion
	case "c":
		// Convert with current options
		m.state = ConversionStateProcessing
		return m, m.performConversion()
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m *ManualConversionModel) toggleOption(index int) {
	switch index {
	case 0:
		m.options.PreserveFormulas = !m.options.PreserveFormulas
	case 1:
		m.options.PreserveStyles = !m.options.PreserveStyles
	case 2:
		m.options.PreserveComments = !m.options.PreserveComments
	case 3:
		m.options.PreserveCharts = !m.options.PreserveCharts
	case 4:
		m.options.PreservePivotTables = !m.options.PreservePivotTables
	case 5:
		m.options.CompactJSON = !m.options.CompactJSON
	case 6:
		m.options.IgnoreEmptyCells = !m.options.IgnoreEmptyCells
	}
}

func (m *ManualConversionModel) handleProcessingKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Only allow quit during processing
	if msg.String() == "ctrl+c" {
		m.resetState()
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	}
	return m, nil
}

func (m *ManualConversionModel) handleResultsKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.resetState()
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "r":
		// Reset for another conversion
		m.resetState()
		return m, m.loadFiles
	}
	return m, nil
}

func (m *ManualConversionModel) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.resetState()
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "r":
		// Retry - reset and reload
		m.resetState()
		return m, m.loadFiles
	}
	return m, nil
}

func (m ManualConversionModel) View() string {
	switch m.state {
	case ConversionStateDirectionSelection:
		return m.renderDirectionSelection()
	case ConversionStateFileSelection:
		return m.renderFileSelection()
	case ConversionStateModeSelection:
		return m.renderModeSelection()
	case ConversionStateSheetSelection:
		return m.renderSheetSelection()
	case ConversionStateOptions:
		return m.renderOptions()
	case ConversionStateProcessing:
		return m.renderProcessing()
	case ConversionStateResults:
		return m.renderResults()
	case ConversionStateError:
		return m.renderError()
	}
	return "Loading..."
}

func (m ManualConversionModel) renderDirectionSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)
	title := titleStyle.Render("Manual File Conversion")

	instructions := styles.MutedStyle.Render("Select conversion direction:")

	directions := []struct {
		name string
		desc string
	}{
		{"Excel → JSON", "Convert Excel files to JSON format"},
		{"JSON → Excel", "Convert JSON files back to Excel format"},
	}

	directionList := make([]string, 0, len(directions))
	for i, direction := range directions {
		cursor := "  "
		directionStyle := lipgloss.NewStyle()

		if i == m.directionCursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			directionStyle = directionStyle.Bold(true)
		}

		line := cursor + directionStyle.Render(direction.name)
		desc := "    " + styles.MutedStyle.Render(direction.desc)
		directionList = append(directionList, line)
		directionList = append(directionList, desc)
		if i < len(directions)-1 {
			directionList = append(directionList, "")
		}
	}

	directionContent := strings.Join(directionList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Enter/Space] Select • [h/?] Toggle help • [q] Back to Tools",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [q] Back to Tools")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		instructions,
		"",
		directionContent,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m ManualConversionModel) renderFileSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)
	var title, instructions, noFilesError string
	var fileList []string

	if m.conversionDirection == ConversionDirectionExcelToJSON {
		title = titleStyle.Render("Excel → JSON Conversion")
		instructions = styles.MutedStyle.Render("Select Excel file to convert:")
		noFilesError = "No Excel files found in current directory"
		fileList = m.files
	} else {
		title = titleStyle.Render("JSON → Excel Conversion")
		instructions = styles.MutedStyle.Render("Select GitCells JSON file to convert back to Excel:")
		noFilesError = "No GitCells JSON files found in .gitcells/data directory"
		fileList = m.jsonFiles
	}

	if len(fileList) == 0 {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			styles.ErrorStyle.Render(noFilesError),
			"",
			styles.HelpStyle.Render("[q] Back to Direction Selection • [r] Retry"),
		)
		return styles.Center(m.width, m.height, content)
	}

	// File list
	renderedFileList := make([]string, 0, len(fileList))
	for i, file := range fileList {
		cursor := "  "
		fileStyle := lipgloss.NewStyle()

		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			fileStyle = fileStyle.Bold(true)
		}

		displayName := file
		if m.conversionDirection == ConversionDirectionJSONToExcel {
			// For JSON files, show the original Excel filename
			displayName = file + ".xlsx"
		}

		renderedFileList = append(renderedFileList, cursor+fileStyle.Render(displayName))
	}

	files := strings.Join(renderedFileList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Enter/Space] Select • [h/?] Toggle help • [q] Back to Direction Selection",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [q] Back to Direction Selection")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		instructions,
		"",
		files,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m ManualConversionModel) renderModeSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)
	title := titleStyle.Render(fmt.Sprintf("Conversion Mode - %s", filepath.Base(m.selectedFile)))

	instructions := styles.MutedStyle.Render("Select conversion mode:")

	modes := []struct {
		name string
		desc string
	}{
		{"Convert All Sheets", "Quick conversion of entire Excel file"},
		{"Select Specific Sheets", "Choose which sheets to convert"},
	}

	modeList := make([]string, 0, len(modes))
	for i, mode := range modes {
		cursor := "  "
		modeStyle := lipgloss.NewStyle()

		if i == m.modeCursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			modeStyle = modeStyle.Bold(true)
		}

		line := cursor + modeStyle.Render(mode.name)
		desc := "    " + styles.MutedStyle.Render(mode.desc)
		modeList = append(modeList, line)
		modeList = append(modeList, desc)
		if i < len(modes)-1 {
			modeList = append(modeList, "")
		}
	}

	modeContent := strings.Join(modeList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Enter/Space] Select • [h/?] Toggle help • [q] Back",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [q] Back")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		instructions,
		"",
		modeContent,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m ManualConversionModel) renderSheetSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)
	title := titleStyle.Render(fmt.Sprintf("Sheet Selection - %s", filepath.Base(m.selectedFile)))

	instructions := styles.MutedStyle.Render("Select sheets to convert (empty selection = all sheets):")

	// Sheet list
	sheetList := make([]string, 0, len(m.availableSheets))
	for i, sheet := range m.availableSheets {
		cursor := "  "
		sheetStyle := lipgloss.NewStyle()
		checkbox := "☐"

		if i == m.sheetCursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			sheetStyle = sheetStyle.Bold(true)
		}

		if m.selectedSheets[sheet.Name] {
			checkbox = lipgloss.NewStyle().Foreground(styles.Success).Render("☑")
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, sheet.Name)
		sheetList = append(sheetList, sheetStyle.Render(line))
	}

	sheets := strings.Join(sheetList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Space] Toggle • [a] All • [n] None • [Enter] Options • [c] Convert • [q] Back",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [c] Convert • [q] Back")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		instructions,
		"",
		sheets,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m ManualConversionModel) renderOptions() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)
	title := titleStyle.Render("Conversion Options")

	instructions := styles.MutedStyle.Render("Configure conversion settings:")

	options := []struct {
		name  string
		value bool
		desc  string
	}{
		{"Preserve Formulas", m.options.PreserveFormulas, "Keep Excel formulas"},
		{"Preserve Styles", m.options.PreserveStyles, "Keep cell formatting"},
		{"Preserve Comments", m.options.PreserveComments, "Keep cell comments"},
		{"Preserve Charts", m.options.PreserveCharts, "Extract chart information"},
		{"Preserve Pivot Tables", m.options.PreservePivotTables, "Extract pivot table structure"},
		{"Compact JSON", m.options.CompactJSON, "Generate compressed JSON"},
		{"Ignore Empty Cells", m.options.IgnoreEmptyCells, "Skip empty cells"},
	}

	optionList := make([]string, 0, len(options))
	for i, option := range options {
		cursor := "  "
		optionStyle := lipgloss.NewStyle()
		checkbox := "☐"

		if i == m.optionCursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			optionStyle = optionStyle.Bold(true)
		}

		if option.value {
			checkbox = lipgloss.NewStyle().Foreground(styles.Success).Render("☑")
		}

		line := fmt.Sprintf("%s %s %s", cursor, checkbox, option.name)
		desc := "    " + styles.MutedStyle.Render(option.desc)
		optionList = append(optionList, optionStyle.Render(line))
		optionList = append(optionList, desc)
	}

	optionContent := strings.Join(optionList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Space] Toggle • [c] Convert • [q] Back",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [c] Convert • [q] Back")
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		instructions,
		"",
		optionContent,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)
}

func (m ManualConversionModel) renderProcessing() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render("Converting..."),
		"",
		styles.MutedStyle.Render(fmt.Sprintf("Converting %s", filepath.Base(m.selectedFile))),
		"",
		"⏳ Processing...",
		"",
		styles.HelpStyle.Render("[Ctrl+C] Cancel"),
	)
	return styles.Center(m.width, m.height, content)
}

func (m ManualConversionModel) renderResults() string {
	var statusStyle lipgloss.Style
	var statusText string

	if m.result != nil && m.result.Success {
		statusStyle = styles.SuccessStyle
		statusText = "✅ Conversion Successful!"
	} else {
		statusStyle = styles.ErrorStyle
		statusText = "❌ Conversion Failed"
	}

	details := ""
	if m.result != nil {
		modeText := "All sheets"
		if m.conversionMode == ConversionModeSelectSheets {
			sheetCount := 0
			for _, selected := range m.selectedSheets {
				if selected {
					sheetCount++
				}
			}
			if sheetCount > 0 {
				modeText = fmt.Sprintf("%d selected sheet(s)", sheetCount)
			}
		}
		details = fmt.Sprintf("Mode:   %s\nInput:  %s\nOutput: %s", modeText, m.result.ExcelPath, m.result.JSONPath)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render("Conversion Results"),
		"",
		statusStyle.Render(statusText),
		"",
		styles.MutedStyle.Render(details),
		"",
		styles.HelpStyle.Render("[r] Convert Another • [q] Back to Tools"),
	)
	return styles.Center(m.width, m.height, content)
}

func (m ManualConversionModel) renderError() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render("Error"),
		"",
		styles.ErrorStyle.Render(m.errorMsg),
		"",
		styles.HelpStyle.Render("[r] Retry • [q] Back to Tools"),
	)
	return styles.Center(m.width, m.height, content)
}

// resetState resets the model to initial state
func (m *ManualConversionModel) resetState() {
	m.state = ConversionStateDirectionSelection
	m.cursor = 0
	m.directionCursor = 0
	m.modeCursor = 0
	m.sheetCursor = 0
	m.optionCursor = 0
	m.selectedFile = ""
	m.conversionDirection = ConversionDirectionExcelToJSON
	m.conversionMode = ConversionModeAll
	m.availableSheets = nil
	m.selectedSheets = make(map[string]bool)
	m.jsonFilePaths = make(map[string]string)
	m.result = nil
	m.errorMsg = ""
}
