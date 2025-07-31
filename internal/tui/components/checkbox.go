package components

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Checkbox struct {
	label   string
	checked bool
	focused bool
}

func NewCheckbox(label string, checked bool) Checkbox {
	return Checkbox{
		label:   label,
		checked: checked,
	}
}

func (c *Checkbox) Toggle() {
	c.checked = !c.checked
}

func (c *Checkbox) SetChecked(checked bool) {
	c.checked = checked
}

func (c *Checkbox) Checked() bool {
	return c.checked
}

func (c *Checkbox) Focus() {
	c.focused = true
}

func (c *Checkbox) Blur() {
	c.focused = false
}

func (c *Checkbox) Update(msg tea.Msg) (Checkbox, tea.Cmd) {
	if !c.focused {
		return *c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case " ", "enter":
			c.Toggle()
		}
	}

	return *c, nil
}

func (c Checkbox) View() string {
	checkStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("82"))
	
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	checkbox := "[ ]"
	if c.checked {
		checkbox = checkStyle.Render("[✓]")
	}

	label := c.label
	if c.focused {
		label = focusedStyle.Render(label)
	} else {
		label = labelStyle.Render(label)
	}

	return checkbox + " " + label
}