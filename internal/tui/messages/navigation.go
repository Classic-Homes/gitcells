package messages

import tea "github.com/charmbracelet/bubbletea"

// RequestMainMenuMsg is sent when a model wants to return to the main menu
type RequestMainMenuMsg struct{}

// RequestMainMenu returns a command that sends RequestMainMenuMsg
func RequestMainMenu() tea.Cmd {
	return func() tea.Msg {
		return RequestMainMenuMsg{}
	}
}
