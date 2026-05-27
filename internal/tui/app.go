// Package tui hosts a minimal dashboard: watcher status + toggle, file
// list, recent activity. Heavier git operations are intentionally not
// surfaced here — pair GitCells with GitHub Desktop (or any git GUI) for
// diff/commit/sync work.
package tui

import (
	"github.com/Classic-Homes/gitcells/internal/tui/models"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	width     int
	height    int
	dashModel tea.Model
}

func NewModel() *Model {
	return &Model{}
}

func (m *Model) Init() tea.Cmd {
	m.dashModel = models.NewDashboardModel()
	return m.dashModel.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
	}
	if _, ok := msg.(tea.QuitMsg); ok {
		return m, tea.Quit
	}

	var cmd tea.Cmd
	if m.dashModel != nil {
		m.dashModel, cmd = m.dashModel.Update(msg)
	}
	return m, cmd
}

func (m *Model) View() string {
	if m.dashModel != nil {
		return m.dashModel.View()
	}
	return "Loading..."
}

func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
