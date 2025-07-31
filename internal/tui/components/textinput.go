package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TextInput struct {
	textinput textinput.Model
	label     string
	focused   bool
}

func NewTextInput(label, placeholder string) TextInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 50

	return TextInput{
		textinput: ti,
		label:     label,
	}
}

func (t *TextInput) SetValue(value string) {
	t.textinput.SetValue(value)
}

func (t *TextInput) Value() string {
	return t.textinput.Value()
}

func (t *TextInput) Focus() tea.Cmd {
	t.focused = true
	return t.textinput.Focus()
}

func (t *TextInput) Blur() {
	t.focused = false
	t.textinput.Blur()
}

func (t *TextInput) Update(msg tea.Msg) (TextInput, tea.Cmd) {
	var cmd tea.Cmd
	t.textinput, cmd = t.textinput.Update(msg)
	return *t, cmd
}

func (t TextInput) View() string {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	
	focusedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99"))

	label := t.label
	if t.focused {
		label = focusedStyle.Render(label)
	} else {
		label = labelStyle.Render(label)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		label,
		t.textinput.View(),
	)
}