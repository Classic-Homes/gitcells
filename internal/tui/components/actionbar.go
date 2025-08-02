package components

import (
	"strings"

	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// ActionBarItem represents a single action in the action bar
type ActionBarItem struct {
	Key         string
	Label       string
	Active      bool
	Highlighted bool
}

// RenderActionBar creates a consistent action bar for all screens
func RenderActionBar(width int, actions []ActionBarItem, helpText string) string {
	actionStrs := make([]string, 0, len(actions))

	for _, action := range actions {
		var style lipgloss.Style
		switch {
		case action.Highlighted:
			style = styles.ActiveButtonStyle
		case action.Active:
			style = styles.ActionStyle
		default:
			style = styles.MutedStyle
		}

		actionStr := style.Render("[" + action.Key + "] " + action.Label)
		actionStrs = append(actionStrs, actionStr)
	}

	actionsJoined := strings.Join(actionStrs, " • ")

	// Calculate padding
	actionsWidth := lipgloss.Width(actionsJoined)
	helpWidth := lipgloss.Width(helpText)
	padding := width - actionsWidth - helpWidth - 4
	if padding < 0 {
		padding = 1
	}

	actionBar := lipgloss.JoinHorizontal(
		lipgloss.Left,
		actionsJoined,
		strings.Repeat(" ", padding),
		styles.MutedStyle.Render(helpText),
	)

	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("241")).
		Width(width).
		Padding(0, 2).
		Render(actionBar)
}

// RenderBreadcrumb creates a breadcrumb navigation
func RenderBreadcrumb(items []string) string {
	var parts []string
	for i, item := range items {
		if i == len(items)-1 {
			// Current page in primary color
			titleStyle := styles.TitleStyle
			parts = append(parts, titleStyle.MarginBottom(0).Render(item))
		} else {
			// Previous pages in muted color
			parts = append(parts, styles.MutedStyle.Render(item))
		}
	}
	return strings.Join(parts, styles.MutedStyle.Render(" › "))
}

// RenderHeader creates a consistent header with title and breadcrumb
func RenderHeader(title string, breadcrumb []string, width int) string {
	// Breadcrumb
	bc := RenderBreadcrumb(breadcrumb)

	// Title
	titleStr := styles.TitleStyle.Render(title)

	// Combine
	header := lipgloss.JoinVertical(
		lipgloss.Left,
		bc,
		titleStr,
	)

	// Add bottom border
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		styles.BorderStyle.Render(strings.Repeat("─", width)),
	)
}
