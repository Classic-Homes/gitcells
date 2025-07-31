package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary   = lipgloss.Color("99")
	Secondary = lipgloss.Color("214")
	Success   = lipgloss.Color("82")
	Warning   = lipgloss.Color("214")
	Error     = lipgloss.Color("196")
	Muted     = lipgloss.Color("241")

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
