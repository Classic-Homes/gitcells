package git

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
)

// ConflictMarker represents different types of conflict markers
type ConflictMarker string

const (
	ConflictStart  ConflictMarker = "<<<<<<< "
	ConflictMiddle ConflictMarker = "======="
	ConflictEnd    ConflictMarker = ">>>>>>> "
)

// ConflictInfo represents information about a merge conflict
type ConflictInfo struct {
	FilePath     string
	HasConflicts bool
	Conflicts    []Conflict
}

// Conflict represents a single conflict within a file
type Conflict struct {
	StartLine int
	EndLine   int
	OurCode   []string
	TheirCode []string
	Base      []string
}

// DetectConflicts scans a file for merge conflict markers
func (c *Client) DetectConflicts(filePath string) (*ConflictInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	info := &ConflictInfo{
		FilePath:  filePath,
		Conflicts: []Conflict{},
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentConflict *Conflict
	var inOurSection, inTheirSection bool

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, string(ConflictStart)):
			// Start of conflict
			currentConflict = &Conflict{
				StartLine: lineNum,
				OurCode:   []string{},
				TheirCode: []string{},
			}
			inOurSection = true
			inTheirSection = false
			info.HasConflicts = true

		case line == string(ConflictMiddle):
			// Middle of conflict - switch from "ours" to "theirs"
			inOurSection = false
			inTheirSection = true

		case strings.HasPrefix(line, string(ConflictEnd)):
			// End of conflict
			if currentConflict != nil {
				currentConflict.EndLine = lineNum
				info.Conflicts = append(info.Conflicts, *currentConflict)
				currentConflict = nil
			}
			inOurSection = false
			inTheirSection = false

		default:
			// Regular content line
			if currentConflict != nil {
				if inOurSection {
					currentConflict.OurCode = append(currentConflict.OurCode, line)
				} else if inTheirSection {
					currentConflict.TheirCode = append(currentConflict.TheirCode, line)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file %s: %w", filePath, err)
	}

	return info, nil
}

// ResolveConflict resolves a conflict by choosing a resolution strategy
func (c *Client) ResolveConflict(filePath string, strategy ConflictResolutionStrategy) error {
	conflictInfo, err := c.DetectConflicts(filePath)
	if err != nil {
		return fmt.Errorf("failed to detect conflicts: %w", err)
	}

	if !conflictInfo.HasConflicts {
		c.logger.Infof("No conflicts found in %s", filePath)
		return nil
	}

	// Read the entire file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	var resolvedLines []string

	i := 0
	for i < len(lines) {
		line := lines[i]

		if strings.HasPrefix(line, string(ConflictStart)) {
			// Find the end of this conflict
			conflictStart := i
			middleIndex := -1
			endIndex := -1

			for j := i + 1; j < len(lines); j++ {
				if lines[j] == string(ConflictMiddle) && middleIndex == -1 {
					middleIndex = j
				} else if strings.HasPrefix(lines[j], string(ConflictEnd)) {
					endIndex = j
					break
				}
			}

			if middleIndex == -1 || endIndex == -1 {
				return fmt.Errorf("malformed conflict in file %s", filePath)
			}

			// Extract our code and their code
			ourCode := lines[conflictStart+1 : middleIndex]
			theirCode := lines[middleIndex+1 : endIndex]

			// Apply resolution strategy
			resolvedCode := c.applyResolutionStrategy(ourCode, theirCode, strategy)
			resolvedLines = append(resolvedLines, resolvedCode...)

			// Skip to after the conflict
			i = endIndex + 1
		} else {
			resolvedLines = append(resolvedLines, line)
			i++
		}
	}

	// Write the resolved content back to the file
	resolvedContent := strings.Join(resolvedLines, "\n")
	if err := os.WriteFile(filePath, []byte(resolvedContent), 0644); err != nil {
		return fmt.Errorf("failed to write resolved file: %w", err)
	}

	c.logger.Infof("Resolved conflicts in %s using strategy: %s", filePath, strategy)
	return nil
}

// ConflictResolutionStrategy defines how conflicts should be resolved
type ConflictResolutionStrategy string

const (
	ResolveOurs   ConflictResolutionStrategy = "ours"   // Keep our changes
	ResolveTheirs ConflictResolutionStrategy = "theirs" // Keep their changes
	ResolveBoth   ConflictResolutionStrategy = "both"   // Keep both (ours first)
	ResolveManual ConflictResolutionStrategy = "manual" // Require manual resolution
)

func (c *Client) applyResolutionStrategy(ourCode, theirCode []string, strategy ConflictResolutionStrategy) []string {
	switch strategy {
	case ResolveOurs:
		return ourCode
	case ResolveTheirs:
		return theirCode
	case ResolveBoth:
		result := make([]string, 0, len(ourCode)+len(theirCode))
		result = append(result, ourCode...)
		result = append(result, theirCode...)
		return result
	default:
		// For manual resolution, we'll leave the conflict markers
		result := make([]string, 0, len(ourCode)+len(theirCode)+3)
		result = append(result, string(ConflictStart)+"HEAD")
		result = append(result, ourCode...)
		result = append(result, string(ConflictMiddle))
		result = append(result, theirCode...)
		result = append(result, string(ConflictEnd)+"incoming")
		return result
	}
}

// GetConflictedFiles returns a list of files with merge conflicts
func (c *Client) GetConflictedFiles() ([]string, error) {
	status, err := c.worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var conflictedFiles []string
	for file, fileStatus := range status {
		// Check if file has conflicts (both modified in working tree and staging area)
		if fileStatus.Staging == git.UpdatedButUnmerged || fileStatus.Worktree == git.UpdatedButUnmerged {
			conflictedFiles = append(conflictedFiles, file)
		}
	}

	return conflictedFiles, nil
}

// HasConflicts checks if there are any conflicted files in the repository
func (c *Client) HasConflicts() (bool, error) {
	conflictedFiles, err := c.GetConflictedFiles()
	if err != nil {
		return false, err
	}
	return len(conflictedFiles) > 0, nil
}