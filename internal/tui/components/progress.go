package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Classic-Homes/gitcells/internal/tui/styles"
)

type ProgressBar struct {
	width      int
	current    int
	total      int
	label      string
	showPercent bool
}

func NewProgressBar(total int) ProgressBar {
	return ProgressBar{
		width:       40,
		total:       total,
		showPercent: true,
	}
}

func (p *ProgressBar) SetLabel(label string) {
	p.label = label
}

func (p *ProgressBar) SetProgress(current int) {
	p.current = current
	if p.current > p.total {
		p.current = p.total
	}
	if p.current < 0 {
		p.current = 0
	}
}

func (p *ProgressBar) SetWidth(width int) {
	p.width = width
}

func (p ProgressBar) Percentage() float64 {
	if p.total == 0 {
		return 0
	}
	return float64(p.current) / float64(p.total) * 100
}

func (p ProgressBar) View() string {
	percentage := p.Percentage()
	
	// Calculate filled width
	barWidth := p.width
	if p.showPercent {
		barWidth -= 7 // Space for percentage
	}
	if p.label != "" {
		barWidth -= len(p.label) + 2 // Space for label
	}
	
	filled := int(float64(barWidth) * (percentage / 100))
	if filled > barWidth {
		filled = barWidth
	}
	
	// Build the bar
	filledStyle := lipgloss.NewStyle().
		Foreground(styles.Success).
		Bold(true)
		
	emptyStyle := lipgloss.NewStyle().
		Foreground(styles.Muted)
	
	bar := filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", barWidth-filled))
	
	// Add percentage
	percentStr := ""
	if p.showPercent {
		percentStr = fmt.Sprintf(" %3.0f%%", percentage)
	}
	
	// Add label
	labelStr := ""
	if p.label != "" {
		labelStr = p.label + " "
	}
	
	return labelStr + bar + percentStr
}

// MultiProgress manages multiple progress bars
type MultiProgress struct {
	bars   map[string]*ProgressBar
	order  []string
	width  int
}

func NewMultiProgress() MultiProgress {
	return MultiProgress{
		bars:  make(map[string]*ProgressBar),
		order: []string{},
		width: 60,
	}
}

func (m *MultiProgress) AddBar(id string, label string, total int) {
	bar := NewProgressBar(total)
	bar.SetLabel(label)
	bar.SetWidth(m.width - len(label) - 2)
	m.bars[id] = &bar
	m.order = append(m.order, id)
}

func (m *MultiProgress) UpdateBar(id string, current int) {
	if bar, exists := m.bars[id]; exists {
		bar.SetProgress(current)
	}
}

func (m *MultiProgress) RemoveBar(id string) {
	delete(m.bars, id)
	newOrder := []string{}
	for _, oid := range m.order {
		if oid != id {
			newOrder = append(newOrder, oid)
		}
	}
	m.order = newOrder
}

func (m MultiProgress) View() string {
	if len(m.bars) == 0 {
		return ""
	}
	
	lines := []string{}
	for _, id := range m.order {
		if bar, exists := m.bars[id]; exists {
			lines = append(lines, bar.View())
		}
	}
	
	return strings.Join(lines, "\n")
}

// SpinnerProgress shows an indeterminate progress spinner
type SpinnerProgress struct {
	frames []string
	frame  int
	label  string
}

func NewSpinnerProgress(label string) SpinnerProgress {
	return SpinnerProgress{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		label:  label,
	}
}

func (s *SpinnerProgress) Next() {
	s.frame = (s.frame + 1) % len(s.frames)
}

func (s SpinnerProgress) View() string {
	spinnerStyle := lipgloss.NewStyle().
		Foreground(styles.Primary)
		
	return spinnerStyle.Render(s.frames[s.frame]) + " " + s.label
}