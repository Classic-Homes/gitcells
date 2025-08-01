package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("99")
	Secondary = lipgloss.Color("214")
	Success   = lipgloss.Color("82")
	Warning   = lipgloss.Color("214")
	Error     = lipgloss.Color("196")
	Muted     = lipgloss.Color("241")
	Info      = lipgloss.Color("39")
	Highlight = lipgloss.Color("212")
	
	// Color values for direct use
	MutedColor     = lipgloss.Color("241")
	HighlightColor = lipgloss.Color("212")

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			MarginBottom(1)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted).
			Padding(1, 2)

	FocusedBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	ButtonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(Primary).
			Padding(0, 2)

	ActiveButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230")).
				Background(Secondary).
				Padding(0, 2)

	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	InfoStyle = lipgloss.NewStyle().
			Foreground(Info)

	HighlightStyle = lipgloss.NewStyle().
			Foreground(Highlight)

	SelectedStyle = lipgloss.NewStyle().
			Background(Primary).
			Foreground(lipgloss.Color("230")).
			Bold(true)

	ViewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Muted)

	// Error log specific styles
	LogTimestampStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	LogFileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Italic(true)

	LogFieldKeyStyle = lipgloss.NewStyle().
				Foreground(Primary)

	LogFieldValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("251"))

	LogSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("238"))

	LogLevelErrorStyle = lipgloss.NewStyle().
				Foreground(Error).
				Bold(true).
				Width(5)

	LogLevelWarnStyle = lipgloss.NewStyle().
				Foreground(Warning).
				Bold(true).
				Width(5)

	LogLevelInfoStyle = lipgloss.NewStyle().
				Foreground(Info).
				Width(5)

	LogLevelDebugStyle = lipgloss.NewStyle().
				Foreground(Muted).
				Width(5)

	LogEntryStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(1)

	LogSelectedEntryStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("237")).
				Foreground(lipgloss.Color("255")).
				PaddingLeft(1).
				PaddingRight(1).
				MarginBottom(1)

	LogDetailsStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			PaddingLeft(3).
			PaddingRight(1).
			PaddingTop(1).
			PaddingBottom(1).
			MarginBottom(1)
)

func CenterHorizontal(width int, content string) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, content)
}

func CenterVertical(height int, content string) string {
	return lipgloss.PlaceVertical(height, lipgloss.Center, content)
}

func Center(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
