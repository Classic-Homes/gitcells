package converter

import (
	"crypto/sha1" // #nosec G505 - filename disambiguation, not security
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/git"
	"github.com/Classic-Homes/gitcells/internal/utils"
	"github.com/Classic-Homes/gitcells/pkg/models"
)

// MaxChunkFileBytes caps any single JSON chunk we'll attempt to deserialize.
// A 256 MB ceiling keeps us safely above realistic per-sheet payloads while
// preventing a pathological file from forcing the whole document into memory
// (which would otherwise OOM well before json.Unmarshal completes).
const MaxChunkFileBytes int64 = 256 * 1024 * 1024

// readJSONFileBounded reads a JSON file into the given destination, refusing
// inputs larger than MaxChunkFileBytes.
func readJSONFileBounded(path string, dst interface{}) error {
	f, err := os.Open(path) // #nosec G304 - path comes from chunk metadata we wrote
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}
	if info.Size() > MaxChunkFileBytes {
		return fmt.Errorf("chunk file %s exceeds size limit (%d > %d)", path, info.Size(), MaxChunkFileBytes)
	}

	dec := json.NewDecoder(io.LimitReader(f, MaxChunkFileBytes+1))
	return dec.Decode(dst)
}

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
	Version     string   `json:"version"`
	Strategy    string   `json:"strategy"`
	MainFile    string   `json:"main_file"`
	ChunkFiles  []string `json:"chunk_files"`
	TotalSheets int      `json:"total_sheets"`
	Created     string   `json:"created"`
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
	gitRoot, err := git.FindRepositoryRoot(excelDir)
	if err != nil {
		// If not in a git repo, use the excel directory
		gitRoot = excelDir
	}

	// Calculate relative path from git root to excel file
	relPath, err := filepath.Rel(gitRoot, excelDir)
	if err != nil {
		relPath = ""
	}

	// Create the .gitcells/data directory structure mirroring the source structure
	chunkDir := filepath.Join(gitRoot, constants.GitCellsDataDir, relPath, excelFile+constants.ChunksDirSuffix)
	if err := os.MkdirAll(chunkDir, constants.DirPermissions); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", chunkDir, "failed to create chunk directory")
	}

	chunkFiles := make([]string, 0, len(doc.Sheets)+1)

	// Write main metadata file
	mainFile := filepath.Join(chunkDir, constants.WorkbookFileName)
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
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", mainFile, "failed to write main file")
	}
	chunkFiles = append(chunkFiles, mainFile)

	// Write individual sheet files
	for _, sheet := range doc.Sheets {
		sheetFile := filepath.Join(chunkDir, s.sheetChunkFilename(sheet.Name))

		// Create a document with just this sheet
		sheetDoc := &SheetChunk{
			Version:          doc.Version,
			WorkbookChecksum: doc.Metadata.Checksum,
			Sheet:            sheet,
		}

		if err := s.writeJSONFile(sheetFile, sheetDoc, options.CompactJSON); err != nil {
			return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", sheetFile, fmt.Sprintf("failed to write sheet %s", sheet.Name))
		}
		chunkFiles = append(chunkFiles, sheetFile)

		s.logger.Debugf("Wrote sheet chunk: %s (%d cells)", sheetFile, len(sheet.Cells))
	}

	// Write chunk metadata
	metadataFile := filepath.Join(chunkDir, constants.ChunkMetadataFile)
	metadata := &ChunkMetadata{
		Version:     "1.0",
		Strategy:    "sheet-based",
		MainFile:    constants.WorkbookFileName,
		ChunkFiles:  s.getRelativeChunkFiles(chunkDir, chunkFiles),
		TotalSheets: len(doc.Sheets),
		Created:     doc.Metadata.Created.Format("2006-01-02T15:04:05Z07:00"),
	}

	if err := s.writeJSONFile(metadataFile, metadata, false); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", metadataFile, "failed to write chunk metadata")
	}

	s.logger.Infof("Successfully wrote %d chunk files to %s", len(chunkFiles), chunkDir)
	return chunkFiles, nil
}

func (s *SheetBasedChunking) ReadChunks(basePath string) (*models.ExcelDocument, error) {
	// Determine where chunks are stored based on the input path
	var chunkDir string

	// If basePath is already a chunk directory, use it directly
	if strings.Contains(basePath, constants.GitCellsDataDir+string(filepath.Separator)) && strings.HasSuffix(basePath, constants.ChunksDirSuffix) {
		chunkDir = basePath
	} else {
		// Otherwise, calculate the chunk directory location
		excelDir := filepath.Dir(basePath)
		excelFile := filepath.Base(basePath)
		excelFile = strings.TrimSuffix(excelFile, ".json")
		excelFile = strings.TrimSuffix(excelFile, ".xlsx")

		gitRoot, err := git.FindRepositoryRoot(excelDir)
		if err != nil {
			// If not in a git repo, use the excel directory
			gitRoot = excelDir
		}
		relPath, err := filepath.Rel(gitRoot, excelDir)
		if err != nil {
			relPath = ""
		}

		chunkDir = filepath.Join(gitRoot, constants.GitCellsDataDir, relPath, excelFile+constants.ChunksDirSuffix)
	}

	// Read chunk metadata
	metadataFile := filepath.Join(chunkDir, constants.ChunkMetadataFile)
	var metadata ChunkMetadata
	if err := readJSONFileBounded(metadataFile, &metadata); err != nil {
		if os.IsNotExist(err) {
			return nil, utils.NewError(utils.ErrorTypeFileSystem, "ReadChunks",
				fmt.Sprintf("JSON file chunks not found. The file '%s' appears to be a standalone JSON file, but GitCells requires chunked JSON files created by the Excel to JSON conversion. Please ensure you're using a JSON file that was created by GitCells.", basePath))
		}
		return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "ReadChunks", metadataFile, "failed to read chunk metadata")
	}

	// Read main workbook file
	mainFile := filepath.Join(chunkDir, metadata.MainFile)
	var doc models.ExcelDocument
	if err := readJSONFileBounded(mainFile, &doc); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "ReadChunks", mainFile, "failed to read main file")
	}

	// Clear sheets array - we'll populate from individual files
	doc.Sheets = []models.Sheet{}

	// Read each sheet file
	for _, chunkFile := range metadata.ChunkFiles {
		if chunkFile == metadata.MainFile {
			continue // Skip main file
		}

		sheetFile := filepath.Join(chunkDir, chunkFile)
		var sheetChunk SheetChunk
		if err := readJSONFileBounded(sheetFile, &sheetChunk); err != nil {
			s.logger.Warnf("Failed to load sheet file %s: %v", sheetFile, err)
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

	if strings.Contains(basePath, constants.GitCellsDataDir+string(filepath.Separator)) && strings.HasSuffix(basePath, constants.ChunksDirSuffix) {
		chunkDir = basePath
	} else {
		excelDir := filepath.Dir(basePath)
		excelFile := filepath.Base(basePath)
		excelFile = strings.TrimSuffix(excelFile, ".json")
		excelFile = strings.TrimSuffix(excelFile, ".xlsx")

		gitRoot, err := git.FindRepositoryRoot(excelDir)
		if err != nil {
			// If not in a git repo, use the excel directory
			gitRoot = excelDir
		}
		relPath, err := filepath.Rel(gitRoot, excelDir)
		if err != nil {
			relPath = ""
		}

		chunkDir = filepath.Join(gitRoot, constants.GitCellsDataDir, relPath, excelFile+constants.ChunksDirSuffix)
	}

	// Check if chunk directory exists
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		return nil, utils.NewError(utils.ErrorTypeFileSystem, "GetChunkPaths", fmt.Sprintf("chunk directory does not exist: %s", chunkDir))
	}

	// Read chunk metadata
	metadataFile := filepath.Join(chunkDir, constants.ChunkMetadataFile)
	var metadata ChunkMetadata
	if err := readJSONFileBounded(metadataFile, &metadata); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "GetChunkPaths", metadataFile, "failed to read chunk metadata")
	}

	// Build full paths
	paths := make([]string, 0, len(metadata.ChunkFiles))
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
		return utils.WrapFileError(err, utils.ErrorTypeConverter, "writeJSONFile", path, "failed to marshal JSON")
	}

	return os.WriteFile(path, jsonData, constants.FilePermissions)
}

// filenameReplacer strips filesystem-hostile chars to underscores. Same set
// across all platforms so chunk filenames are portable Win/Mac/Linux.
var filenameReplacer = strings.NewReplacer(
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

// sheetChunkFilename builds the on-disk JSON filename for a sheet. When
// sanitization changes the name, a short content-hash suffix is appended so
// adversarial names like "A B" and "A_B" don't collide and silently
// overwrite each other. The original sheet name is preserved inside the
// chunk payload (SheetChunk.Sheet.Name), so the suffix is purely for
// filename uniqueness.
func (s *SheetBasedChunking) sheetChunkFilename(sheetName string) string {
	sanitized := filenameReplacer.Replace(sheetName)
	if sanitized == sheetName {
		return fmt.Sprintf("sheet_%s.json", sanitized)
	}
	sum := sha1.Sum([]byte(sheetName)) // #nosec G401 - collision-resistance only, not auth
	return fmt.Sprintf("sheet_%s_%s.json", sanitized, hex.EncodeToString(sum[:])[:6])
}

func (s *SheetBasedChunking) getRelativeChunkFiles(baseDir string, fullPaths []string) []string {
	relativePaths := make([]string, 0, len(fullPaths))
	for _, fullPath := range fullPaths {
		relPath, _ := filepath.Rel(baseDir, fullPath)
		relativePaths = append(relativePaths, relPath)
	}
	return relativePaths
}
