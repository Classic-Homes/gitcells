package models

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestUnifiedNavigation tests that all models respond to escape key with RequestMainMenuMsg
func TestUnifiedNavigation(t *testing.T) {
	tests := []struct {
		name        string
		model       tea.Model
		setupFunc   func(tea.Model) tea.Model
		description string
	}{
		{
			name:        "DashboardModel",
			model:       NewDashboardModel(),
			description: "Dashboard should return to main menu on escape",
		},
		{
			name:        "SetupModel",
			model:       NewSetupModel(),
			description: "Setup should return to main menu on escape",
		},
		{
			name:        "ErrorLogModel",
			model:       NewErrorLogModel(),
			description: "Error Log should return to main menu on escape",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := tt.model
			if tt.setupFunc != nil {
				model = tt.setupFunc(model)
			}

			// Send escape key
			_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEscape})

			if cmd == nil {
				t.Errorf("%s: Expected a command to be returned on escape key", tt.name)
			} else {
				// Execute command to get the message
				msg := cmd()
				msgType := fmt.Sprintf("%T", msg)
				if msgType != "messages.RequestMainMenuMsg" {
					t.Errorf("%s: Expected RequestMainMenuMsg, got %s", tt.name, msgType)
				}
			}
		})
	}
}

// TestQuitBehavior tests that Ctrl+C and Q still quit the application
func TestQuitBehavior(t *testing.T) {
	models := []struct {
		name  string
		model tea.Model
	}{
		{"DashboardModel", NewDashboardModel()},
		{"SetupModel", NewSetupModel()},
		{"ErrorLogModel", NewErrorLogModel()},
		{"SettingsModel", NewSettingsModel()},
	}

	for _, m := range models {
		t.Run(m.name+"_CtrlC", func(t *testing.T) {
			_, cmd := m.model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
			if cmd == nil {
				t.Errorf("%s: Expected quit command on Ctrl+C", m.name)
			}
		})

		t.Run(m.name+"_Q", func(t *testing.T) {
			_, cmd := m.model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
			if cmd == nil {
				t.Errorf("%s: Expected quit command on 'q' key", m.name)
			}
		})
	}
}
