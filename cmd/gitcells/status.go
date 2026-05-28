package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type FileStatus struct {
	ExcelPath    string     `json:"excel_path"`
	JSONPath     string     `json:"json_path,omitempty"`
	Status       string     `json:"status"` // "synced", "modified", "new", "missing"
	ExcelModTime time.Time  `json:"excel_mod_time"`
	JSONModTime  time.Time  `json:"json_mod_time,omitempty"`
	ExcelSize    int64      `json:"excel_size"`
	JSONSize     int64      `json:"json_size,omitempty"`
	LastSyncTime *time.Time `json:"last_sync_time,omitempty"`
	HasChanges   bool       `json:"has_changes"`
}

func newStatusCommand(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [directory]",
		Short: "Show status of Excel files and their JSON representations",
		Long:  "Display the synchronization status of Excel files and JSON representations in the given directory (defaults to current directory)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			detailed, _ := cmd.Flags().GetBool("detailed")
			jsonMode, _ := cmd.Flags().GetBool("json")
			includePatterns, _ := cmd.Flags().GetStringSlice("include")

			dir, err := resolveDirArg(args)
			if err != nil {
				return err
			}

			// Find all Excel files
			excelFiles, err := findExcelFiles(dir, includePatterns)
			if err != nil {
				return utils.WrapError(err, utils.ErrorTypeFileSystem, "findExcelFiles", "failed to scan for Excel files")
			}

			if len(excelFiles) == 0 {
				if jsonMode {
					return outputStatusJSON(nil)
				}
				fmt.Println("No Excel files found in the current directory")
				return nil
			}

			// Check status for each file
			statuses := make([]FileStatus, 0, len(excelFiles))
			for _, excelPath := range excelFiles {
				status, err := getFileStatus(excelPath, logger)
				if err != nil {
					logger.Warnf("Failed to get status for %s: %v", excelPath, err)
					continue
				}
				statuses = append(statuses, status)
			}

			// Display results
			if jsonMode {
				return outputStatusJSON(statuses)
			}
			displayStatus(statuses, detailed)

			return nil
		},
	}

	cmd.Flags().Bool("detailed", false, "show detailed status information")
	cmd.Flags().Bool("json", false, "output machine-readable JSON (no emoji or color)")
	cmd.Flags().StringSlice("include", []string{"*.xlsx", "*.xls", "*.xlsm"}, "file patterns to include (repeatable or comma-separated)")

	return cmd
}

func findExcelFiles(dir string, patterns []string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		// Skip .gitcells directory
		if strings.Contains(path, constants.GitCellsDir) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip temporary Excel files
		if strings.HasPrefix(d.Name(), constants.ExcelTempPrefix) {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check if file matches any pattern
		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, d.Name())
			if err != nil {
				continue
			}
			if matched {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

func getFileStatus(excelPath string, logger *logrus.Logger) (FileStatus, error) {
	status := FileStatus{
		ExcelPath: excelPath,
		Status:    "new",
	}

	// Get Excel file info
	excelInfo, err := os.Stat(excelPath)
	if err != nil {
		return status, err
	}
	status.ExcelModTime = excelInfo.ModTime()
	status.ExcelSize = excelInfo.Size()

	// Determine JSON path. Mirror chunking.go's behavior precisely —
	// when the Excel file is inside a git repo, chunks live at
	// <repoRoot>/.gitcells/data; when it isn't, chunks live beside the
	// Excel file at <excelDir>/.gitcells/data. The old fallback of "."
	// (cwd) silently reported every file as "new" whenever the user ran
	// `gitcells status` from outside the repo root.
	excelDir := filepath.Dir(excelPath)
	gitRoot, err := git.FindRepositoryRoot(excelDir)
	if err != nil {
		gitRoot = excelDir
		logger.Debugf("file %s is outside any git repository; using excel dir as data root", excelPath)
	}

	relPath, err := filepath.Rel(gitRoot, excelPath)
	if err != nil {
		relPath = excelPath
	}

	jsonDir := filepath.Join(gitRoot, constants.GitCellsDataDir, filepath.Dir(relPath))
	baseName := strings.TrimSuffix(filepath.Base(excelPath), filepath.Ext(excelPath))

	// Check for chunked files only
	chunkDir := filepath.Join(jsonDir, baseName+constants.ChunksDirSuffix)
	workbookJsonPath := filepath.Join(chunkDir, constants.WorkbookFileName)
	status.JSONPath = workbookJsonPath

	// Check if JSON exists
	jsonInfo, err := os.Stat(status.JSONPath)
	if err != nil {
		if os.IsNotExist(err) {
			status.Status = "new"
			return status, nil
		}
		return status, err
	}

	status.JSONModTime = jsonInfo.ModTime()
	status.JSONSize = jsonInfo.Size()

	// Read JSON metadata to check sync status
	metadata, err := readJSONMetadata(status.JSONPath)
	if err != nil {
		logger.Debugf("Failed to read JSON metadata: %v", err)
		// If we can't read metadata, compare modification times
		if status.ExcelModTime.After(status.JSONModTime) {
			status.Status = "modified"
			status.HasChanges = true
		} else {
			status.Status = "synced"
		}
	} else {
		// Use metadata for more accurate status
		if metadata.Modified.Equal(status.ExcelModTime) || metadata.Modified.After(status.ExcelModTime) {
			status.Status = "synced"
			status.LastSyncTime = &metadata.Created
		} else {
			status.Status = "modified"
			status.HasChanges = true
			status.LastSyncTime = &metadata.Created
		}
	}

	return status, nil
}

func readJSONMetadata(jsonPath string) (*models.DocumentMetadata, error) {
	file, err := os.Open(jsonPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var doc models.ExcelDocument
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&doc); err != nil {
		return nil, err
	}

	return &doc.Metadata, nil
}

func displayStatus(statuses []FileStatus, detailed bool) {
	// Summary counts
	counts := map[string]int{
		"synced":   0,
		"modified": 0,
		"new":      0,
		"missing":  0,
	}

	for _, s := range statuses {
		counts[s.Status]++
	}

	// Display summary
	fmt.Println("\n📊 GitCells Status Summary")
	fmt.Println("═══════════════════════════")
	fmt.Printf("✅ Synced:    %d files\n", counts["synced"])
	fmt.Printf("📝 Modified:  %d files\n", counts["modified"])
	fmt.Printf("🆕 New:       %d files\n", counts["new"])
	fmt.Printf("❌ Missing:   %d files\n", counts["missing"])
	fmt.Println()

	// Display file details
	if len(statuses) > 0 {
		fmt.Println("File Status:")
		fmt.Println("────────────")

		for _, s := range statuses {
			statusIcon := getStatusIcon(s.Status)
			fmt.Printf("%s %s\n", statusIcon, s.ExcelPath)

			if detailed {
				fmt.Printf("   Status: %s\n", s.Status)
				fmt.Printf("   Excel modified: %s\n", s.ExcelModTime.Format("2006-01-02 15:04:05"))
				if s.JSONPath != "" && s.Status != "new" {
					fmt.Printf("   JSON modified:  %s\n", s.JSONModTime.Format("2006-01-02 15:04:05"))
					if s.LastSyncTime != nil {
						fmt.Printf("   Last sync:      %s\n", s.LastSyncTime.Format("2006-01-02 15:04:05"))
					}
				}
				fmt.Printf("   Size: %s\n", formatFileSize(s.ExcelSize))
				fmt.Println()
			}
		}
	}

	// Display action hints. `sync` converts both modified AND new files
	// (see sync.go: status == "modified" || status == "new"), so reflect
	// that here. Without --commit, sync only updates the JSON chunks; the
	// hint mentions both so the user doesn't follow up wondering why the
	// repo isn't dirty afterward.
	if counts["modified"] > 0 || counts["new"] > 0 {
		fmt.Println("\nHint: run 'gitcells sync' to convert modified and new files (add --commit to also commit them)")
	}
}

// outputStatusJSON emits the same FileStatus slice + a summary block as
// JSON to stdout. No emoji, no color — meant for scripting.
func outputStatusJSON(statuses []FileStatus) error {
	counts := map[string]int{
		"synced":   0,
		"modified": 0,
		"new":      0,
		"missing":  0,
	}
	for _, s := range statuses {
		counts[s.Status]++
	}
	payload := struct {
		Files   []FileStatus   `json:"files"`
		Summary map[string]int `json:"summary"`
	}{
		Files:   statuses,
		Summary: counts,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(payload)
}

func getStatusIcon(status string) string {
	switch status {
	case "synced":
		return "✅"
	case "modified":
		return "📝"
	case "new":
		return "🆕"
	case "missing":
		return "❌"
	default:
		return "❓"
	}
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
