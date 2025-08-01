package tui

import (
	"fmt"

	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/models"
	"github.com/Classic-Homes/gitcells/internal/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeMenu Mode = iota
	ModeSetup
	ModeDashboard
	ModeDiff
	ModeSettings
	ModeErrorLog
)

type Model struct {
	mode          Mode
	width         int
	height        int
	quitting      bool
	menuCursor    int
	setupModel    tea.Model
	dashModel     tea.Model
	diffModel     tea.Model
	settingsModel tea.Model
	errorLogModel tea.Model
}

type modeChangeMsg struct {
	mode Mode
}

type backToMenuMsg struct{}

var menuItems = []struct {
	title string
	desc  string
	mode  Mode
}{
	{"Setup Wizard", "Configure GitCells for your Excel tracking repository", ModeSetup},
	{"Status Dashboard", "Monitor Excel file tracking and conversion status", ModeDashboard},
	{"Diff Viewer", "Compare Excel files and view differences side-by-side", ModeDiff},
	{"Settings", "Update, uninstall, and manage GitCells system settings", ModeSettings},
	{"Error Logs", "View application errors and troubleshooting information", ModeErrorLog},
}

func NewModel() Model {
	return Model{
		mode: ModeMenu,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.mode == ModeMenu {
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.menuCursor > 0 {
					m.menuCursor--
				}
			case "down", "j":
				if m.menuCursor < len(menuItems)-1 {
					m.menuCursor++
				}
			case "enter":
				selectedMode := menuItems[m.menuCursor].mode
				utils.LogUserAction("menu_select", map[string]any{
					"selected_item": m.menuCursor,
					"mode":          selectedMode,
				})
				return m, changeMode(selectedMode)
			}
		} else if msg.String() == "ctrl+l" {
			// For non-menu modes, let the individual models handle keys first
			// Only handle global shortcuts that don't conflict with model navigation
			if m.mode != ModeErrorLog {
				return m, changeMode(ModeErrorLog)
			}
		}

	case modeChangeMsg:
		oldMode := m.mode
		m.mode = msg.mode
		utils.LogModeChange(fmt.Sprintf("%d", oldMode), fmt.Sprintf("%d", msg.mode))
		switch m.mode {
		case ModeMenu:
			return m, nil
		case ModeSetup:
			if m.setupModel == nil {
				setupModel := models.NewSetupEnhancedModel()
				m.setupModel = setupModel
			}
			return m, m.setupModel.Init()
		case ModeDashboard:
			if m.dashModel == nil {
				m.dashModel = models.NewDashboardEnhancedModel()
			}
			return m, m.dashModel.Init()
		case ModeDiff:
			if m.diffModel == nil {
				m.diffModel = models.NewDiffModel()
			}
			return m, m.diffModel.Init()
		case ModeErrorLog:
			if m.errorLogModel == nil {
				m.errorLogModel = models.NewErrorLogEnhancedModel()
			}
			return m, m.errorLogModel.Init()
		case ModeSettings:
			if m.settingsModel == nil {
				m.settingsModel = models.NewSettingsModel()
			} else {
				// Reset to main view when re-entering settings
				if settingsModel, ok := m.settingsModel.(models.SettingsModel); ok {
					settingsModel = settingsModel.ResetToMainView()
					m.settingsModel = settingsModel
				}
			}
			return m, m.settingsModel.Init()
		}

	case backToMenuMsg:
		m.mode = ModeMenu
		return m, nil

	case messages.RequestMainMenuMsg:
		return m, backToMenu()
	}

	var cmd tea.Cmd
	switch m.mode {
	case ModeMenu:
		// Menu mode is handled above
	case ModeSetup:
		if m.setupModel != nil {
			m.setupModel, cmd = m.setupModel.Update(msg)
		}
	case ModeDashboard:
		if m.dashModel != nil {
			m.dashModel, cmd = m.dashModel.Update(msg)
		}
	case ModeDiff:
		if m.diffModel != nil {
			m.diffModel, cmd = m.diffModel.Update(msg)
		}
	case ModeErrorLog:
		if m.errorLogModel != nil {
			m.errorLogModel, cmd = m.errorLogModel.Update(msg)
		}
	case ModeSettings:
		if m.settingsModel != nil {
			m.settingsModel, cmd = m.settingsModel.Update(msg)
		}
	}

	return m, cmd
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	switch m.mode {
	case ModeMenu:
		return m.renderMenu()
	case ModeSetup:
		if m.setupModel != nil {
			return m.setupModel.View()
		}
	case ModeDashboard:
		if m.dashModel != nil {
			return m.dashModel.View()
		}
	case ModeDiff:
		if m.diffModel != nil {
			return m.diffModel.View()
		}
	case ModeErrorLog:
		if m.errorLogModel != nil {
			return m.errorLogModel.View()
		}
	case ModeSettings:
		if m.settingsModel != nil {
			return m.settingsModel.View()
		}
	}

	return "Loading..."
}

func (m Model) renderMenu() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	subtitleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(2)

	menuStyle := lipgloss.NewStyle().
		Padding(2, 4)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	s := titleStyle.Render("GitCells") + "\n"
	s += subtitleStyle.Render("Excel Version Control Management") + "\n\n"

	for i, item := range menuItems {
		cursor := "  "
		if i == m.menuCursor {
			cursor = cursorStyle.Render("▶ ")
		}
		s += fmt.Sprintf("%s%s\n", cursor, item.title)
		s += fmt.Sprintf("    %s\n\n", descStyle.Render(item.desc))
	}

	s += "\n" + descStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, Ctrl+L for error logs, q to quit")

	return menuStyle.Render(s)
}

func changeMode(mode Mode) tea.Cmd {
	return func() tea.Msg {
		return modeChangeMsg{mode: mode}
	}
}

func backToMenu() tea.Cmd {
	return func() tea.Msg {
		return backToMenuMsg{}
	}
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
