package models

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/Classic-Homes/gitcells/internal/tui/adapter"
	"github.com/Classic-Homes/gitcells/internal/tui/components"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	"github.com/Classic-Homes/gitcells/internal/tui/types"
)

type SetupEnhancedModel struct {
	width    int
	height   int
	step     int
	focused  int
	finished bool
	error    string
	
	// Step 1 - Directory
	dirInput components.TextInput
	
	// Step 2 - Excel patterns
	patternInput components.TextInput
	
	// Step 3 - Git settings
	autoCommit     components.Checkbox
	autoPush       components.Checkbox
	commitTemplate components.TextInput
	
	// Configuration data
	config types.SetupConfig
}

func NewSetupEnhancedModel() SetupEnhancedModel {
	// Get current directory
	cwd, _ := os.Getwd()
	
	m := SetupEnhancedModel{
		dirInput:       components.NewTextInput("Repository Directory:", cwd),
		patternInput:   components.NewTextInput("Excel File Pattern:", "*.xlsx"),
		autoCommit:     components.NewCheckbox("Enable auto-commit", true),
		autoPush:       components.NewCheckbox("Enable auto-push", false),
		commitTemplate: components.NewTextInput("Commit Message Template:", "GitCells: {action} {filename}"),
	}
	
	// Set initial values
	m.dirInput.SetValue(cwd)
	m.patternInput.SetValue("*.xlsx")
	m.commitTemplate.SetValue("GitCells: {action} {filename}")
	
	// Focus first input
	m.dirInput.Focus()
	
	return m
}

func (m SetupEnhancedModel) Init() tea.Cmd {
	return nil
}

func (m SetupEnhancedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
			
		case "tab", "down":
			m.focusNext()
			
		case "shift+tab", "up":
			m.focusPrev()
			
		case "enter":
			if m.canProceed() {
				if m.step == 3 { // Last step
					m.saveConfig()
					m.finished = true
				} else {
					m.step++
					m.focused = 0
					m.focusCurrentInput()
				}
			}
			
		case "left", "h":
			if m.step > 0 {
				m.step--
				m.focused = 0
				m.focusCurrentInput()
			}
		}
	}

	// Update current inputs based on step
	var cmd tea.Cmd
	switch m.step {
	case 0:
		m.dirInput, cmd = m.dirInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case 1:
		m.patternInput, cmd = m.patternInput.Update(msg)
		cmds = append(cmds, cmd)
		
	case 2:
		switch m.focused {
		case 0:
			m.autoCommit, cmd = m.autoCommit.Update(msg)
		case 1:
			m.autoPush, cmd = m.autoPush.Update(msg)
		case 2:
			m.commitTemplate, cmd = m.commitTemplate.Update(msg)
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m SetupEnhancedModel) View() string {
	if m.finished {
		return m.renderComplete()
	}

	containerStyle := lipgloss.NewStyle().
		Padding(2, 4).
		Width(80)

	stepIndicator := m.renderStepIndicator()
	
	var content string
	switch m.step {
	case 0:
		content = m.renderDirectoryStep()
	case 1:
		content = m.renderPatternStep()
	case 2:
		content = m.renderGitStep()
	case 3:
		content = m.renderReviewStep()
	}

	help := styles.HelpStyle.Render("Tab/↓: Next field • Shift+Tab/↑: Previous field • Enter: Continue • ←/h: Back • Esc: Cancel")

	return containerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			stepIndicator,
			"",
			content,
			"",
			help,
		),
	)
}

func (m SetupEnhancedModel) renderStepIndicator() string {
	steps := []string{"Directory", "Patterns", "Git Settings", "Review"}
	
	var indicators []string
	for i, step := range steps {
		style := lipgloss.NewStyle().Foreground(styles.Muted)
		if i == m.step {
			style = style.Foreground(styles.Primary).Bold(true)
		} else if i < m.step {
			style = style.Foreground(styles.Success)
		}
		
		num := fmt.Sprintf("%d", i+1)
		if i < m.step {
			num = "✓"
		}
		
		indicators = append(indicators, style.Render(fmt.Sprintf("%s %s", num, step)))
	}
	
	return lipgloss.JoinHorizontal(lipgloss.Top, indicators[0], " → ", indicators[1], " → ", indicators[2], " → ", indicators[3])
}

func (m SetupEnhancedModel) renderDirectoryStep() string {
	title := styles.TitleStyle.Render("Select Repository Directory")
	desc := styles.MutedStyle.Render("Choose the directory containing your Excel files")
	
	// Check if directory exists
	info := ""
	if dir := m.dirInput.Value(); dir != "" {
		if stat, err := os.Stat(dir); err == nil && stat.IsDir() {
			info = styles.SuccessStyle.Render("✓ Directory exists")
		} else {
			info = styles.WarningStyle.Render("⚠ Directory does not exist (will be created)")
		}
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		desc,
		"",
		m.dirInput.View(),
		info,
	)
}

func (m SetupEnhancedModel) renderPatternStep() string {
	title := styles.TitleStyle.Render("Configure Excel File Patterns")
	desc := styles.MutedStyle.Render("Specify which Excel files to track (e.g., *.xlsx, reports/*.xlsx)")
	
	examples := styles.MutedStyle.Render(`
Examples:
  *.xlsx           - All Excel files in the directory
  reports/*.xlsx   - Excel files in the reports subdirectory
  Budget*.xlsx     - Files starting with "Budget"`)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		desc,
		"",
		m.patternInput.View(),
		examples,
	)
}

func (m *SetupEnhancedModel) renderGitStep() string {
	title := styles.TitleStyle.Render("Git Integration Settings")
	desc := styles.MutedStyle.Render("Configure how GitCells interacts with Git")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		desc,
		"",
		m.autoCommit.View(),
		"",
		m.autoPush.View(),
		"",
		m.commitTemplate.View(),
	)
}

func (m SetupEnhancedModel) renderReviewStep() string {
	title := styles.TitleStyle.Render("Review Configuration")
	desc := styles.MutedStyle.Render("Please review your settings before initializing")
	
	// Prepare config for display
	m.config = types.SetupConfig{
		Directory:      m.dirInput.Value(),
		Pattern:        m.patternInput.Value(),
		AutoCommit:     m.autoCommit.Checked(),
		AutoPush:       m.autoPush.Checked(),
		CommitTemplate: m.commitTemplate.Value(),
	}
	
	configBox := styles.BoxStyle.Render(fmt.Sprintf(`
Directory:        %s
Pattern:          %s
Auto-commit:      %v
Auto-push:        %v
Commit template:  %s`,
		m.config.Directory,
		m.config.Pattern,
		m.config.AutoCommit,
		m.config.AutoPush,
		m.config.CommitTemplate,
	))
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		desc,
		"",
		configBox,
		"",
		styles.SuccessStyle.Render("Press Enter to initialize GitCells with these settings"),
	)
}

func (m SetupEnhancedModel) renderComplete() string {
	completeBox := styles.BoxStyle.
		Width(60).
		Render(
			styles.SuccessStyle.Render("✓ GitCells Setup Complete!") + "\n\n" +
				"Your repository has been initialized with the following:\n\n" +
				"• Configuration saved to .gitcells.yaml\n" +
				"• Git repository initialized\n" +
				"• .gitignore created with Excel patterns\n\n" +
				"You can now run 'gitcells watch' to start monitoring Excel files.",
		)

	return styles.Center(m.width, m.height, completeBox)
}

func (m *SetupEnhancedModel) focusNext() {
	m.blurCurrent()
	
	switch m.step {
	case 0, 1:
		// Only one input per step
		return
	case 2:
		m.focused++
		if m.focused > 2 {
			m.focused = 0
		}
	}
	
	m.focusCurrentInput()
}

func (m *SetupEnhancedModel) focusPrev() {
	m.blurCurrent()
	
	switch m.step {
	case 0, 1:
		// Only one input per step
		return
	case 2:
		m.focused--
		if m.focused < 0 {
			m.focused = 2
		}
	}
	
	m.focusCurrentInput()
}

func (m *SetupEnhancedModel) focusCurrentInput() tea.Cmd {
	switch m.step {
	case 0:
		return m.dirInput.Focus()
	case 1:
		return m.patternInput.Focus()
	case 2:
		switch m.focused {
		case 0:
			m.autoCommit.Focus()
		case 1:
			m.autoPush.Focus()
		case 2:
			return m.commitTemplate.Focus()
		}
	}
	return nil
}

func (m *SetupEnhancedModel) blurCurrent() {
	m.dirInput.Blur()
	m.patternInput.Blur()
	m.autoCommit.Blur()
	m.autoPush.Blur()
	m.commitTemplate.Blur()
}

func (m SetupEnhancedModel) canProceed() bool {
	switch m.step {
	case 0:
		return m.dirInput.Value() != ""
	case 1:
		return m.patternInput.Value() != ""
	case 2:
		return m.commitTemplate.Value() != ""
	default:
		return true
	}
}

func (m *SetupEnhancedModel) saveConfig() {
	m.config = types.SetupConfig{
		Directory:      m.dirInput.Value(),
		Pattern:        m.patternInput.Value(),
		AutoCommit:     m.autoCommit.Checked(),
		AutoPush:       m.autoPush.Checked(),
		CommitTemplate: m.commitTemplate.Value(),
	}
	
	// Create directory if needed
	os.MkdirAll(m.config.Directory, 0755)
	
	// Save configuration using adapter
	configAdapter := adapter.NewConfigAdapter(m.config.Directory)
	if err := configAdapter.SaveSetupConfig(m.config); err != nil {
		m.error = fmt.Sprintf("Failed to save configuration: %v", err)
		return
	}
	
	// Create .gitignore
	if err := configAdapter.CreateGitIgnore(m.config.Directory); err != nil {
		m.error = fmt.Sprintf("Failed to create .gitignore: %v", err)
		return
	}
	
	// Initialize git repository
	if _, err := adapter.NewGitAdapter(m.config.Directory); err != nil {
		m.error = fmt.Sprintf("Failed to initialize git repository: %v", err)
		return
	}
}

// Helper to get absolute path
func absPath(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}