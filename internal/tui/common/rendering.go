package common

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// MenuItem represents a selectable menu item
type MenuItem struct {
	Title       string
	Description string
	Key         string
}

// RenderMenuItems renders a list of menu items with cursor highlighting
func RenderMenuItems(items []MenuItem, cursor int, width int, cursorStyle, descStyle lipgloss.Style) string {
	var b strings.Builder

	for i, item := range items {
		title := item.Title
		desc := item.Description

		if i == cursor {
			title = cursorStyle.Render("▸ " + title)
			if desc != "" {
				desc = descStyle.Render(fmt.Sprintf("  %s", desc))
			}
		} else {
			title = "  " + title
			if desc != "" {
				desc = descStyle.Render(fmt.Sprintf("  %s", desc))
			}
		}

		b.WriteString(title + "\n")
		if desc != "" {
			b.WriteString(desc + "\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

// RenderBooleanValue renders a boolean value as "Enabled" or "Disabled" with color
func RenderBooleanValue(value bool, enabledStyle, disabledStyle lipgloss.Style) string {
	if value {
		return enabledStyle.Render("Enabled")
	}
	return disabledStyle.Render("Disabled")
}

// RenderSettingItem renders a single setting item with label and value
func RenderSettingItem(label, value string, selected bool, cursorStyle, labelStyle, valueStyle lipgloss.Style, width int) string {
	cursor := "  "
	if selected {
		cursor = cursorStyle.Render("▸ ")
		label = cursorStyle.Render(label)
	} else {
		label = labelStyle.Render(label)
	}

	value = valueStyle.Render(value)

	// Calculate spacing to align values
	labelWidth := lipgloss.Width(label) + 2 // +2 for cursor
	valueWidth := lipgloss.Width(value)
	spacing := width - labelWidth - valueWidth - 4 // -4 for margins

	if spacing < 2 {
		spacing = 2
	}

	return fmt.Sprintf("%s%s%s%s", cursor, label, strings.Repeat(" ", spacing), value)
}

// RenderHeader renders a styled header
func RenderHeader(title string, style lipgloss.Style, width int) string {
	return style.Width(width).Render(title)
}

// RenderFooter renders a footer with help text
func RenderFooter(helpText string, style lipgloss.Style, width int) string {
	return style.Width(width).Render(helpText)
}

// RenderStatus renders a status message with appropriate styling
func RenderStatus(status string, isError bool, successStyle, errorStyle lipgloss.Style) string {
	if isError {
		return errorStyle.Render(status)
	}
	return successStyle.Render(status)
}

// RenderConfirmDialog renders a confirmation dialog box
func RenderConfirmDialog(title, message string, width, height int, borderStyle lipgloss.Style) string {
	content := fmt.Sprintf("%s\n\n%s\n\n[y] Yes  [n] No", title, message)

	box := borderStyle.
		Width(width-4).
		Height(height-4).
		Padding(1, 2).
		Render(content)

	// Center the box
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// RenderProgressIndicator renders a simple progress indicator
func RenderProgressIndicator(message string, style lipgloss.Style) string {
	spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
	// For static rendering, just use first frame
	return style.Render(fmt.Sprintf("%c %s", spinner[0], message))
}

// WrapText wraps text to fit within a specified width
func WrapText(text string, width int) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	for _, word := range words {
		wordLen := len(word)

		// If adding this word would exceed width, start new line
		if currentWidth > 0 && currentWidth+1+wordLen > width {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}

		// Add word to current line
		if currentWidth > 0 {
			currentLine.WriteString(" ")
			currentWidth++
		}
		currentLine.WriteString(word)
		currentWidth += wordLen
	}

	// Add final line
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// TruncateString truncates a string to the specified length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// RenderList renders a numbered or bulleted list
func RenderList(items []string, numbered bool, style lipgloss.Style) string {
	var b strings.Builder

	for i, item := range items {
		var prefix string
		if numbered {
			prefix = fmt.Sprintf("%d. ", i+1)
		} else {
			prefix = "• "
		}

		b.WriteString(style.Render(prefix + item))
		if i < len(items)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
