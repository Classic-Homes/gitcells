package models

import (
	"fmt"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewDashboardModel(t *testing.T) {
	t.Run("creates new dashboard model", func(t *testing.T) {
		model := NewDashboardModel()
		assert.NotNil(t, model)

		dashModel, ok := model.(*DashboardModel)
		assert.True(t, ok)
		assert.NotNil(t, dashModel.operations)
		assert.NotNil(t, dashModel.progressBars)
		assert.False(t, dashModel.lastUpdate.IsZero())
		assert.Equal(t, 0, dashModel.selectedTab)
		assert.Equal(t, 0, dashModel.scrollOffset)
		assert.False(t, dashModel.showHelp)
	})
}

func TestDashboardModel_Init(t *testing.T) {
	t.Run("returns batch command", func(t *testing.T) {
		model := NewDashboardModel().(*DashboardModel)
		cmd := model.Init()
		assert.NotNil(t, cmd)
	})
}

func TestDashboardModel_Update(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)

	t.Run("handles window size message", func(t *testing.T) {
		msg := tea.WindowSizeMsg{Width: 100, Height: 50}
		updatedModel, cmd := model.Update(msg)

		dashModel := updatedModel.(DashboardModel)
		assert.Equal(t, 100, dashModel.width)
		assert.Equal(t, 50, dashModel.height)
		// cmd might be nil or a batch command
		_ = cmd
	})

	t.Run("handles quit keys", func(t *testing.T) {
		testCases := []string{"ctrl+c", "q"}
		for _, key := range testCases {
			t.Run(fmt.Sprintf("quit with %s", key), func(t *testing.T) {
				msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
				if key == "ctrl+c" {
					msg = tea.KeyMsg{Type: tea.KeyCtrlC}
				}

				_, cmd := model.Update(msg)
				assert.NotNil(t, cmd)
			})
		}
	})

	t.Run("handles tab switching", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyTab}
		updatedModel, _ := model.Update(msg)

		dashModel := updatedModel.(DashboardModel)
		assert.Equal(t, 1, dashModel.selectedTab)

		// Test wrapping
		dashModel.selectedTab = 2
		updatedModel, _ = dashModel.Update(msg)
		dashModel = updatedModel.(DashboardModel)
		assert.Equal(t, 0, dashModel.selectedTab)
	})

	t.Run("handles help toggle", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
		updatedModel, _ := model.Update(msg)

		dashModel := updatedModel.(DashboardModel)
		assert.True(t, dashModel.showHelp)

		// Toggle again
		updatedModel, _ = dashModel.Update(msg)
		dashModel = updatedModel.(DashboardModel)
		assert.False(t, dashModel.showHelp)
	})

	t.Run("handles scroll navigation", func(t *testing.T) {
		// Reset model to start fresh
		testModel := NewDashboardModel().(*DashboardModel)

		// Test down key
		msg := tea.KeyMsg{Type: tea.KeyDown}
		updatedModel, _ := testModel.Update(msg)
		dashModel := updatedModel.(DashboardModel)
		assert.Equal(t, 1, dashModel.scrollOffset)

		// Test up key - should decrease but not go below 0
		msg = tea.KeyMsg{Type: tea.KeyUp}
		updatedModel, _ = dashModel.Update(msg)
		dashModel = updatedModel.(DashboardModel)
		assert.Equal(t, 0, dashModel.scrollOffset)

		// Test up key again - should stay at 0
		updatedModel, _ = dashModel.Update(msg)
		dashModel = updatedModel.(DashboardModel)
		assert.Equal(t, 0, dashModel.scrollOffset)
	})

	t.Run("handles dashboard tick message", func(t *testing.T) {
		// Add a mock operation in progress
		model.operations = []FileOperation{
			{
				ID:        "test-op",
				Type:      OpConvert,
				FileName:  "test.xlsx",
				Status:    StatusInProgress,
				Progress:  50,
				StartTime: time.Now(),
			},
		}

		msg := dashboardTickMsg(time.Now())
		updatedModel, cmd := model.Update(msg)

		dashModel := updatedModel.(DashboardModel)
		assert.NotNil(t, cmd)
		assert.Equal(t, 60, dashModel.operations[0].Progress) // Should increase by 10
	})

	t.Run("handles data loaded message", func(t *testing.T) {
		msg := dataLoadedMsg{
			totalFiles: 5,
			branch:     "main",
			hasChanges: true,
		}

		updatedModel, _ := model.Update(msg)
		dashModel := updatedModel.(DashboardModel)

		assert.Equal(t, 5, dashModel.totalFiles)
		assert.Equal(t, "main", dashModel.syncStatus.Branch)
		assert.True(t, dashModel.syncStatus.HasChanges)
		assert.False(t, dashModel.syncStatus.IsSynced)
	})

	t.Run("handles operation update message", func(t *testing.T) {
		// Use a fresh model to avoid conflicts with previous tests
		freshModel := NewDashboardModel().(*DashboardModel)

		op := FileOperation{
			ID:        "new-op",
			Type:      OpSync,
			FileName:  "sync.xlsx",
			Status:    StatusCompleted,
			Progress:  100,
			StartTime: time.Now(),
		}

		msg := operationUpdateMsg{operation: op}
		updatedModel, _ := freshModel.Update(msg)

		dashModel := updatedModel.(DashboardModel)
		assert.Len(t, dashModel.operations, 1)
		assert.Equal(t, "new-op", dashModel.operations[0].ID)
		assert.Equal(t, OpSync, dashModel.operations[0].Type)
		assert.Equal(t, StatusCompleted, dashModel.operations[0].Status)
	})
}

func TestDashboardModel_View(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)
	model.width = 100
	model.height = 50

	t.Run("renders main view", func(t *testing.T) {
		view := model.View()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "GitCells Dashboard")
	})

	t.Run("renders help when showHelp is true", func(t *testing.T) {
		model.showHelp = true
		view := model.View()
		assert.NotEmpty(t, view)
		assert.Contains(t, view, "GitCells Dashboard Help")
		assert.Contains(t, view, "Navigation:")
		assert.Contains(t, view, "Actions:")
	})

	t.Run("renders different tabs", func(t *testing.T) {
		model.showHelp = false

		// Test Overview tab
		model.selectedTab = 0
		view := model.View()
		assert.Contains(t, view, "Watching")

		// Test Operations tab
		model.selectedTab = 1
		view = model.View()
		assert.NotEmpty(t, view)

		// Test Commits tab
		model.selectedTab = 2
		view = model.View()
		assert.NotEmpty(t, view)
	})
}

func TestDashboardModel_RenderMethods(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)
	model.width = 100
	model.height = 50
	model.totalFiles = 5

	t.Run("renderHeader", func(t *testing.T) {
		header := model.renderHeader()
		assert.NotEmpty(t, header)
		assert.Contains(t, header, "GitCells Dashboard")
		assert.Contains(t, header, "Tracking: 5 files")
	})

	t.Run("renderTabs", func(t *testing.T) {
		tabs := model.renderTabs()
		assert.NotEmpty(t, tabs)
		assert.Contains(t, tabs, "Overview")
		assert.Contains(t, tabs, "Operations")
		assert.Contains(t, tabs, "Commits")
	})

	t.Run("renderOverview", func(t *testing.T) {
		overview := model.renderOverview()
		assert.NotEmpty(t, overview)
		assert.Contains(t, overview, "Watching")
		assert.Contains(t, overview, "Auto-Sync")
		assert.Contains(t, overview, "Excel Files")
	})

	t.Run("renderOperations with no operations", func(t *testing.T) {
		operations := model.renderOperations()
		assert.Contains(t, operations, "No operations in progress")
	})

	t.Run("renderOperations with operations", func(t *testing.T) {
		model.operations = []FileOperation{
			{
				ID:        "test-op",
				Type:      OpConvert,
				FileName:  "test.xlsx",
				Status:    StatusInProgress,
				Progress:  75,
				StartTime: time.Now(),
			},
			{
				ID:        "completed-op",
				Type:      OpSync,
				FileName:  "completed.xlsx",
				Status:    StatusCompleted,
				Progress:  100,
				StartTime: time.Now().Add(-2 * time.Minute),
			},
		}

		operations := model.renderOperations()
		assert.NotEmpty(t, operations)
		assert.Contains(t, operations, "test.xlsx")
		assert.Contains(t, operations, "Completed Operations")
	})

	t.Run("renderCommits with no commits", func(t *testing.T) {
		commits := model.renderCommits()
		assert.Contains(t, commits, "No recent commits")
	})

	t.Run("renderCommits with commits", func(t *testing.T) {
		model.recentCommits = []CommitInfo{
			{
				Hash:    "abcd1234567890",
				Message: "Update Excel files",
				Time:    time.Now().Add(-1 * time.Hour),
				Files:   3,
			},
		}

		commits := model.renderCommits()
		assert.NotEmpty(t, commits)
		assert.Contains(t, commits, "abcd123")
		assert.Contains(t, commits, "Update Excel files")
		assert.Contains(t, commits, "(3 files)")
	})

	t.Run("renderFooter", func(t *testing.T) {
		footer := model.renderFooter()
		assert.NotEmpty(t, footer)
		assert.Contains(t, footer, "[Tab] Switch tabs")
		assert.Contains(t, footer, "[q] Quit")
	})
}

func TestDashboardModel_HelperMethods(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)

	t.Run("getOperationIcon", func(t *testing.T) {
		testCases := []struct {
			opType   OperationType
			expected string
		}{
			{OpConvert, "🔄"},
			{OpSync, "🔄"},
			{OpWatch, "👁️"},
			{OperationType(999), "•"}, // Unknown type
		}

		for _, tc := range testCases {
			op := FileOperation{Type: tc.opType}
			icon := model.getOperationIcon(op)
			assert.Equal(t, tc.expected, icon)
		}
	})

	t.Run("getOperationStatus", func(t *testing.T) {
		testCases := []struct {
			status   OperationStatus
			progress int
			error    error
			expected string
		}{
			{StatusPending, 0, nil, "Pending"},
			{StatusInProgress, 50, nil, "In Progress (50%)"},
			{StatusCompleted, 100, nil, "Completed"},
			{StatusFailed, 0, fmt.Errorf("test error"), "Failed: test error"},
			{StatusFailed, 0, nil, "Failed"},
			{OperationStatus(999), 0, nil, "Unknown"}, // Unknown status
		}

		for _, tc := range testCases {
			op := FileOperation{
				Status:   tc.status,
				Progress: tc.progress,
				Error:    tc.error,
			}
			status := model.getOperationStatus(op)
			assert.Equal(t, tc.expected, status)
		}
	})
}

func TestDashboardModel_CommandMethods(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)

	t.Run("loadInitialData returns command", func(t *testing.T) {
		cmd := model.loadInitialData()
		assert.NotNil(t, cmd)

		// Execute the command to get the message
		msg := cmd()
		dataMsg, ok := msg.(dataLoadedMsg)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, dataMsg.totalFiles, 0)
	})

	t.Run("refreshData returns command", func(t *testing.T) {
		cmd := model.refreshData()
		assert.NotNil(t, cmd)
	})

	t.Run("startConversion returns command", func(t *testing.T) {
		cmd := model.startConversion()
		assert.NotNil(t, cmd)
	})

	t.Run("toggleWatcher returns command", func(t *testing.T) {
		cmd := model.toggleWatcher()
		assert.NotNil(t, cmd)

		// Execute the command
		msg := cmd()
		opMsg, ok := msg.(operationUpdateMsg)
		assert.True(t, ok)
		assert.Equal(t, OpWatch, opMsg.operation.Type)
		assert.Equal(t, "File watching", opMsg.operation.FileName)
	})
}

func TestDashboardModel_UpdateOperation(t *testing.T) {
	model := NewDashboardModel().(*DashboardModel)

	t.Run("adds new operation", func(t *testing.T) {
		op := FileOperation{
			ID:        "new-op",
			Type:      OpConvert,
			FileName:  "test.xlsx",
			Status:    StatusInProgress,
			Progress:  0,
			StartTime: time.Now(),
		}

		msg := operationUpdateMsg{operation: op}
		model.updateOperation(msg)

		assert.Len(t, model.operations, 1)
		assert.Equal(t, "new-op", model.operations[0].ID)
	})

	t.Run("updates existing operation", func(t *testing.T) {
		// First add an operation
		op := FileOperation{
			ID:        "existing-op",
			Type:      OpConvert,
			FileName:  "test.xlsx",
			Status:    StatusInProgress,
			Progress:  50,
			StartTime: time.Now(),
		}
		model.operations = []FileOperation{op}

		// Update the operation
		updatedOp := op
		updatedOp.Progress = 100
		updatedOp.Status = StatusCompleted

		msg := operationUpdateMsg{operation: updatedOp}
		model.updateOperation(msg)

		assert.Len(t, model.operations, 1)
		assert.Equal(t, 100, model.operations[0].Progress)
		assert.Equal(t, StatusCompleted, model.operations[0].Status)
	})
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30 sec"},
		{45 * time.Second, "45 sec"},
		{1 * time.Minute, "1 min"},
		{2*time.Minute + 30*time.Second, "2 min"},
		{1 * time.Hour, "60 min"},
	}

	for _, tc := range testCases {
		result := formatDuration(tc.duration)
		assert.Equal(t, tc.expected, result)
	}
}

func TestDashboardTickMsg(t *testing.T) {
	t.Run("dashboardTick returns command", func(t *testing.T) {
		cmd := dashboardTick()
		assert.NotNil(t, cmd)
	})
}

func TestDataStructures(t *testing.T) {
	t.Run("SyncStatus", func(t *testing.T) {
		status := SyncStatus{
			Branch:       "main",
			IsSynced:     true,
			LastCommit:   time.Now(),
			HasChanges:   false,
			RemoteAhead:  0,
			RemoteBehind: 1,
		}

		assert.Equal(t, "main", status.Branch)
		assert.True(t, status.IsSynced)
		assert.False(t, status.HasChanges)
		assert.Equal(t, 0, status.RemoteAhead)
		assert.Equal(t, 1, status.RemoteBehind)
	})

	t.Run("FileOperation", func(t *testing.T) {
		startTime := time.Now()
		op := FileOperation{
			ID:        "test-id",
			Type:      OpConvert,
			FileName:  "test.xlsx",
			Status:    StatusInProgress,
			Progress:  75,
			StartTime: startTime,
			Error:     fmt.Errorf("test error"),
		}

		assert.Equal(t, "test-id", op.ID)
		assert.Equal(t, OpConvert, op.Type)
		assert.Equal(t, "test.xlsx", op.FileName)
		assert.Equal(t, StatusInProgress, op.Status)
		assert.Equal(t, 75, op.Progress)
		assert.Equal(t, startTime, op.StartTime)
		assert.EqualError(t, op.Error, "test error")
	})

	t.Run("CommitInfo", func(t *testing.T) {
		commitTime := time.Now()
		commit := CommitInfo{
			Hash:    "abcd1234",
			Message: "Test commit",
			Time:    commitTime,
			Files:   3,
		}

		assert.Equal(t, "abcd1234", commit.Hash)
		assert.Equal(t, "Test commit", commit.Message)
		assert.Equal(t, commitTime, commit.Time)
		assert.Equal(t, 3, commit.Files)
	})
}

func TestOperationTypeConstants(t *testing.T) {
	assert.Equal(t, 0, int(OpConvert))
	assert.Equal(t, 1, int(OpSync))
	assert.Equal(t, 2, int(OpWatch))
}

func TestOperationStatusConstants(t *testing.T) {
	assert.Equal(t, 0, int(StatusPending))
	assert.Equal(t, 1, int(StatusInProgress))
	assert.Equal(t, 2, int(StatusCompleted))
	assert.Equal(t, 3, int(StatusFailed))
}
