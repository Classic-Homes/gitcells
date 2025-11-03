package tui

import (
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	ModeDashboard Mode = iota
	ModeDiff
	ModeSettings
	ModeErrorLog
	ModeSetup
)

type Model struct {
	mode          Mode
	width         int
	height        int
	dashModel     tea.Model
	diffModel     tea.Model
	settingsModel tea.Model
	errorLogModel tea.Model
	setupModel    tea.Model
}

func NewModel() *Model {
	return &Model{
		mode: ModeDashboard,
	}
}

func (m *Model) Init() tea.Cmd {
	// Initialize dashboard immediately
	dashModel := models.NewDashboardModel()
	m.dashModel = dashModel
	return dashModel.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward to active model
		return m.forwardToActiveModel(msg)

	case tea.QuitMsg:
		return m, tea.Quit

	case messages.RequestMainMenuMsg:
		// Go back to dashboard instead of menu
		m.mode = ModeDashboard
		if m.dashModel == nil {
			m.dashModel = models.NewDashboardModel()
			return m, m.dashModel.Init()
		}
		return m, nil

	case messages.RequestModeChangeMsg:
		switch msg.Mode {
		case "diff":
			m.mode = ModeDiff
			if m.diffModel == nil {
				m.diffModel = models.NewDiffModel()
			}
			return m, m.diffModel.Init()

		case "settings":
			m.mode = ModeSettings
			if m.settingsModel == nil {
				m.settingsModel = models.NewSettingsModel()
			}
			return m, m.settingsModel.Init()

		case "errorlog":
			m.mode = ModeErrorLog
			if m.errorLogModel == nil {
				m.errorLogModel = models.NewErrorLogModel()
			}
			return m, m.errorLogModel.Init()

		case "setup":
			m.mode = ModeSetup
			if m.setupModel == nil {
				m.setupModel = models.NewSetupModel()
			}
			return m, m.setupModel.Init()

		default:
			// Default to dashboard
			m.mode = ModeDashboard
			return m, nil
		}
	}

	// Forward all other messages to active model
	return m.forwardToActiveModel(msg)
}

func (m *Model) forwardToActiveModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case ModeDashboard:
		if m.dashModel != nil {
			m.dashModel, cmd = m.dashModel.Update(msg)
		}
	case ModeDiff:
		if m.diffModel != nil {
			m.diffModel, cmd = m.diffModel.Update(msg)
		}
	case ModeSettings:
		if m.settingsModel != nil {
			m.settingsModel, cmd = m.settingsModel.Update(msg)
		}
	case ModeErrorLog:
		if m.errorLogModel != nil {
			m.errorLogModel, cmd = m.errorLogModel.Update(msg)
		}
	case ModeSetup:
		if m.setupModel != nil {
			m.setupModel, cmd = m.setupModel.Update(msg)
		}
	}

	return m, cmd
}

func (m *Model) View() string {
	switch m.mode {
	case ModeDashboard:
		if m.dashModel != nil {
			return m.dashModel.View()
		}
	case ModeDiff:
		if m.diffModel != nil {
			return m.diffModel.View()
		}
	case ModeSettings:
		if m.settingsModel != nil {
			return m.settingsModel.View()
		}
	case ModeErrorLog:
		if m.errorLogModel != nil {
			return m.errorLogModel.View()
		}
	case ModeSetup:
		if m.setupModel != nil {
			return m.setupModel.View()
		}
	}

	return "Loading..."
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
