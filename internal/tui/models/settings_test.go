package models

import (
	"fmt"
	"testing"

	"github.com/Classic-Homes/gitcells/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSettingsModel_Navigation(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()

	// Test main menu navigation to features
	model.cursor = 3 // Position on "Feature Settings" (now at index 3)
	newModel, _ := model.handleSelectionAndReturn()
	updatedModel := newModel.(SettingsModel)

	if updatedModel.currentView != viewFeatures {
		t.Errorf("Expected currentView to be viewFeatures, got %v", updatedModel.currentView)
	}

	if updatedModel.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0, got %d", updatedModel.cursor)
	}

	if updatedModel.status != "Entered feature settings" {
		t.Errorf("Expected status to be 'Entered feature settings', got %s", updatedModel.status)
	}
}

func TestSettingsModel_FeatureToggle(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()
	model.currentView = viewFeatures
	model.cursor = 0 // Position on "Experimental Features"

	// Get initial value
	initialValue := model.config.Features.EnableExperimentalFeatures

	// Toggle the setting
	newModel, cmd := model.handleSelectionAndReturn()
	updatedModel := newModel.(SettingsModel)

	// Check that the value was toggled
	if updatedModel.config.Features.EnableExperimentalFeatures == initialValue {
		t.Errorf("Expected feature to be toggled from %v", initialValue)
	}

	// Check that a save command was returned
	if cmd == nil {
		t.Error("Expected save command to be returned")
	}
}

func TestSettingsModel_UpdateToggle(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()
	model.currentView = viewUpdates
	model.cursor = 0 // Position on "Auto Check Updates"

	// Get initial value
	initialValue := model.config.Updates.AutoCheckUpdates

	// Toggle the setting
	newModel, cmd := model.handleSelectionAndReturn()
	updatedModel := newModel.(SettingsModel)

	// Check that the value was toggled
	if updatedModel.config.Updates.AutoCheckUpdates == initialValue {
		t.Errorf("Expected setting to be toggled from %v", initialValue)
	}

	// Check that a save command was returned
	if cmd == nil {
		t.Error("Expected save command to be returned")
	}
}

func TestSettingsModel_KeyboardNavigation(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()

	// Test down arrow key
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	updatedModel := newModel.(SettingsModel)

	if updatedModel.cursor != 1 {
		t.Errorf("Expected cursor to move to 1, got %d", updatedModel.cursor)
	}

	// Test up arrow key
	newModel, _ = updatedModel.Update(tea.KeyMsg{Type: tea.KeyUp})
	updatedModel = newModel.(SettingsModel)

	if updatedModel.cursor != 0 {
		t.Errorf("Expected cursor to move back to 0, got %d", updatedModel.cursor)
	}
}

func TestSettingsModel_EscapeNavigation(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()
	model.currentView = viewFeatures
	model.cursor = 2

	// Test escape key to go back to main settings menu
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updatedModel := newModel.(SettingsModel)

	if updatedModel.currentView != viewMain {
		t.Errorf("Expected to return to viewMain, got %v", updatedModel.currentView)
	}

	if updatedModel.cursor != 0 {
		t.Errorf("Expected cursor to be reset to 0, got %d", updatedModel.cursor)
	}

	if updatedModel.status != "Returned to main settings" {
		t.Errorf("Expected status to show return message, got %s", updatedModel.status)
	}
}

func TestSettingsModel_EscapeFromMainSettings(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()
	model.currentView = viewMain
	model.cursor = 1

	// Test escape key from main settings should request going back to main TUI menu
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	updatedModel := newModel.(SettingsModel)

	// Model state should remain the same
	if updatedModel.currentView != viewMain {
		t.Errorf("Expected to remain in viewMain, got %v", updatedModel.currentView)
	}

	// Should return a command to request main menu
	if cmd == nil {
		t.Error("Expected a command to be returned")
	} else {
		// Execute command to get the message
		msg := cmd()
		if fmt.Sprintf("%T", msg) != "messages.RequestMainMenuMsg" {
			t.Errorf("Expected RequestMainMenuMsg, got %T", msg)
		}
	}
}

func TestSettingsModel_ResetToMainView(t *testing.T) {
	model := NewSettingsModel()
	model.config = config.GetDefault()
	model.currentView = viewFeatures
	model.cursor = 3
	model.status = "Some status"
	model.showConfirm = true
	model.updating = true

	// Test reset method
	resetModel := model.ResetToMainView()

	if resetModel.currentView != viewMain {
		t.Errorf("Expected currentView to be viewMain, got %v", resetModel.currentView)
	}

	if resetModel.cursor != 0 {
		t.Errorf("Expected cursor to be 0, got %d", resetModel.cursor)
	}

	if resetModel.status != "Ready" {
		t.Errorf("Expected status to be 'Ready', got %s", resetModel.status)
	}

	if resetModel.showConfirm {
		t.Error("Expected showConfirm to be false")
	}

	if resetModel.updating {
		t.Error("Expected updating to be false")
	}
}
