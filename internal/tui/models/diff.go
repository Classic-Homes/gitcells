package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sirupsen/logrus"
)

type DiffState int

const (
	DiffStateFileSelection DiffState = iota
	DiffStateViewing
	DiffStateError
)

type DiffModel struct {
	state         DiffState
	width         int
	height        int
	cursor        int
	files         []string
	selectedFile1 string
	selectedFile2 string
	diffViewer    *components.DiffViewer
	errorMsg      string
	showHelp      bool
}

func NewDiffModel() DiffModel {
	return DiffModel{
		state:    DiffStateFileSelection,
		showHelp: true,
	}
}

func (m DiffModel) Init() tea.Cmd {
	return m.loadExcelFiles
}

func (m DiffModel) loadExcelFiles() tea.Msg {
	var files []string

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
			// Skip .git and other hidden directories
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".xlsx" || ext == ".xls" || ext == ".xlsm" {
			// Make path relative to current directory
			relPath, err := filepath.Rel(cwd, path)
			if err != nil {
				relPath = path
			}
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return errMsg{err}
	}

	return filesLoadedMsg{files}
}

type filesLoadedMsg struct {
	files []string
}

type errMsg struct {
	err error
}

func (m DiffModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.diffViewer != nil {
			m.diffViewer.SetDimensions(msg.Width, msg.Height)
		}
		return m, nil

	case filesLoadedMsg:
		m.files = msg.files
		if len(m.files) == 0 {
			m.state = DiffStateError
			m.errorMsg = "No Excel files found in current directory"
		}
		return m, nil

	case errMsg:
		m.state = DiffStateError
		m.errorMsg = msg.err.Error()
		return m, nil

	case diffComputedMsg:
		return m.handleDiffComputed(msg.diff)

	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}

	return m, nil
}

func (m DiffModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case DiffStateFileSelection:
		return m.handleFileSelectionKeys(msg)
	case DiffStateViewing:
		return m.handleDiffViewerKeys(msg)
	case DiffStateError:
		return m.handleErrorKeys(msg)
	}
	return m, nil
}

func (m DiffModel) handleFileSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.files)-1 {
			m.cursor++
		}
	case "enter", " ":
		if len(m.files) > 0 && m.cursor < len(m.files) {
			if m.selectedFile1 == "" {
				m.selectedFile1 = m.files[m.cursor]
			} else if m.selectedFile2 == "" && m.files[m.cursor] != m.selectedFile1 {
				m.selectedFile2 = m.files[m.cursor]
				return m, m.loadAndCompareDiff
			}
		}
	case "r":
		// Reset selection
		m.selectedFile1 = ""
		m.selectedFile2 = ""
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m DiffModel) handleDiffViewerKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		// Go back to file selection
		m.state = DiffStateFileSelection
		m.selectedFile1 = ""
		m.selectedFile2 = ""
		m.diffViewer = nil
		return m, nil
	case "tab":
		if m.diffViewer != nil {
			m.diffViewer.NextMode()
		}
	case "d":
		if m.diffViewer != nil {
			m.diffViewer.ToggleDetails()
		}
	case "up", "k":
		if m.diffViewer != nil {
			m.diffViewer.ScrollUp()
		}
	case "down", "j":
		if m.diffViewer != nil {
			m.diffViewer.ScrollDown()
		}
	case "left", "h":
		if m.diffViewer != nil {
			m.diffViewer.SelectPrev()
		}
	case "right", "l":
		if m.diffViewer != nil {
			m.diffViewer.SelectNext()
		}
	}
	return m, nil
}

func (m DiffModel) handleErrorKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "r":
		// Retry loading files
		m.state = DiffStateFileSelection
		m.errorMsg = ""
		return m, m.loadExcelFiles
	}
	return m, nil
}

func (m DiffModel) loadAndCompareDiff() tea.Msg {
	// Load both Excel files and compute diff
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Reduce noise in TUI
	conv := converter.NewConverter(logger)

	doc1, err := conv.ExcelToJSON(m.selectedFile1, converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
		IgnoreEmptyCells: false,
	})
	if err != nil {
		return errMsg{fmt.Errorf("failed to load %s: %w", m.selectedFile1, err)}
	}

	doc2, err := conv.ExcelToJSON(m.selectedFile2, converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
		IgnoreEmptyCells: false,
	})
	if err != nil {
		return errMsg{fmt.Errorf("failed to load %s: %w", m.selectedFile2, err)}
	}

	diff := models.ComputeDiff(doc1, doc2)
	return diffComputedMsg{diff}
}

type diffComputedMsg struct {
	diff *models.ExcelDiff
}

func (m DiffModel) View() string {
	switch m.state {
	case DiffStateFileSelection:
		return m.renderFileSelection()
	case DiffStateViewing:
		return m.renderDiffViewer()
	case DiffStateError:
		return m.renderError()
	}
	return "Loading..."
}

func (m DiffModel) renderFileSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)

	title := titleStyle.Render("Excel File Diff Viewer")

	instructions := ""
	if m.selectedFile1 == "" {
		instructions = styles.MutedStyle.Render("Select first file to compare:")
	} else if m.selectedFile2 == "" {
		instructions = fmt.Sprintf("First file: %s\n%s",
			styles.SuccessStyle.Render(m.selectedFile1),
			styles.MutedStyle.Render("Select second file to compare:"))
	}

	if len(m.files) == 0 {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			styles.ErrorStyle.Render("No Excel files found in current directory"),
			"",
			styles.HelpStyle.Render("[q] Back to menu • [r] Retry"),
		)
		return styles.Center(m.width, m.height, content)
	}

	// File list
	fileList := make([]string, 0, len(m.files))
	for i, file := range m.files {
		cursor := "  "
		fileStyle := lipgloss.NewStyle()

		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(styles.Primary).Render("▶ ")
			fileStyle = fileStyle.Bold(true)
		}

		// Mark selected files
		if file == m.selectedFile1 {
			fileStyle = fileStyle.Foreground(styles.Success)
			cursor += "[1] "
		} else if file == m.selectedFile2 {
			fileStyle = fileStyle.Foreground(styles.Success)
			cursor += "[2] "
		}

		fileList = append(fileList, cursor+fileStyle.Render(file))
	}

	files := strings.Join(fileList, "\n")

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Enter/Space] Select • [r] Reset selection • [h/?] Toggle help • [q] Back to menu",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [q] Back to menu")
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

func (m DiffModel) renderDiffViewer() string {
	if m.diffViewer == nil {
		return "Loading diff..."
	}
	return m.diffViewer.View()
}

func (m DiffModel) renderError() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		styles.TitleStyle.Render("Error"),
		"",
		styles.ErrorStyle.Render(m.errorMsg),
		"",
		styles.HelpStyle.Render("[r] Retry • [q] Back to menu"),
	)
	return styles.Center(m.width, m.height, content)
}

// Handle diff computed message
func (m DiffModel) handleDiffComputed(diff *models.ExcelDiff) (DiffModel, tea.Cmd) {
	viewer := components.NewDiffViewer(diff)
	viewer.SetDimensions(m.width, m.height)
	m.diffViewer = &viewer
	m.state = DiffStateViewing

	// Log the diff comparison
	utils.LogUserAction("diff_comparison", map[string]any{
		"file1":         m.selectedFile1,
		"file2":         m.selectedFile2,
		"has_changes":   diff.HasChanges(),
		"total_changes": diff.Summary.TotalChanges,
		"cell_changes":  diff.Summary.CellChanges,
	})

	return m, nil
}
