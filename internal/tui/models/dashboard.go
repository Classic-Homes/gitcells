package models

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type DashboardModel struct {
	width       int
	height      int
	watching    int
	files       int
	synced      bool
	lastCommit  time.Time
	operations  []Operation
	tickCount   int
}

type Operation struct {
	Type     string
	FileName string
	Status   string
	Progress int
}

type tickMsg time.Time

func NewDashboardModel() DashboardModel {
	return DashboardModel{
		watching:   3,
		files:      15,
		synced:     true,
		lastCommit: time.Now().Add(-2 * time.Minute),
		operations: []Operation{
			{Type: "Converting", FileName: "Budget2024.xlsx", Status: "In Progress", Progress: 45},
			{Type: "Completed", FileName: "Report.xlsx", Status: "Success", Progress: 100},
			{Type: "Skipped", FileName: "~$TempFile.xlsx", Status: "Temp File", Progress: 0},
		},
	}
}

func (m DashboardModel) Init() tea.Cmd {
	return tick()
}

func (m DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.tickCount++
		if m.operations[0].Progress < 100 {
			m.operations[0].Progress += 5
		}
		return m, tick()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m DashboardModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99"))

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(1, 2).
		Width(60)

	s := boxStyle.Render(
		titleStyle.Render("GitCells Status Dashboard") + "\n\n" +
			fmt.Sprintf("Watching: %d directories, %d Excel files\n", m.watching, m.files) +
			fmt.Sprintf("Status: %s Synced | Last commit: %s ago\n\n",
				statusStyle.Render("✓"),
				formatDuration(time.Since(m.lastCommit))) +
			"File Operations:\n" +
			m.renderOperations() + "\n\n" +
			infoStyle.Render("[w]atch [c]onvert [s]ync [q]uit [?]help"),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
}

func (m DashboardModel) renderOperations() string {
	s := ""
	for _, op := range m.operations {
		icon := "  "
		style := lipgloss.NewStyle()

		switch op.Type {
		case "Converting":
			icon = "► "
			style = style.Foreground(lipgloss.Color("214"))
		case "Completed":
			icon = "✓ "
			style = style.Foreground(lipgloss.Color("82"))
		case "Skipped":
			icon = "⚠ "
			style = style.Foreground(lipgloss.Color("241"))
		}

		line := fmt.Sprintf("%s%s: %s", icon, op.Type, op.FileName)
		if op.Progress > 0 && op.Progress < 100 {
			line += fmt.Sprintf(" (%d%%)", op.Progress)
		}
		if op.Status != "" && op.Type != "Converting" {
			line += fmt.Sprintf(" (%s)", op.Status)
		}
		s += style.Render(line) + "\n"
	}
	return s
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d sec", int(d.Seconds()))
	}
	return fmt.Sprintf("%d min", int(d.Minutes()))
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}