package git

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/pkg/models"
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
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log the error but don't override the main function's error
			fmt.Printf("Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

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
	ResolveOurs        ConflictResolutionStrategy = "ours"        // Keep our changes
	ResolveTheirs      ConflictResolutionStrategy = "theirs"      // Keep their changes
	ResolveBoth        ConflictResolutionStrategy = "both"        // Keep both (ours first)
	ResolveManual      ConflictResolutionStrategy = "manual"      // Require manual resolution
	ResolveSmartMerge  ConflictResolutionStrategy = "smart"       // Intelligent merge for Excel JSON
	ResolveNewestValue ConflictResolutionStrategy = "newest"      // Use the most recent timestamp
	ResolveInteractive ConflictResolutionStrategy = "interactive" // Prompt user for each conflict
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
	case ResolveInteractive:
		return c.resolveInteractively(ourCode, theirCode)
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

// resolveInteractively prompts the user to choose how to resolve each conflict
func (c *Client) resolveInteractively(ourCode, theirCode []string) []string {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("INTERACTIVE CONFLICT RESOLUTION")
	fmt.Println(strings.Repeat("=", 80))

	// Try to parse as Excel JSON to provide better context
	ourJSON := strings.Join(ourCode, "\n")
	theirJSON := strings.Join(theirCode, "\n")

	var ourDoc, theirDoc models.ExcelDocument
	ourErr := json.Unmarshal([]byte(ourJSON), &ourDoc)
	theirErr := json.Unmarshal([]byte(theirJSON), &theirDoc)

	if ourErr == nil && theirErr == nil {
		// Both are valid Excel JSON - show detailed diff
		fmt.Println("📊 Excel Document Conflict Detected")
		fmt.Printf("Your version: Modified %s\n", ourDoc.Metadata.Modified.Format("2006-01-02 15:04:05"))
		fmt.Printf("Their version: Modified %s\n", theirDoc.Metadata.Modified.Format("2006-01-02 15:04:05"))

		// Show sheet differences
		c.showExcelConflictSummary(&ourDoc, &theirDoc)
	} else {
		// Generic conflict display
		fmt.Println("📄 File Conflict Detected")
	}

	fmt.Println("\nYOUR VERSION (HEAD):")
	fmt.Println(strings.Repeat("-", 40))
	for i, line := range ourCode {
		fmt.Printf("%3d: %s\n", i+1, line)
	}

	fmt.Println("\nTHEIR VERSION (incoming):")
	fmt.Println(strings.Repeat("-", 40))
	for i, line := range theirCode {
		fmt.Printf("%3d: %s\n", i+1, line)
	}

	fmt.Println("\nRESOLUTION OPTIONS:")
	fmt.Println("1. Keep your version (y/yours)")
	fmt.Println("2. Keep their version (t/theirs)")
	fmt.Println("3. Keep both versions (b/both)")
	fmt.Println("4. Smart merge (s/smart) - Excel-aware merge")
	fmt.Println("5. Use newest timestamp (n/newest)")
	fmt.Println("6. Edit manually (e/edit)")
	fmt.Println("7. Skip this conflict (skip)")

	for {
		fmt.Print("\nChoose resolution [1-7, or y/t/b/s/n/e/skip]: ")

		var choice string
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}
		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "1", "y", "yours":
			fmt.Println("✅ Keeping your version")
			return ourCode
		case "2", "t", "theirs":
			fmt.Println("✅ Keeping their version")
			return theirCode
		case "3", "b", "both":
			fmt.Println("✅ Keeping both versions")
			result := make([]string, 0, len(ourCode)+len(theirCode))
			result = append(result, ourCode...)
			result = append(result, theirCode...)
			return result
		case "4", "s", "smart":
			fmt.Println("✅ Applying smart merge")
			return c.smartMergeExcelJSON(ourCode, theirCode)
		case "5", "n", "newest":
			fmt.Println("✅ Using newest timestamp")
			return c.resolveByTimestamp(ourCode, theirCode)
		case "6", "e", "edit":
			fmt.Println("✅ Opening for manual editing")
			return c.editManually(ourCode, theirCode)
		case "7", "skip":
			fmt.Println("⏭️  Skipping - leaving conflict markers")
			result := make([]string, 0, len(ourCode)+len(theirCode)+3)
			result = append(result, string(ConflictStart)+"HEAD")
			result = append(result, ourCode...)
			result = append(result, string(ConflictMiddle))
			result = append(result, theirCode...)
			result = append(result, string(ConflictEnd)+"incoming")
			return result
		default:
			fmt.Printf("❌ Invalid choice '%s'. Please choose 1-7 or use shortcuts.\n", choice)
		}
	}
}

// showExcelConflictSummary shows a summary of differences between two Excel documents
func (c *Client) showExcelConflictSummary(ours, theirs *models.ExcelDocument) {
	fmt.Println("\n📋 CONFLICT SUMMARY:")

	// Compare sheet counts
	if len(ours.Sheets) != len(theirs.Sheets) {
		fmt.Printf("• Sheet count differs: yours=%d, theirs=%d\n", len(ours.Sheets), len(theirs.Sheets))
	}

	// Create sheet maps
	ourSheets := make(map[string]*models.Sheet)
	theirSheets := make(map[string]*models.Sheet)

	for i := range ours.Sheets {
		ourSheets[ours.Sheets[i].Name] = &ours.Sheets[i]
	}
	for i := range theirs.Sheets {
		theirSheets[theirs.Sheets[i].Name] = &theirs.Sheets[i]
	}

	// Find sheet differences
	allSheets := make(map[string]bool)
	for name := range ourSheets {
		allSheets[name] = true
	}
	for name := range theirSheets {
		allSheets[name] = true
	}

	for sheetName := range allSheets {
		ourSheet, hasOur := ourSheets[sheetName]
		theirSheet, hasTheir := theirSheets[sheetName]

		if hasOur && hasTheir {
			ourCells := len(ourSheet.Cells)
			theirCells := len(theirSheet.Cells)
			if ourCells != theirCells {
				fmt.Printf("• Sheet '%s': cell count differs (yours=%d, theirs=%d)\n", sheetName, ourCells, theirCells)
			}
		} else if hasOur && !hasTheir {
			fmt.Printf("• Sheet '%s': only in your version\n", sheetName)
		} else if !hasOur && hasTheir {
			fmt.Printf("• Sheet '%s': only in their version\n", sheetName)
		}
	}

	// Compare defined names
	if len(ours.DefinedNames) != len(theirs.DefinedNames) {
		fmt.Printf("• Defined names count differs: yours=%d, theirs=%d\n", len(ours.DefinedNames), len(theirs.DefinedNames))
	}
}

// editManually allows the user to manually edit the conflicted content
func (c *Client) editManually(ourCode, theirCode []string) []string {
	fmt.Println("\n📝 MANUAL EDITING MODE")
	fmt.Println("Enter your resolved content line by line.")
	fmt.Println("Type 'END' on a line by itself to finish.")
	fmt.Println("Type 'CANCEL' to abort manual editing.")
	fmt.Println(strings.Repeat("-", 50))

	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print(">>> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "END" {
			break
		}
		if line == "CANCEL" {
			fmt.Println("❌ Manual editing canceled, keeping original conflict")
			result := make([]string, 0, len(ourCode)+len(theirCode)+3)
			result = append(result, string(ConflictStart)+"HEAD")
			result = append(result, ourCode...)
			result = append(result, string(ConflictMiddle))
			result = append(result, theirCode...)
			result = append(result, string(ConflictEnd)+"incoming")
			return result
		}

		lines = append(lines, line)
	}

	if len(lines) == 0 {
		fmt.Println("⚠️  No content entered, falling back to smart merge")
		return c.smartMergeExcelJSON(ourCode, theirCode)
	}

	fmt.Printf("✅ Manual resolution complete (%d lines)\n", len(lines))
	return lines
}

// InteractiveConflictResolver provides a more advanced interactive interface
type InteractiveConflictResolver struct {
	client *Client
}

// NewInteractiveConflictResolver creates a new interactive conflict resolver
func NewInteractiveConflictResolver(client *Client) *InteractiveConflictResolver {
	return &InteractiveConflictResolver{client: client}
}

// ResolveAllConflicts provides an interactive interface for resolving all conflicts
func (icr *InteractiveConflictResolver) ResolveAllConflicts() error {
	conflictedFiles, err := icr.client.GetConflictedFiles()
	if err != nil {
		return fmt.Errorf("failed to get conflicted files: %w", err)
	}

	if len(conflictedFiles) == 0 {
		fmt.Println("✅ No conflicts found!")
		return nil
	}

	fmt.Printf("\n🔧 Found %d conflicted file(s):\n", len(conflictedFiles))
	for i, file := range conflictedFiles {
		fmt.Printf("%d. %s\n", i+1, file)
	}

	fmt.Println("\nResolving conflicts interactively...")

	for i, file := range conflictedFiles {
		fmt.Printf("\n" + strings.Repeat("=", 80))
		fmt.Printf("RESOLVING CONFLICT %d/%d: %s", i+1, len(conflictedFiles), file)
		fmt.Printf("\n" + strings.Repeat("=", 80))

		if filepath.Ext(file) == ".json" {
			err := icr.client.ResolveConflict(file, ResolveInteractive)
			if err != nil {
				fmt.Printf("❌ Failed to resolve %s: %v\n", file, err)
				continue
			}
		} else {
			fmt.Printf("⏭️  Skipping non-JSON file: %s\n", file)
		}
	}

	fmt.Println("\n🎉 Interactive conflict resolution complete!")
	return nil
}
