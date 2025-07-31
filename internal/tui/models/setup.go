package models

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SetupModel struct {
	step     int
	width    int
	height   int
	cursor   int
	inputs   map[string]string
	finished bool
}

type SetupStep struct {
	title       string
	description string
	inputs      []string
}

var setupSteps = []SetupStep{
	{
		title:       "Welcome to GitCells Setup",
		description: "This wizard will help you configure GitCells for your repository",
		inputs:      []string{},
	},
	{
		title:       "Select Repository Directory",
		description: "Choose the directory containing your Excel files",
		inputs:      []string{"directory"},
	},
	{
		title:       "Configure Excel Patterns",
		description: "Specify which Excel files to track (e.g., *.xlsx, reports/*.xlsx)",
		inputs:      []string{"pattern"},
	},
	{
		title:       "Git Integration Settings",
		description: "Configure how GitCells interacts with Git",
		inputs:      []string{"auto_commit", "auto_push", "commit_template"},
	},
	{
		title:       "Review Configuration",
		description: "Review your settings before initializing",
		inputs:      []string{},
	},
}

func NewSetupModel() SetupModel {
	return SetupModel{
		step:   0,
		inputs: make(map[string]string),
	}
}

func (m SetupModel) Init() tea.Cmd {
	return nil
}

func (m SetupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab", "down":
			m.cursor++
			if m.cursor >= len(setupSteps[m.step].inputs) {
				m.cursor = 0
			}
		case "shift+tab", "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(setupSteps[m.step].inputs) - 1
			}
		case "enter":
			if m.step < len(setupSteps)-1 {
				m.step++
				m.cursor = 0
			} else {
				m.finished = true
			}
		case "backspace":
			if m.step > 0 {
				m.step--
				m.cursor = 0
			}
		}
	}

	return m, nil
}

func (m SetupModel) View() string {
	if m.finished {
		return m.renderComplete()
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("99")).
		MarginBottom(1)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginBottom(2)

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4)

	step := setupSteps[m.step]
	s := headerStyle.Render(fmt.Sprintf("Step %d/%d: %s", m.step+1, len(setupSteps), step.title)) + "\n"
	s += descStyle.Render(step.description) + "\n"

	if len(step.inputs) > 0 {
		s += "\n"
		for i, input := range step.inputs {
			cursor := "  "
			if i == m.cursor {
				cursor = "▶ "
			}
			value := m.inputs[input]
			if value == "" {
				value = "<empty>"
			}
			s += fmt.Sprintf("%s%s: %s\n", cursor, input, value)
		}
	}

	s += "\n\n" + descStyle.Render("Press Enter to continue, Backspace to go back, Esc to return to menu")

	return containerStyle.Render(s)
}

func (m SetupModel) renderComplete() string {
	completeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("82")).
		Padding(2, 4)

	return completeStyle.Render("✓ Setup Complete!\n\nPress Esc to return to menu")
}
