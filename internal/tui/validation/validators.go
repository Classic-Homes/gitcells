package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/constants"
)

// ValidateDirectory checks if a directory path is valid
func ValidateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("directory path cannot be empty")
	}

	// Expand home directory
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		path = filepath.Join(home, path[1:])
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Check if it's a valid path format
	if !filepath.IsAbs(absPath) {
		return fmt.Errorf("path must be absolute")
	}

	// Check if directory exists and is writable
	if stat, err := os.Stat(absPath); err == nil {
		// Directory exists
		if !stat.IsDir() {
			return fmt.Errorf("path exists but is not a directory")
		}

		// Check write permissions
		testFile := filepath.Join(absPath, ".gitcells_test")
		if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
			return fmt.Errorf("directory is not writable: %w", err)
		}
		os.Remove(testFile)
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("cannot access directory: %w", err)
	} else {
		// Directory doesn't exist - check if parent is writable
		parent := filepath.Dir(absPath)
		if parent != absPath { // Avoid infinite loop at root
			if stat, err := os.Stat(parent); err == nil {
				if !stat.IsDir() {
					return fmt.Errorf("parent path is not a directory")
				}
				// Check if we can create in parent
				testDir := filepath.Join(parent, ".gitcells_test_dir")
				if err := os.Mkdir(testDir, 0755); err != nil {
					return fmt.Errorf("cannot create directory (parent not writable): %w", err)
				}
				os.Remove(testDir)
			} else if os.IsNotExist(err) {
				return fmt.Errorf("parent directory does not exist")
			} else {
				return fmt.Errorf("cannot access parent directory: %w", err)
			}
		}
	}

	return nil
}

// ValidateExcelPattern checks if a file pattern is valid for Excel files
func ValidateExcelPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	// Check if pattern is valid
	_, err := filepath.Match(pattern, "test.xlsx")
	if err != nil {
		return fmt.Errorf("invalid pattern syntax: %w", err)
	}

	// Ensure it targets Excel files
	validExtensions := make([]string, len(constants.ExcelExtensions)+1)
	copy(validExtensions, constants.ExcelExtensions)
	validExtensions[len(constants.ExcelExtensions)] = ".xlsb"
	hasValidExt := false

	for _, ext := range validExtensions {
		if strings.Contains(pattern, ext) {
			hasValidExt = true
			break
		}
	}

	if !hasValidExt && !strings.Contains(pattern, "*") {
		return fmt.Errorf("pattern should target Excel files (*.xlsx, *.xls, *.xlsm, *.xlsb)")
	}

	return nil
}

// ValidateCommitTemplate checks if a commit template is valid
func ValidateCommitTemplate(template string) error {
	if template == "" {
		return fmt.Errorf("commit template cannot be empty")
	}

	// Check for required placeholders
	validPlaceholders := []string{"{action}", "{filename}", "{timestamp}", "{user}"}
	hasPlaceholder := false

	for _, placeholder := range validPlaceholders {
		if strings.Contains(template, placeholder) {
			hasPlaceholder = true
			break
		}
	}

	if !hasPlaceholder {
		return fmt.Errorf("template should contain at least one placeholder: %s", strings.Join(validPlaceholders, ", "))
	}

	// Check for unmatched braces
	openCount := strings.Count(template, "{")
	closeCount := strings.Count(template, "}")
	if openCount != closeCount {
		return fmt.Errorf("template has unmatched braces")
	}

	return nil
}

// ValidateBranchName checks if a branch name is valid for git
func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Check length
	if len(name) > 255 {
		return fmt.Errorf("branch name too long (max 255 characters)")
	}

	// Check for invalid characters
	invalidChars := []string{" ", "~", "^", ":", "?", "*", "[", "\\", "..", "@{", "//"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("branch name contains invalid character: %s", char)
		}
	}

	// Check for invalid start/end
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot start or end with '/'")
	}

	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot start or end with '.'")
	}

	if strings.HasSuffix(name, ".lock") {
		return fmt.Errorf("branch name cannot end with '.lock'")
	}

	return nil
}

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// DirectoryInfo provides information about a directory
type DirectoryInfo struct {
	Exists      bool
	IsDirectory bool
	IsWritable  bool
	IsGitRepo   bool
	HasGitCells bool
	ExcelCount  int
}

// InspectDirectory gathers information about a directory
func InspectDirectory(path string) (*DirectoryInfo, error) {
	info := &DirectoryInfo{}

	// Check if exists
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return info, nil
		}
		return nil, err
	}

	info.Exists = true
	info.IsDirectory = stat.IsDir()

	if !info.IsDirectory {
		return info, nil
	}

	// Check if writable
	testFile := filepath.Join(path, ".gitcells_test")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err == nil {
		info.IsWritable = true
		os.Remove(testFile)
	}

	// Check for .git directory
	gitPath := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitPath); err == nil && stat.IsDir() {
		info.IsGitRepo = true
	}

	// Check for .gitcells.yaml
	gitcellsPath := filepath.Join(path, constants.ConfigFileName)
	if _, err := os.Stat(gitcellsPath); err == nil {
		info.HasGitCells = true
	}

	// Count Excel files
	excelExts := make([]string, len(constants.ExcelExtensions)+1)
	copy(excelExts, constants.ExcelExtensions)
	excelExts[len(constants.ExcelExtensions)] = ".xlsb"
	walkErr := filepath.Walk(path, func(p string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		if fileInfo.IsDir() {
			// Skip hidden directories
			if strings.HasPrefix(fileInfo.Name(), ".") && fileInfo.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(fileInfo.Name()))
		for _, validExt := range excelExts {
			if ext == validExt {
				info.ExcelCount++
				break
			}
		}

		return nil
	})
	// Ignore walk errors as they're not critical for directory info
	_ = walkErr

	return info, nil
}
