package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/pkg/models"
)

// ChunkingStrategy defines how to split Excel data into multiple files
type ChunkingStrategy interface {
	// WriteChunks writes the document as multiple JSON files
	WriteChunks(doc *models.ExcelDocument, basePath string, options ConvertOptions) ([]string, error)
	
	// ReadChunks reads multiple JSON files back into a document
	ReadChunks(basePath string) (*models.ExcelDocument, error)
	
	// GetChunkPaths returns expected chunk file paths for a given base path
	GetChunkPaths(basePath string) ([]string, error)
}

// ChunkMetadata stores information about chunked files
type ChunkMetadata struct {
	Version      string   `json:"version"`
	Strategy     string   `json:"strategy"`
	MainFile     string   `json:"main_file"`
	ChunkFiles   []string `json:"chunk_files"`
	TotalSheets  int      `json:"total_sheets"`
	Created      string   `json:"created"`
}

// SheetBasedChunking implements sheet-level file splitting
type SheetBasedChunking struct {
	logger Logger
}

// Logger interface to avoid circular import
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

func NewSheetBasedChunking(logger Logger) ChunkingStrategy {
	return &SheetBasedChunking{
		logger: logger,
	}
}


func (s *SheetBasedChunking) WriteChunks(doc *models.ExcelDocument, basePath string, options ConvertOptions) ([]string, error) {
	// Determine the root directory and relative path for the Excel file
	excelDir := filepath.Dir(basePath)
	excelFile := filepath.Base(basePath)
	
	// Remove .json extension if present
	excelFile = strings.TrimSuffix(excelFile, ".json")
	
	// Find the git root or use current directory
	gitRoot := s.findGitRoot(excelDir)
	
	// Calculate relative path from git root to excel file
	relPath, err := filepath.Rel(gitRoot, excelDir)
	if err != nil {
		relPath = ""
	}
	
	// Create the .gitcells/data directory structure mirroring the source structure
	chunkDir := filepath.Join(gitRoot, ".gitcells", "data", relPath, excelFile+"_chunks")
	if err := os.MkdirAll(chunkDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create chunk directory: %w", err)
	}
	
	var chunkFiles []string
	
	// Write main metadata file
	mainFile := filepath.Join(chunkDir, "workbook.json")
	mainDoc := &models.ExcelDocument{
		Version:      doc.Version,
		Metadata:     doc.Metadata,
		DefinedNames: doc.DefinedNames,
		Properties:   doc.Properties,
		Sheets:       []models.Sheet{}, // Empty sheets, just metadata
	}
	
	// Add sheet references to main doc
	for _, sheet := range doc.Sheets {
		// Include only essential sheet metadata
		sheetRef := models.Sheet{
			Name:   sheet.Name,
			Index:  sheet.Index,
			Hidden: sheet.Hidden,
		}
		mainDoc.Sheets = append(mainDoc.Sheets, sheetRef)
	}
	
	// Write main file
	if err := s.writeJSONFile(mainFile, mainDoc, options.CompactJSON); err != nil {
		return nil, fmt.Errorf("failed to write main file: %w", err)
	}
	chunkFiles = append(chunkFiles, mainFile)
	
	// Write individual sheet files
	for _, sheet := range doc.Sheets {
		sheetFile := filepath.Join(chunkDir, s.sanitizeFilename(fmt.Sprintf("sheet_%s.json", sheet.Name)))
		
		// Create a document with just this sheet
		sheetDoc := &SheetChunk{
			Version:  doc.Version,
			WorkbookChecksum: doc.Metadata.Checksum,
			Sheet:    sheet,
		}
		
		if err := s.writeJSONFile(sheetFile, sheetDoc, options.CompactJSON); err != nil {
			return nil, fmt.Errorf("failed to write sheet %s: %w", sheet.Name, err)
		}
		chunkFiles = append(chunkFiles, sheetFile)
		
		s.logger.Debugf("Wrote sheet chunk: %s (%d cells)", sheetFile, len(sheet.Cells))
	}
	
	// Write chunk metadata
	metadataFile := filepath.Join(chunkDir, ".gitcells_chunks.json")
	metadata := &ChunkMetadata{
		Version:     "1.0",
		Strategy:    "sheet-based",
		MainFile:    "workbook.json",
		ChunkFiles:  s.getRelativeChunkFiles(chunkDir, chunkFiles),
		TotalSheets: len(doc.Sheets),
		Created:     doc.Metadata.Created.Format("2006-01-02T15:04:05Z07:00"),
	}
	
	if err := s.writeJSONFile(metadataFile, metadata, false); err != nil {
		return nil, fmt.Errorf("failed to write chunk metadata: %w", err)
	}
	
	s.logger.Infof("Successfully wrote %d chunk files to %s", len(chunkFiles), chunkDir)
	return chunkFiles, nil
}

func (s *SheetBasedChunking) ReadChunks(basePath string) (*models.ExcelDocument, error) {
	// Determine where chunks are stored based on the input path
	var chunkDir string
	
	// If basePath is already a chunk directory, use it directly
	if strings.Contains(basePath, ".gitcells/data/") && strings.HasSuffix(basePath, "_chunks") {
		chunkDir = basePath
	} else {
		// Otherwise, calculate the chunk directory location
		excelDir := filepath.Dir(basePath)
		excelFile := filepath.Base(basePath)
		excelFile = strings.TrimSuffix(excelFile, ".json")
		excelFile = strings.TrimSuffix(excelFile, ".xlsx")
		
		gitRoot := s.findGitRoot(excelDir)
		relPath, err := filepath.Rel(gitRoot, excelDir)
		if err != nil {
			relPath = ""
		}
		
		chunkDir = filepath.Join(gitRoot, ".gitcells", "data", relPath, excelFile+"_chunks")
	}
	
	// Read chunk metadata
	metadataFile := filepath.Join(chunkDir, ".gitcells_chunks.json")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk metadata: %w", err)
	}
	
	var metadata ChunkMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse chunk metadata: %w", err)
	}
	
	// Read main workbook file
	mainFile := filepath.Join(chunkDir, metadata.MainFile)
	mainData, err := os.ReadFile(mainFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read main file: %w", err)
	}
	
	var doc models.ExcelDocument
	if err := json.Unmarshal(mainData, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse main file: %w", err)
	}
	
	// Clear sheets array - we'll populate from individual files
	doc.Sheets = []models.Sheet{}
	
	// Read each sheet file
	for _, chunkFile := range metadata.ChunkFiles {
		if chunkFile == metadata.MainFile {
			continue // Skip main file
		}
		
		sheetFile := filepath.Join(chunkDir, chunkFile)
		sheetData, err := os.ReadFile(sheetFile)
		if err != nil {
			s.logger.Warnf("Failed to read sheet file %s: %v", sheetFile, err)
			continue
		}
		
		var sheetChunk SheetChunk
		if err := json.Unmarshal(sheetData, &sheetChunk); err != nil {
			s.logger.Warnf("Failed to parse sheet file %s: %v", sheetFile, err)
			continue
		}
		
		doc.Sheets = append(doc.Sheets, sheetChunk.Sheet)
		s.logger.Debugf("Loaded sheet %s with %d cells", sheetChunk.Sheet.Name, len(sheetChunk.Sheet.Cells))
	}
	
	s.logger.Infof("Successfully read %d sheets from chunks", len(doc.Sheets))
	return &doc, nil
}

func (s *SheetBasedChunking) GetChunkPaths(basePath string) ([]string, error) {
	// Determine where chunks are stored
	var chunkDir string
	
	if strings.Contains(basePath, ".gitcells/data/") && strings.HasSuffix(basePath, "_chunks") {
		chunkDir = basePath
	} else {
		excelDir := filepath.Dir(basePath)
		excelFile := filepath.Base(basePath)
		excelFile = strings.TrimSuffix(excelFile, ".json")
		excelFile = strings.TrimSuffix(excelFile, ".xlsx")
		
		gitRoot := s.findGitRoot(excelDir)
		relPath, err := filepath.Rel(gitRoot, excelDir)
		if err != nil {
			relPath = ""
		}
		
		chunkDir = filepath.Join(gitRoot, ".gitcells", "data", relPath, excelFile+"_chunks")
	}
	
	// Check if chunk directory exists
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("chunk directory does not exist: %s", chunkDir)
	}
	
	// Read chunk metadata
	metadataFile := filepath.Join(chunkDir, ".gitcells_chunks.json")
	metadataData, err := os.ReadFile(metadataFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunk metadata: %w", err)
	}
	
	var metadata ChunkMetadata
	if err := json.Unmarshal(metadataData, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse chunk metadata: %w", err)
	}
	
	// Build full paths
	var paths []string
	for _, chunkFile := range metadata.ChunkFiles {
		paths = append(paths, filepath.Join(chunkDir, chunkFile))
	}
	
	return paths, nil
}

// SheetChunk represents a single sheet in a separate file
type SheetChunk struct {
	Version          string       `json:"version"`
	WorkbookChecksum string       `json:"workbook_checksum"`
	Sheet            models.Sheet `json:"sheet"`
}

// Helper methods

func (s *SheetBasedChunking) writeJSONFile(path string, data interface{}, compact bool) error {
	var jsonData []byte
	var err error
	
	if compact {
		jsonData, err = json.Marshal(data)
	} else {
		jsonData, err = json.MarshalIndent(data, "", "  ")
	}
	
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return os.WriteFile(path, jsonData, 0644)
}

func (s *SheetBasedChunking) sanitizeFilename(name string) string {
	// Replace invalid filename characters
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}

func (s *SheetBasedChunking) getRelativeChunkFiles(baseDir string, fullPaths []string) []string {
	var relativePaths []string
	for _, fullPath := range fullPaths {
		relPath, _ := filepath.Rel(baseDir, fullPath)
		relativePaths = append(relativePaths, relPath)
	}
	return relativePaths
}

// findGitRoot finds the git repository root, or returns the current directory
func (s *SheetBasedChunking) findGitRoot(startDir string) string {
	dir := startDir
	for {
		// Check if .git directory exists
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir
		}
		
		// Check if we've reached the root
		parent := filepath.Dir(dir)
		if parent == dir {
			// Return the original directory if no git root found
			return startDir
		}
		dir = parent
	}
}

// Future-proofing for hybrid chunking strategy
type HybridChunking struct {
	sheetBased      *SheetBasedChunking
	maxCellsPerFile int
	logger          Logger
}

// This is a placeholder for future implementation
func NewHybridChunking(logger Logger, maxCellsPerFile int) ChunkingStrategy {
	return &HybridChunking{
		sheetBased:      &SheetBasedChunking{logger: logger},
		maxCellsPerFile: maxCellsPerFile,
		logger:          logger,
	}
}


func (h *HybridChunking) WriteChunks(doc *models.ExcelDocument, basePath string, options ConvertOptions) ([]string, error) {
	// Future implementation will split large sheets into ranges
	// For now, delegate to sheet-based chunking
	return nil, fmt.Errorf("hybrid chunking not yet implemented")
}

func (h *HybridChunking) ReadChunks(basePath string) (*models.ExcelDocument, error) {
	// Future implementation
	return nil, fmt.Errorf("hybrid chunking not yet implemented")
}

func (h *HybridChunking) GetChunkPaths(basePath string) ([]string, error) {
	// Future implementation
	return nil, fmt.Errorf("hybrid chunking not yet implemented")
}