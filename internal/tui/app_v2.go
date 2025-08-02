package tui

import (
	"github.com/Classic-Homes/gitcells/internal/tui/messages"
	"github.com/Classic-Homes/gitcells/internal/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

type ModeV2 int

const (
	ModeV2Dashboard ModeV2 = iota
	ModeV2Diff
	ModeV2Settings
	ModeV2ErrorLog
	ModeV2Setup
)

type ModelV2 struct {
	mode          ModeV2
	width         int
	height        int
	dashModel     tea.Model
	diffModel     tea.Model
	settingsModel tea.Model
	errorLogModel tea.Model
	setupModel    tea.Model
}

func NewModelV2() *ModelV2 {
	return &ModelV2{
		mode: ModeV2Dashboard,
	}
}

func (m *ModelV2) Init() tea.Cmd {
	// Initialize dashboard immediately
	dashModel := models.NewUnifiedDashboardModel()
	m.dashModel = dashModel
	return dashModel.Init()
}

func (m *ModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.mode = ModeV2Dashboard
		if m.dashModel == nil {
			m.dashModel = models.NewUnifiedDashboardModel()
			return m, m.dashModel.Init()
		}
		return m, nil

	case messages.RequestModeChangeMsg:
		switch msg.Mode {
		case "diff":
			m.mode = ModeV2Diff
			if m.diffModel == nil {
				m.diffModel = models.NewDiffModel()
			}
			return m, m.diffModel.Init()

		case "settings":
			m.mode = ModeV2Settings
			if m.settingsModel == nil {
				m.settingsModel = models.NewSettingsModelV2()
			}
			return m, m.settingsModel.Init()

		case "errorlog":
			m.mode = ModeV2ErrorLog
			if m.errorLogModel == nil {
				m.errorLogModel = models.NewErrorLogModelV2()
			}
			return m, m.errorLogModel.Init()

		case "setup":
			m.mode = ModeV2Setup
			if m.setupModel == nil {
				m.setupModel = models.NewSetupModel()
			}
			return m, m.setupModel.Init()

		default:
			// Default to dashboard
			m.mode = ModeV2Dashboard
			return m, nil
		}
	}

	// Forward all other messages to active model
	return m.forwardToActiveModel(msg)
}

func (m *ModelV2) forwardToActiveModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.mode {
	case ModeV2Dashboard:
		if m.dashModel != nil {
			m.dashModel, cmd = m.dashModel.Update(msg)
		}
	case ModeV2Diff:
		if m.diffModel != nil {
			m.diffModel, cmd = m.diffModel.Update(msg)
		}
	case ModeV2Settings:
		if m.settingsModel != nil {
			m.settingsModel, cmd = m.settingsModel.Update(msg)
		}
	case ModeV2ErrorLog:
		if m.errorLogModel != nil {
			m.errorLogModel, cmd = m.errorLogModel.Update(msg)
		}
	case ModeV2Setup:
		if m.setupModel != nil {
			m.setupModel, cmd = m.setupModel.Update(msg)
		}
	}

	return m, cmd
}

func (m *ModelV2) View() string {
	switch m.mode {
	case ModeV2Dashboard:
		if m.dashModel != nil {
			return m.dashModel.View()
		}
	case ModeV2Diff:
		if m.diffModel != nil {
			return m.diffModel.View()
		}
	case ModeV2Settings:
		if m.settingsModel != nil {
			return m.settingsModel.View()
		}
	case ModeV2ErrorLog:
		if m.errorLogModel != nil {
			return m.errorLogModel.View()
		}
	case ModeV2Setup:
		if m.setupModel != nil {
			return m.setupModel.View()
		}
	}

	return "Loading..."
}

func RunV2() error {
	p := tea.NewProgram(NewModelV2(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
