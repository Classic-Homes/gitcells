package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BranchModel struct {
	width    int
	height   int
	branches []Branch
	cursor   int
}

type Branch struct {
	Name       string
	Current    bool
	HasChanges bool
	Conflicts  int
}

func NewBranchModel() BranchModel {
	return BranchModel{
		branches: []Branch{
			{Name: "main", Current: true, HasChanges: false, Conflicts: 0},
			{Name: "feature/budget-2024", Current: false, HasChanges: true, Conflicts: 0},
			{Name: "feature/quarterly-report", Current: false, HasChanges: false, Conflicts: 2},
			{Name: "fix/formula-updates", Current: false, HasChanges: true, Conflicts: 0},
		},
		cursor: 0,
	}
}

func (m BranchModel) Init() tea.Cmd {
	return nil
}

func (m BranchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.branches)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

func (m BranchModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	currentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82")).
		Bold(true)

	changesStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	conflictStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4)

	s := titleStyle.Render("Branch Management") + "\n\n"

	for i, branch := range m.branches {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}

		name := branch.Name
		if branch.Current {
			name = currentStyle.Render("* " + name)
		} else {
			name = "  " + name
		}

		status := ""
		if branch.HasChanges {
			status += changesStyle.Render(" [changes]")
		}
		if branch.Conflicts > 0 {
			status += conflictStyle.Render(fmt.Sprintf(" [%d conflicts]", branch.Conflicts))
		}

		s += cursor + name + status + "\n"
	}

	s += "\n\n" + infoStyle.Render("Actions: [n]ew [s]witch [m]erge [d]elete | Navigate: ↑/↓ or j/k | Press Esc to go back")

	return containerStyle.Render(s)
}