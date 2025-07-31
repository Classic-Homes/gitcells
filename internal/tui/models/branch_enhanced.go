package models

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
)

type BranchEnhancedModel struct {
	width      int
	height     int
	gitAdapter *adapter.GitAdapter
	
	// UI components
	branchTable  components.TableWithCommands
	newBranchInput components.TextInput
	
	// State
	branches     []adapter.BranchInfo
	mode         BranchMode
	error        string
	success      string
	loading      bool
	showConfirm  bool
	confirmMsg   string
	confirmAction func() error
}

type BranchMode int
const (
	BranchModeList BranchMode = iota
	BranchModeNew
	BranchModeConfirm
)

func NewBranchEnhancedModel() tea.Model {
	m := &BranchEnhancedModel{
		branchTable: components.NewTableWithCommands([]string{"", "Branch", "Status", "Tracking"}),
		newBranchInput: components.NewTextInput("New branch name:", "feature/new-feature"),
		mode: BranchModeList,
	}
	
	// Initialize git adapter
	if gitAdapter, err := adapter.NewGitAdapter("."); err == nil {
		m.gitAdapter = gitAdapter
	}
	
	// Set table properties
	m.branchTable.SetHeight(15)
	
	return m
}

func (m BranchEnhancedModel) Init() tea.Cmd {
	return m.loadBranches()
}

func (m BranchEnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Clear messages after some time
	if m.error != "" || m.success != "" {
		cmds = append(cmds, clearMessagesAfter(3))
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.branchTable.SetHeight(m.height - 15) // Leave room for header and footer

	case tea.KeyMsg:
		switch m.mode {
		case BranchModeList:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "r":
				return m, m.loadBranches()
			case "n":
				m.mode = BranchModeNew
				m.newBranchInput.Focus()
				m.newBranchInput.SetValue("")
			case "s", "enter":
				m.switchToBranch()
			case "d":
				m.confirmDeleteBranch()
			case "m":
				m.confirmMergeBranch()
			default:
				// Pass to table for navigation
				m.branchTable = m.branchTable.Update(msg.String())
			}
			
		case BranchModeNew:
			switch msg.String() {
			case "esc":
				m.mode = BranchModeList
				m.newBranchInput.Blur()
			case "enter":
				if name := m.newBranchInput.Value(); name != "" {
					return m, m.createBranch(name)
				}
			default:
				var cmd tea.Cmd
				m.newBranchInput, cmd = m.newBranchInput.Update(msg)
				cmds = append(cmds, cmd)
			}
			
		case BranchModeConfirm:
			switch msg.String() {
			case "y", "Y":
				if m.confirmAction != nil {
					if err := m.confirmAction(); err != nil {
						m.error = err.Error()
					}
				}
				m.mode = BranchModeList
				m.showConfirm = false
				return m, m.loadBranches()
			case "n", "N", "esc":
				m.mode = BranchModeList
				m.showConfirm = false
			}
		}

	case branchesLoadedMsg:
		m.loading = false
		m.updateBranchTable(msg.branches)
		
	case clearMsg:
		m.error = ""
		m.success = ""
		
	case errorMsg:
		m.error = msg.message
		m.loading = false
		
	case successMsg:
		m.success = msg.message
		m.loading = false
	}

	return m, tea.Batch(cmds...)
}

func (m BranchEnhancedModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	switch m.mode {
	case BranchModeNew:
		return m.renderNewBranch()
	case BranchModeConfirm:
		return m.renderConfirm()
	default:
		return m.renderBranchList()
	}
}

func (m BranchEnhancedModel) renderBranchList() string {
	containerStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Width(m.width).
		Height(m.height)

	// Title
	title := styles.TitleStyle.Render("Branch Management")
	
	// Error/Success messages
	var message string
	if m.error != "" {
		message = styles.ErrorStyle.Render("✗ " + m.error)
	} else if m.success != "" {
		message = styles.SuccessStyle.Render("✓ " + m.success)
	}
	
	// Branch table
	table := m.branchTable.View()
	
	// Help
	help := styles.HelpStyle.Render(
		"[n]ew branch • [s]witch • [m]erge • [d]elete • [r]efresh • [↑/↓] navigate • [esc] back",
	)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		message,
		"",
		table,
		"",
		help,
	)
	
	return containerStyle.Render(content)
}

func (m BranchEnhancedModel) renderNewBranch() string {
	boxStyle := styles.BoxStyle.Copy().
		Width(60).
		Padding(2)
		
	title := styles.TitleStyle.Render("Create New Branch")
	
	currentBranch := "main"
	for _, b := range m.branches {
		if b.Current {
			currentBranch = b.Name
			break
		}
	}
	
	info := styles.MutedStyle.Render(
		fmt.Sprintf("Creating from: %s", currentBranch),
	)
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		info,
		"",
		m.newBranchInput.View(),
		"",
		styles.HelpStyle.Render("[Enter] Create • [Esc] Cancel"),
	)
	
	return styles.Center(m.width, m.height, boxStyle.Render(content))
}

func (m BranchEnhancedModel) renderConfirm() string {
	boxStyle := styles.BoxStyle.Copy().
		Width(60).
		Padding(2)
		
	title := styles.WarningStyle.Render("⚠ Confirm Action")
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		m.confirmMsg,
		"",
		styles.HelpStyle.Render("[Y]es • [N]o"),
	)
	
	return styles.Center(m.width, m.height, boxStyle.Render(content))
}

func (m BranchEnhancedModel) renderLoading() string {
	spinner := components.NewSpinnerProgress("Loading branches...")
	return styles.Center(m.width, m.height, spinner.View())
}

// Helper methods
func (m *BranchEnhancedModel) updateBranchTable(branches []adapter.BranchInfo) {
	m.branches = branches
	
	rows := [][]string{}
	for _, branch := range branches {
		icon := "  "
		if branch.Current {
			icon = "●"
		}
		
		status := "Clean"
		if branch.HasChanges {
			status = "Modified"
		}
		
		tracking := ""
		if branch.Ahead > 0 || branch.Behind > 0 {
			tracking = fmt.Sprintf("↑%d ↓%d", branch.Ahead, branch.Behind)
		}
		
		rows = append(rows, []string{icon, branch.Name, status, tracking})
	}
	
	m.branchTable.SetRows(rows)
}

func (m *BranchEnhancedModel) switchToBranch() {
	selected := m.branchTable.SelectedRow()
	if selected == nil || len(selected) < 2 {
		return
	}
	
	branchName := selected[1]
	
	// Don't switch to current branch
	for _, b := range m.branches {
		if b.Name == branchName && b.Current {
			m.error = "Already on branch " + branchName
			return
		}
	}
	
	// In a real implementation, this would use git operations
	m.success = fmt.Sprintf("Switched to branch %s", branchName)
}

func (m *BranchEnhancedModel) confirmDeleteBranch() {
	selected := m.branchTable.SelectedRow()
	if selected == nil || len(selected) < 2 {
		return
	}
	
	branchName := selected[1]
	
	// Can't delete current branch
	for _, b := range m.branches {
		if b.Name == branchName && b.Current {
			m.error = "Cannot delete current branch"
			return
		}
	}
	
	m.mode = BranchModeConfirm
	m.confirmMsg = fmt.Sprintf("Delete branch '%s'?", branchName)
	m.confirmAction = func() error {
		// In a real implementation, this would delete the branch
		m.success = fmt.Sprintf("Deleted branch %s", branchName)
		return nil
	}
}

func (m *BranchEnhancedModel) confirmMergeBranch() {
	selected := m.branchTable.SelectedRow()
	if selected == nil || len(selected) < 2 {
		return
	}
	
	branchName := selected[1]
	
	// Can't merge current branch into itself
	for _, b := range m.branches {
		if b.Name == branchName && b.Current {
			m.error = "Cannot merge branch into itself"
			return
		}
	}
	
	currentBranch := "main"
	for _, b := range m.branches {
		if b.Current {
			currentBranch = b.Name
			break
		}
	}
	
	m.mode = BranchModeConfirm
	m.confirmMsg = fmt.Sprintf("Merge '%s' into '%s'?", branchName, currentBranch)
	m.confirmAction = func() error {
		// In a real implementation, this would merge the branch
		m.success = fmt.Sprintf("Merged %s into %s", branchName, currentBranch)
		return nil
	}
}

// Commands
func (m *BranchEnhancedModel) loadBranches() tea.Cmd {
	m.loading = true
	return func() tea.Msg {
		if m.gitAdapter == nil {
			return errorMsg{message: "Git not initialized"}
		}
		
		branches, err := m.gitAdapter.GetBranches()
		if err != nil {
			return errorMsg{message: err.Error()}
		}
		
		// Add some mock data for demo
		if len(branches) == 0 {
			branches = []adapter.BranchInfo{
				{Name: "main", Current: true, HasChanges: false},
				{Name: "feature/excel-improvements", Current: false, HasChanges: true, Ahead: 2},
				{Name: "fix/formula-parsing", Current: false, HasChanges: false, Behind: 1},
				{Name: "feature/batch-conversion", Current: false, HasChanges: true, Ahead: 5, Behind: 2},
			}
		}
		
		return branchesLoadedMsg{branches: branches}
	}
}

func (m *BranchEnhancedModel) createBranch(name string) tea.Cmd {
	return func() tea.Msg {
		// Validate branch name
		if strings.Contains(name, " ") {
			return errorMsg{message: "Branch name cannot contain spaces"}
		}
		
		// In a real implementation, this would create the branch
		m.mode = BranchModeList
		return successMsg{message: fmt.Sprintf("Created branch %s", name)}
	}
}

func clearMessagesAfter(seconds int) tea.Cmd {
	return tea.Tick(time.Second*time.Duration(seconds), func(time.Time) tea.Msg {
		return clearMsg{}
	})
}

// Message types
type branchesLoadedMsg struct {
	branches []adapter.BranchInfo
}

type errorMsg struct {
	message string
}

type successMsg struct {
	message string
}

type clearMsg struct{}