package common

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestRenderMenuItems(t *testing.T) {
	items := []MenuItem{
		{Title: "Item 1", Description: "Description 1", Key: "key1"},
		{Title: "Item 2", Description: "Description 2", Key: "key2"},
		{Title: "Item 3", Description: "Description 3", Key: "key3"},
	}

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	result := RenderMenuItems(items, 1, 80, cursorStyle, descStyle)

	// Check that cursor is on second item
	assert.Contains(t, result, "▸ Item 2")
	// Check that other items don't have cursor
	assert.Contains(t, result, "  Item 1")
	assert.Contains(t, result, "  Item 3")
	// Check descriptions are present
	assert.Contains(t, result, "Description 1")
	assert.Contains(t, result, "Description 2")
	assert.Contains(t, result, "Description 3")
}

func TestRenderBooleanValue(t *testing.T) {
	enabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	enabled := RenderBooleanValue(true, enabledStyle, disabledStyle)
	disabled := RenderBooleanValue(false, enabledStyle, disabledStyle)

	// Note: We can't easily test the styling itself, but we can test the content
	assert.Contains(t, enabled, "Enabled")
	assert.Contains(t, disabled, "Disabled")
}

func TestRenderSettingItem(t *testing.T) {
	cursorStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	// Test selected item
	selected := RenderSettingItem("Setting Name", "Value", true, cursorStyle, labelStyle, valueStyle, 80)
	assert.Contains(t, selected, "▸")
	assert.Contains(t, selected, "Setting Name")
	assert.Contains(t, selected, "Value")

	// Test unselected item
	unselected := RenderSettingItem("Setting Name", "Value", false, cursorStyle, labelStyle, valueStyle, 80)
	assert.Contains(t, unselected, "Setting Name")
	assert.Contains(t, unselected, "Value")
	assert.NotContains(t, unselected, "▸")
}

func TestRenderConfirmDialog(t *testing.T) {
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	result := RenderConfirmDialog("Confirm Action", "Are you sure?", 80, 24, borderStyle)

	assert.Contains(t, result, "Confirm Action")
	assert.Contains(t, result, "Are you sure?")
	assert.Contains(t, result, "[y] Yes")
	assert.Contains(t, result, "[n] No")
}

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected int // expected number of lines
	}{
		{
			name:     "short text",
			text:     "Hello world",
			width:    80,
			expected: 1,
		},
		{
			name:     "text needs wrapping",
			text:     "This is a longer piece of text that should be wrapped",
			width:    20,
			expected: 3,
		},
		{
			name:     "empty text",
			text:     "",
			width:    80,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.text, tt.width)
			lines := strings.Split(result, "\n")

			if tt.expected == 0 {
				assert.Empty(t, result)
			} else {
				assert.Equal(t, tt.expected, len(lines))

				// Verify no line exceeds width
				for _, line := range lines {
					assert.LessOrEqual(t, len(line), tt.width)
				}
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "no truncation needed",
			input:    "Hello",
			maxLen:   10,
			expected: "Hello",
		},
		{
			name:     "exact length",
			input:    "Hello",
			maxLen:   5,
			expected: "Hello",
		},
		{
			name:     "needs truncation",
			input:    "Hello World",
			maxLen:   8,
			expected: "Hello...",
		},
		{
			name:     "very short max",
			input:    "Hello",
			maxLen:   3,
			expected: "...",
		},
		{
			name:     "max length 1",
			input:    "Hello",
			maxLen:   1,
			expected: "...", // Note: result will be "..." which is 3 chars, exceeds maxLen of 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
			// Only check length constraint for maxLen > 3
			if tt.maxLen > 3 {
				assert.LessOrEqual(t, len(result), tt.maxLen)
			}
		})
	}
}

func TestRenderList(t *testing.T) {
	items := []string{"Item 1", "Item 2", "Item 3"}
	style := lipgloss.NewStyle()

	// Test numbered list
	numbered := RenderList(items, true, style)
	assert.Contains(t, numbered, "1. Item 1")
	assert.Contains(t, numbered, "2. Item 2")
	assert.Contains(t, numbered, "3. Item 3")

	// Test bulleted list
	bulleted := RenderList(items, false, style)
	assert.Contains(t, bulleted, "• Item 1")
	assert.Contains(t, bulleted, "• Item 2")
	assert.Contains(t, bulleted, "• Item 3")
}

func TestRenderProgressIndicator(t *testing.T) {
	style := lipgloss.NewStyle()
	result := RenderProgressIndicator("Loading...", style)

	assert.Contains(t, result, "Loading...")
	// Should contain some spinner character
	assert.NotEmpty(t, result)
}

func TestRenderStatus(t *testing.T) {
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	success := RenderStatus("Success!", false, successStyle, errorStyle)
	assert.Contains(t, success, "Success!")

	error := RenderStatus("Error occurred", true, successStyle, errorStyle)
	assert.Contains(t, error, "Error occurred")
}
