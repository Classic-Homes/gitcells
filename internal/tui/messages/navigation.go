package messages

import tea "github.com/charmbracelet/bubbletea"

// RequestMainMenuMsg is sent when a model wants to return to the main menu
type RequestMainMenuMsg struct{}

// RequestModeChangeMsg is sent when a model wants to switch to a specific mode
type RequestModeChangeMsg struct {
	Mode string
}

// RequestMainMenu returns a command that sends RequestMainMenuMsg
func RequestMainMenu() tea.Cmd {
	return func() tea.Msg {
		return RequestMainMenuMsg{}
	}
}

// RequestModeChange returns a command that sends RequestModeChangeMsg
func RequestModeChange(mode string) tea.Cmd {
	return func() tea.Msg {
		return RequestModeChangeMsg{Mode: mode}
	}
}
