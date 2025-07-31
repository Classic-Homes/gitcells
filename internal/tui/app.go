package tui

import (
	"fmt"

	"github.com/Classic-Homes/gitcells/internal/tui/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Mode int

const (
	ModeMenu Mode = iota
	ModeSetup
	ModeDashboard
	ModeBranch
	ModeConflict
)

type Model struct {
	mode        Mode
	width       int
	height      int
	quitting    bool
	menuCursor  int
	setupModel  tea.Model
	dashModel   tea.Model
	branchModel tea.Model
	conflModel  tea.Model
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
	{"Setup Wizard", "Configure GitCells for your repository", ModeSetup},
	{"Status Dashboard", "Monitor file sync and conversion status", ModeDashboard},
	{"Branch Management", "Create, switch, and merge branches", ModeBranch},
	{"Conflict Resolution", "Resolve Excel merge conflicts", ModeConflict},
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
				return m, changeMode(selectedMode)
			}
		} else {
			switch msg.String() {
			case "esc":
				return m, backToMenu()
			}
		}

	case modeChangeMsg:
		m.mode = msg.mode
		switch m.mode {
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
		case ModeBranch:
			if m.branchModel == nil {
				m.branchModel = models.NewBranchEnhancedModel()
			}
			return m, m.branchModel.Init()
		case ModeConflict:
			if m.conflModel == nil {
				m.conflModel = models.NewConflictEnhancedModel()
			}
			return m, m.conflModel.Init()
		}

	case backToMenuMsg:
		m.mode = ModeMenu
		return m, nil
	}

	var cmd tea.Cmd
	switch m.mode {
	case ModeSetup:
		if m.setupModel != nil {
			m.setupModel, cmd = m.setupModel.Update(msg)
		}
	case ModeDashboard:
		if m.dashModel != nil {
			m.dashModel, cmd = m.dashModel.Update(msg)
		}
	case ModeBranch:
		if m.branchModel != nil {
			m.branchModel, cmd = m.branchModel.Update(msg)
		}
	case ModeConflict:
		if m.conflModel != nil {
			m.conflModel, cmd = m.conflModel.Update(msg)
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
	case ModeBranch:
		if m.branchModel != nil {
			return m.branchModel.View()
		}
	case ModeConflict:
		if m.conflModel != nil {
			return m.conflModel.View()
		}
	}

	return "Loading..."
}

func (m Model) renderMenu() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	menuStyle := lipgloss.NewStyle().
		Padding(2, 4)

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	s := titleStyle.Render("GitCells TUI") + "\n\n"

	for i, item := range menuItems {
		cursor := "  "
		if i == m.menuCursor {
			cursor = cursorStyle.Render("▶ ")
		}
		s += fmt.Sprintf("%s%s\n", cursor, item.title)
		s += fmt.Sprintf("    %s\n\n", descStyle.Render(item.desc))
	}

	s += "\n" + descStyle.Render("Use ↑/↓ or j/k to navigate, Enter to select, q to quit")

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
