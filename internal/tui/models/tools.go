package models

import (
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ToolsState int

const (
	ToolsStateSelection ToolsState = iota
	ToolsStateDiff
	ToolsStateConversion
)

type ToolsModel struct {
	state           ToolsState
	width           int
	height          int
	cursor          int
	showHelp        bool
	diffModel       DiffModel
	conversionModel ManualConversionModel
}

var toolsMenuItems = []struct {
	title string
	desc  string
}{
	{"Diff Viewer", "Compare Excel files and view differences side-by-side"},
	{"Manual Conversions", "Convert specific Excel files with custom options"},
}

func NewToolsModel() ToolsModel {
	return ToolsModel{
		state:    ToolsStateSelection,
		showHelp: true,
	}
}

func (m ToolsModel) Init() tea.Cmd {
	return nil
}

func (m ToolsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Pass size to sub-models
		if m.state == ToolsStateDiff {
			m.diffModel.Update(msg)
		} else if m.state == ToolsStateConversion {
			m.conversionModel.Update(msg)
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	default:
		// Handle messages for sub-models
		var cmd tea.Cmd
		if m.state == ToolsStateDiff {
			diffModel, diffCmd := m.diffModel.Update(msg)
			if diffModel, ok := diffModel.(DiffModel); ok {
				m.diffModel = diffModel
				cmd = diffCmd
			}
		} else if m.state == ToolsStateConversion {
			convModel, convCmd := m.conversionModel.Update(msg)
			if convModel, ok := convModel.(ManualConversionModel); ok {
				m.conversionModel = convModel
				cmd = convCmd
			}
		}
		return m, cmd
	}
}

func (m ToolsModel) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case ToolsStateSelection:
		return m.handleSelectionKeys(msg)
	case ToolsStateDiff:
		// Handle keys that might return to tools menu
		if msg.String() == "ctrl+c" || msg.String() == "q" || msg.String() == "esc" {
			m.state = ToolsStateSelection
			return m, nil
		}
		// Pass other keys to diff model
		diffModel, cmd := m.diffModel.handleKeyPress(msg)
		if diffModel, ok := diffModel.(DiffModel); ok {
			m.diffModel = diffModel
		}
		return m, cmd
	case ToolsStateConversion:
		// Handle keys that might return to tools menu
		if msg.String() == "ctrl+c" || msg.String() == "q" || msg.String() == "esc" {
			m.state = ToolsStateSelection
			return m, nil
		}
		// Pass other keys to conversion model
		convModel, cmd := m.conversionModel.handleKeyPress(msg)
		if convModel, ok := convModel.(ManualConversionModel); ok {
			m.conversionModel = convModel
		}
		return m, cmd
	}
	return m, nil
}

func (m ToolsModel) handleSelectionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		return m, func() tea.Msg { return messages.RequestMainMenuMsg{} }
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(toolsMenuItems)-1 {
			m.cursor++
		}
	case "enter", " ":
		switch m.cursor {
		case 0: // Diff Viewer
			if m.diffModel.state == 0 && len(m.diffModel.files) == 0 {
				m.diffModel = NewDiffModel()
			}
			m.state = ToolsStateDiff
			return m, m.diffModel.Init()
		case 1: // Manual Conversions
			if m.conversionModel.state == 0 && len(m.conversionModel.files) == 0 {
				m.conversionModel = NewManualConversionModel()
			}
			m.state = ToolsStateConversion
			return m, m.conversionModel.Init()
		}
	case "h", "?":
		m.showHelp = !m.showHelp
	}
	return m, nil
}

func (m ToolsModel) View() string {
	switch m.state {
	case ToolsStateSelection:
		return m.renderSelection()
	case ToolsStateDiff:
		return m.diffModel.View()
	case ToolsStateConversion:
		return m.conversionModel.View()
	}
	return "Loading..."
}

func (m ToolsModel) renderSelection() string {
	titleStyle := styles.TitleStyle.MarginBottom(1)

	title := titleStyle.Render("Tools")

	menuStyle := lipgloss.NewStyle().
		Padding(2, 4)

	cursorStyle := lipgloss.NewStyle().
		Foreground(styles.Primary)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.Muted)

	content := title + "\n\n"

	for i, item := range toolsMenuItems {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▶ ")
		}
		content += cursor + item.title + "\n"
		content += "    " + descStyle.Render(item.desc) + "\n\n"
	}

	help := ""
	if m.showHelp {
		help = styles.HelpStyle.Render(
			"[↑/↓] Navigate • [Enter/Space] Select • [h/?] Toggle help • [q] Back to menu",
		)
	} else {
		help = styles.HelpStyle.Render("[h/?] Show help • [q] Back to menu")
	}

	content += "\n" + help

	return menuStyle.Render(content)
}
