package git

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/Classic-Homes/sheetsync/pkg/models"
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
	ResolveOurs         ConflictResolutionStrategy = "ours"         // Keep our changes
	ResolveTheirs       ConflictResolutionStrategy = "theirs"       // Keep their changes
	ResolveBoth         ConflictResolutionStrategy = "both"         // Keep both (ours first)
	ResolveManual       ConflictResolutionStrategy = "manual"       // Require manual resolution
	ResolveSmartMerge   ConflictResolutionStrategy = "smart"        // Intelligent merge for Excel JSON
	ResolveNewestValue  ConflictResolutionStrategy = "newest"       // Use the most recent timestamp
	ResolveInteractive  ConflictResolutionStrategy = "interactive"  // Prompt user for each conflict
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
	case ResolveSmartMerge:
		return c.smartMergeExcelJSON(ourCode, theirCode)
	case ResolveNewestValue:
		return c.resolveByTimestamp(ourCode, theirCode)
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

// smartMergeExcelJSON attempts to intelligently merge Excel JSON conflicts
func (c *Client) smartMergeExcelJSON(ourCode, theirCode []string) []string {
	// Try to parse both sides as JSON
	ourJSON := strings.Join(ourCode, "\n")
	theirJSON := strings.Join(theirCode, "\n")

	var ourDoc, theirDoc models.ExcelDocument
	ourErr := json.Unmarshal([]byte(ourJSON), &ourDoc)
	theirErr := json.Unmarshal([]byte(theirJSON), &theirDoc)

	// If both are valid Excel JSON documents, merge them intelligently
	if ourErr == nil && theirErr == nil {
		merged := c.mergeExcelDocuments(&ourDoc, &theirDoc)
		if mergedJSON, err := json.MarshalIndent(merged, "", "  "); err == nil {
			return strings.Split(string(mergedJSON), "\n")
		}
	}

	// If smart merge fails, fall back to timestamp-based resolution
	c.logger.Warn("Smart merge failed, falling back to timestamp resolution")
	return c.resolveByTimestamp(ourCode, theirCode)
}

// mergeExcelDocuments merges two Excel documents intelligently
func (c *Client) mergeExcelDocuments(ours, theirs *models.ExcelDocument) *models.ExcelDocument {
	merged := &models.ExcelDocument{
		Version:      ours.Version,
		Metadata:     ours.Metadata, // Keep our metadata as base
		Sheets:       []models.Sheet{},
		DefinedNames: make(map[string]string),
		Properties:   ours.Properties,
	}

	// Use the newer modification time
	if theirs.Metadata.Modified.After(ours.Metadata.Modified) {
		merged.Metadata.Modified = theirs.Metadata.Modified
	}

	// Create sheet maps for easier lookup
	ourSheets := make(map[string]*models.Sheet)
	theirSheets := make(map[string]*models.Sheet)

	for i := range ours.Sheets {
		ourSheets[ours.Sheets[i].Name] = &ours.Sheets[i]
	}
	for i := range theirs.Sheets {
		theirSheets[theirs.Sheets[i].Name] = &theirs.Sheets[i]
	}

	// Merge sheets
	allSheetNames := make(map[string]bool)
	for name := range ourSheets {
		allSheetNames[name] = true
	}
	for name := range theirSheets {
		allSheetNames[name] = true
	}

	for sheetName := range allSheetNames {
		ourSheet, hasOur := ourSheets[sheetName]
		theirSheet, hasTheir := theirSheets[sheetName]

		if hasOur && hasTheir {
			// Merge sheet cells
			mergedSheet := c.mergeSheetsIntelligently(ourSheet, theirSheet)
			merged.Sheets = append(merged.Sheets, *mergedSheet)
		} else if hasOur {
			// Only in ours
			merged.Sheets = append(merged.Sheets, *ourSheet)
		} else {
			// Only in theirs
			merged.Sheets = append(merged.Sheets, *theirSheet)
		}
	}

	// Merge defined names (theirs wins on conflicts)
	for name, value := range ours.DefinedNames {
		merged.DefinedNames[name] = value
	}
	for name, value := range theirs.DefinedNames {
		merged.DefinedNames[name] = value // Overwrites if exists
	}

	return merged
}

// mergeSheetsIntelligently merges two sheets by combining their cells
func (c *Client) mergeSheetsIntelligently(ours, theirs *models.Sheet) *models.Sheet {
	merged := &models.Sheet{
		Name:         ours.Name,
		Index:        ours.Index,
		Cells:        make(map[string]models.Cell),
		MergedCells:  []models.MergedCell{}, // Start empty, will be populated
		RowHeights:   make(map[int]float64),
		ColumnWidths: make(map[string]float64),
		Hidden:       ours.Hidden,
		Protection:   ours.Protection,
	}

	// Start with our cells
	for cellRef, cell := range ours.Cells {
		merged.Cells[cellRef] = cell
	}

	// Add/override with their cells
	for cellRef, theirCell := range theirs.Cells {
		ourCell, exists := merged.Cells[cellRef]
		
		if !exists {
			// New cell, add it
			merged.Cells[cellRef] = theirCell
		} else {
			// Cell exists in both, choose the one with more recent formula or non-empty value
			if theirCell.Formula != "" && ourCell.Formula == "" {
				// They have formula, we don't
				merged.Cells[cellRef] = theirCell
			} else if theirCell.Value != nil && theirCell.Value != "" && (ourCell.Value == nil || ourCell.Value == "") {
				// They have value, we don't
				merged.Cells[cellRef] = theirCell
			}
			// Otherwise keep ours
		}
	}

	// Merge row heights and column widths (use the larger values)
	for row, height := range ours.RowHeights {
		merged.RowHeights[row] = height
	}
	for row, height := range theirs.RowHeights {
		if existingHeight, exists := merged.RowHeights[row]; !exists || height > existingHeight {
			merged.RowHeights[row] = height
		}
	}

	for col, width := range ours.ColumnWidths {
		merged.ColumnWidths[col] = width
	}
	for col, width := range theirs.ColumnWidths {
		if existingWidth, exists := merged.ColumnWidths[col]; !exists || width > existingWidth {
			merged.ColumnWidths[col] = width
		}
	}

	// Merge merged cells (combine both lists, remove duplicates)
	mergedCellsMap := make(map[string]bool)
	for _, mc := range ours.MergedCells {
		if !mergedCellsMap[mc.Range] {
			merged.MergedCells = append(merged.MergedCells, mc)
			mergedCellsMap[mc.Range] = true
		}
	}
	for _, mc := range theirs.MergedCells {
		if !mergedCellsMap[mc.Range] {
			merged.MergedCells = append(merged.MergedCells, mc)
			mergedCellsMap[mc.Range] = true
		}
	}

	return merged
}

// resolveByTimestamp chooses the version with the newer timestamp
func (c *Client) resolveByTimestamp(ourCode, theirCode []string) []string {
	ourJSON := strings.Join(ourCode, "\n")
	theirJSON := strings.Join(theirCode, "\n")

	var ourDoc, theirDoc models.ExcelDocument
	ourErr := json.Unmarshal([]byte(ourJSON), &ourDoc)
	theirErr := json.Unmarshal([]byte(theirJSON), &theirDoc)

	// If both are valid JSON with timestamps, use the newer one
	if ourErr == nil && theirErr == nil {
		if theirDoc.Metadata.Modified.After(ourDoc.Metadata.Modified) {
			c.logger.Info("Resolving conflict using newer timestamp (theirs)")
			return theirCode
		} else {
			c.logger.Info("Resolving conflict using newer timestamp (ours)")
			return ourCode
		}
	}

	// If one parses and the other doesn't, use the valid one
	if ourErr == nil && theirErr != nil {
		c.logger.Info("Resolving conflict: ours is valid JSON, theirs is not")
		return ourCode
	}
	if theirErr == nil && ourErr != nil {
		c.logger.Info("Resolving conflict: theirs is valid JSON, ours is not")
		return theirCode
	}

	// Both are invalid or we can't determine, default to ours
	c.logger.Warn("Unable to resolve conflict by timestamp, defaulting to ours")
	return ourCode
}

// ResolveExcelConflicts is a high-level method to resolve conflicts in Excel JSON files
func (c *Client) ResolveExcelConflicts(strategy ConflictResolutionStrategy) error {
	conflictedFiles, err := c.GetConflictedFiles()
	if err != nil {
		return fmt.Errorf("failed to get conflicted files: %w", err)
	}

	if len(conflictedFiles) == 0 {
		c.logger.Info("No conflicted files found")
		return nil
	}

	var resolvedFiles []string
	var failedFiles []string

	for _, file := range conflictedFiles {
		// Only process JSON files (Excel representations)
		if filepath.Ext(file) == ".json" {
			c.logger.Infof("Resolving conflicts in %s", file)
			if err := c.ResolveConflict(file, strategy); err != nil {
				c.logger.Errorf("Failed to resolve conflicts in %s: %v", file, err)
				failedFiles = append(failedFiles, file)
			} else {
				resolvedFiles = append(resolvedFiles, file)
			}
		}
	}

	// Stage the resolved files
	for _, file := range resolvedFiles {
		if _, err := c.worktree.Add(file); err != nil {
			c.logger.Warnf("Failed to stage resolved file %s: %v", file, err)
		}
	}

	if len(failedFiles) > 0 {
		return fmt.Errorf("failed to resolve conflicts in %d file(s): %v", len(failedFiles), failedFiles)
	}

	c.logger.Infof("Successfully resolved conflicts in %d file(s)", len(resolvedFiles))
	return nil
}