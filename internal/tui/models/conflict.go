package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConflictModel struct {
	width      int
	height     int
	conflicts  []Conflict
	current    int
	resolution string
}

type Conflict struct {
	File       string
	Sheet      string
	Cell       string
	OurValue   string
	TheirValue string
}

func NewConflictModel() ConflictModel {
	return ConflictModel{
		conflicts: []Conflict{
			{
				File:       "Budget2024.xlsx",
				Sheet:      "Summary",
				Cell:       "B15",
				OurValue:   "=$B$10+$B$11+$B$12",
				TheirValue: "=$B$10+$B$11+$B$12+$B$13",
			},
			{
				File:       "Budget2024.xlsx",
				Sheet:      "Q1",
				Cell:       "D20",
				OurValue:   "1,250,000",
				TheirValue: "1,275,000",
			},
		},
		current: 0,
	}
}

func (m ConflictModel) Init() tea.Cmd {
	return nil
}

func (m ConflictModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l":
			if m.current < len(m.conflicts)-1 {
				m.current++
			}
		case "left", "h":
			if m.current > 0 {
				m.current--
			}
		case "o":
			m.resolution = "ours"
		case "t":
			m.resolution = "theirs"
		}
	}

	return m, nil
}

func (m ConflictModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	conflictStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196"))

	oursStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))

	theirsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("214"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("241")).
		Padding(1, 2)

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4)

	conflict := m.conflicts[m.current]

	s := titleStyle.Render("Conflict Resolution") + "\n\n"
	s += conflictStyle.Render(fmt.Sprintf("Conflict %d of %d", m.current+1, len(m.conflicts))) + "\n\n"
	s += fmt.Sprintf("File: %s\nSheet: %s\nCell: %s\n\n", conflict.File, conflict.Sheet, conflict.Cell)

	oursBox := boxStyle.Render(
		oursStyle.Render("Our Version") + "\n\n" + conflict.OurValue,
	)

	theirsBox := boxStyle.Render(
		theirsStyle.Render("Their Version") + "\n\n" + conflict.TheirValue,
	)

	s += lipgloss.JoinHorizontal(lipgloss.Top, oursBox, "  ", theirsBox) + "\n\n"

	if m.resolution != "" {
		s += fmt.Sprintf("Resolution: %s\n\n", m.resolution)
	}

	s += infoStyle.Render("Actions: [o]urs [t]heirs [e]dit | Navigate: ←/→ or h/l | Press Esc to go back")

	return containerStyle.Render(s)
}
