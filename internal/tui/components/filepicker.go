package components

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FilePicker struct {
	currentPath     string
	selectedPath    string
	items           []FileItem
	cursor          int
	showHidden      bool
	directoriesOnly bool
	extensions      []string
	height          int
	width           int
}

type FileItem struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
}

func NewFilePicker(startPath string, directoriesOnly bool) FilePicker {
	if startPath == "" {
		startPath, _ = os.Getwd()
	}

	fp := FilePicker{
		currentPath:     startPath,
		selectedPath:    startPath,
		directoriesOnly: directoriesOnly,
		height:          20,
		width:           60,
	}

	fp.loadItems()
	return fp
}

func (fp *FilePicker) SetExtensions(extensions []string) {
	fp.extensions = extensions
	fp.loadItems()
}

func (fp *FilePicker) SelectedPath() string {
	return fp.selectedPath
}

func (fp FilePicker) Init() tea.Cmd {
	return nil
}

func (fp FilePicker) Update(msg tea.Msg) (FilePicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		fp.width = msg.Width
		fp.height = msg.Height - 10 // Leave room for other UI elements

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if fp.cursor > 0 {
				fp.cursor--
			}

		case "down", "j":
			if fp.cursor < len(fp.items)-1 {
				fp.cursor++
			}

		case "enter", "right", "l":
			if fp.cursor < len(fp.items) {
				item := fp.items[fp.cursor]
				if item.IsDir {
					fp.currentPath = item.Path
					fp.selectedPath = item.Path
					fp.loadItems()
					fp.cursor = 0
				} else if !fp.directoriesOnly {
					fp.selectedPath = item.Path
				}
			}

		case "left", "h", "backspace":
			parent := filepath.Dir(fp.currentPath)
			if parent != fp.currentPath {
				fp.currentPath = parent
				fp.selectedPath = parent
				fp.loadItems()
				fp.cursor = 0
			}

		case ".":
			fp.showHidden = !fp.showHidden
			fp.loadItems()

		case "g":
			fp.cursor = 0

		case "G":
			fp.cursor = len(fp.items) - 1

		case "~":
			home, _ := os.UserHomeDir()
			fp.currentPath = home
			fp.selectedPath = home
			fp.loadItems()
			fp.cursor = 0
		}
	}

	return fp, nil
}

func (fp FilePicker) View() string {
	if len(fp.items) == 0 {
		return styles.BoxStyle.Render("No items to display")
	}

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		MarginBottom(1)

	header := headerStyle.Render(fp.formatPath(fp.currentPath))

	// Calculate visible items
	visibleHeight := fp.height - 4 // Account for header and borders
	if visibleHeight < 5 {
		visibleHeight = 5
	}

	startIdx := 0
	if fp.cursor >= visibleHeight {
		startIdx = fp.cursor - visibleHeight + 1
	}

	endIdx := startIdx + visibleHeight
	if endIdx > len(fp.items) {
		endIdx = len(fp.items)
	}

	// Build item list
	var items []string
	for i := startIdx; i < endIdx; i++ {
		item := fp.items[i]

		// Icon
		icon := "📄"
		if item.IsDir {
			icon = "📁"
		}

		// Name
		name := item.Name
		if item.IsDir {
			name += "/"
		}

		// Style
		style := lipgloss.NewStyle()
		if i == fp.cursor {
			style = style.
				Foreground(styles.Primary).
				Bold(true).
				Background(lipgloss.Color("236"))
		}

		// Size (for files)
		sizeStr := ""
		if !item.IsDir && !fp.directoriesOnly {
			sizeStr = formatFileSize(item.Size)
		}

		line := fmt.Sprintf(" %s %-40s %10s", icon, name, sizeStr)
		items = append(items, style.Render(line))
	}

	// Scroll indicator
	scrollInfo := ""
	if len(fp.items) > visibleHeight {
		scrollInfo = styles.MutedStyle.Render(
			fmt.Sprintf(" (%d/%d)", fp.cursor+1, len(fp.items)),
		)
	}

	// Help text
	help := styles.HelpStyle.Render(
		"↑/↓: Navigate • Enter: Select • ←: Parent • .: Toggle hidden • ~: Home",
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header+scrollInfo,
		strings.Join(items, "\n"),
		"",
		help,
	)

	return styles.BoxStyle.
		Width(fp.width).
		Height(fp.height).
		Render(content)
}

func (fp *FilePicker) loadItems() {
	fp.items = []FileItem{}

	// Add parent directory option if not at root
	if fp.currentPath != "/" && fp.currentPath != filepath.Dir(fp.currentPath) {
		fp.items = append(fp.items, FileItem{
			Name:  "..",
			Path:  filepath.Dir(fp.currentPath),
			IsDir: true,
		})
	}

	// Read directory
	files, err := os.ReadDir(fp.currentPath)
	if err != nil {
		return
	}

	for _, file := range files {
		// Skip hidden files unless enabled
		if !fp.showHidden && strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Skip files if directories only
		if fp.directoriesOnly && !file.IsDir() {
			continue
		}

		// Filter by extension if specified
		if !file.IsDir() && len(fp.extensions) > 0 {
			hasValidExt := false
			for _, ext := range fp.extensions {
				if strings.HasSuffix(strings.ToLower(file.Name()), strings.ToLower(ext)) {
					hasValidExt = true
					break
				}
			}
			if !hasValidExt {
				continue
			}
		}

		info, err := file.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}

		fp.items = append(fp.items, FileItem{
			Name:  file.Name(),
			Path:  filepath.Join(fp.currentPath, file.Name()),
			IsDir: file.IsDir(),
			Size:  size,
		})
	}

	// Sort directories first, then files
	sort.Slice(fp.items[1:], func(i, j int) bool {
		i++ // Skip ".." entry
		j++
		if fp.items[i].IsDir != fp.items[j].IsDir {
			return fp.items[i].IsDir
		}
		return strings.ToLower(fp.items[i].Name) < strings.ToLower(fp.items[j].Name)
	})
}

func (fp FilePicker) formatPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}

func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
