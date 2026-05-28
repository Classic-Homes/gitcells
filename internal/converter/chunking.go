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
	"time"

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

// WriteChunks writes the document to disk atomically. All chunks are
// staged into a sibling temp directory, each file fsync'd, and then the
// staging directory is swapped into place via os.Rename. If anything
// fails before the swap, the previous chunk dir (if any) is left intact;
// a crash mid-write leaves orphan staging dirs that can be safely removed.
func (s *SheetBasedChunking) WriteChunks(doc *models.ExcelDocument, basePath string, options ConvertOptions) ([]string, error) {
	excelDir := filepath.Dir(basePath)
	excelFile := filepath.Base(basePath)
	// Strip both .xlsx and .json so the chunk dir name matches what
	// resolveChunkDir computes on read. Mismatch here used to make
	// writer produce "sample.xlsx_chunks" while reader looked for
	// "sample_chunks", silently breaking round-trip via basePath.
	excelFile = strings.TrimSuffix(excelFile, ".json")
	excelFile = strings.TrimSuffix(excelFile, ".xlsx")

	gitRoot, err := git.FindRepositoryRoot(excelDir)
	if err != nil {
		gitRoot = excelDir
	}

	relPath, err := filepath.Rel(gitRoot, excelDir)
	if err != nil {
		relPath = ""
	}

	chunkDir := filepath.Join(gitRoot, constants.GitCellsDataDir, relPath, excelFile+constants.ChunksDirSuffix)
	parentDir := filepath.Dir(chunkDir)
	if err := os.MkdirAll(parentDir, constants.DirPermissions); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", parentDir, "failed to create chunk parent directory")
	}

	stagingDir := chunkDir + ".tmp." + uniqueSuffix()
	if err := os.MkdirAll(stagingDir, constants.DirPermissions); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", stagingDir, "failed to create staging directory")
	}

	committed := false
	defer func() {
		if !committed {
			if rmErr := os.RemoveAll(stagingDir); rmErr != nil {
				s.logger.Warnf("Failed to clean up staging dir %s: %v", stagingDir, rmErr)
			}
		}
	}()

	stagedFiles := make([]string, 0, len(doc.Sheets)+1)

	mainStaged := filepath.Join(stagingDir, constants.WorkbookFileName)
	mainDoc := &models.ExcelDocument{
		Version:      doc.Version,
		Metadata:     doc.Metadata,
		DefinedNames: doc.DefinedNames,
		Properties:   doc.Properties,
		Sheets:       []models.Sheet{},
	}
	for _, sheet := range doc.Sheets {
		mainDoc.Sheets = append(mainDoc.Sheets, models.Sheet{
			Name:   sheet.Name,
			Index:  sheet.Index,
			Hidden: sheet.Hidden,
		})
	}
	if err := s.writeJSONFile(mainStaged, mainDoc, options.CompactJSON); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", mainStaged, "failed to write main file")
	}
	stagedFiles = append(stagedFiles, mainStaged)

	// Pre-scan sheet names so case-insensitive filesystems (macOS APFS,
	// Windows NTFS) can't silently clobber one chunk file with another.
	// Two sheets named "Sheet1" and "sheet1" both sanitize to themselves
	// — without this map they'd hit the same lowercased path on disk.
	// Force-hash *every* member of a colliding group so the disambiguation
	// is symmetric and the JSON stays byte-stable on subsequent writes.
	caseInsensitiveCount := make(map[string]int, len(doc.Sheets))
	for _, sheet := range doc.Sheets {
		key := strings.ToLower(filenameReplacer.Replace(sheet.Name))
		caseInsensitiveCount[key]++
	}

	for _, sheet := range doc.Sheets {
		key := strings.ToLower(filenameReplacer.Replace(sheet.Name))
		forceHash := caseInsensitiveCount[key] > 1
		sheetFile := filepath.Join(stagingDir, s.sheetChunkFilename(sheet.Name, forceHash))
		sheetDoc := &SheetChunk{
			Version:          doc.Version,
			WorkbookChecksum: doc.Metadata.Checksum,
			Sheet:            sheet,
		}
		if err := s.writeJSONFile(sheetFile, sheetDoc, options.CompactJSON); err != nil {
			return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", sheetFile, fmt.Sprintf("failed to write sheet %s", sheet.Name))
		}
		stagedFiles = append(stagedFiles, sheetFile)
		s.logger.Debugf("Staged sheet chunk: %s (%d cells)", sheetFile, len(sheet.Cells))
	}

	metadataFile := filepath.Join(stagingDir, constants.ChunkMetadataFile)
	metadata := &ChunkMetadata{
		Version:     "1.0",
		Strategy:    "sheet-based",
		MainFile:    constants.WorkbookFileName,
		ChunkFiles:  s.getRelativeChunkFiles(stagingDir, stagedFiles),
		TotalSheets: len(doc.Sheets),
		Created:     doc.Metadata.Created.Format("2006-01-02T15:04:05Z07:00"),
	}
	if err := s.writeJSONFile(metadataFile, metadata, false); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", metadataFile, "failed to write chunk metadata")
	}

	if err := atomicReplaceDir(stagingDir, chunkDir, s.logger); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeFileSystem, "WriteChunks", chunkDir, "failed to install chunk directory")
	}
	committed = true

	finalFiles := make([]string, 0, len(stagedFiles))
	for _, p := range stagedFiles {
		rel, err := filepath.Rel(stagingDir, p)
		if err != nil {
			rel = filepath.Base(p)
		}
		finalFiles = append(finalFiles, filepath.Join(chunkDir, rel))
	}

	s.logger.Infof("Successfully wrote %d chunk files to %s", len(finalFiles), chunkDir)
	return finalFiles, nil
}

// uniqueSuffix returns a short string guaranteed unique within a single
// process across rapid calls. Used for staging-dir naming.
func uniqueSuffix() string {
	return fmt.Sprintf("%d.%d", os.Getpid(), time.Now().UnixNano())
}

// atomicReplaceDir installs src as dst. If dst doesn't exist, this is a
// single rename. If dst exists, the existing directory is renamed aside
// first, then src is renamed into place; on success the backup is
// removed (best-effort), on failure the backup is restored.
//
// os.Rename over an existing directory fails on Windows, so the
// rename-aside dance is required for cross-platform correctness.
func atomicReplaceDir(src, dst string, logger Logger) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		return os.Rename(src, dst)
	} else if err != nil {
		return err
	}

	backup := dst + ".old." + uniqueSuffix()
	if err := os.Rename(dst, backup); err != nil {
		return fmt.Errorf("rename existing chunks aside: %w", err)
	}
	if err := os.Rename(src, dst); err != nil {
		if restoreErr := os.Rename(backup, dst); restoreErr != nil {
			return fmt.Errorf("install new chunks (%v); restore backup also failed (%v)", err, restoreErr)
		}
		return fmt.Errorf("install new chunks: %w", err)
	}
	if rmErr := os.RemoveAll(backup); rmErr != nil {
		logger.Warnf("Failed to remove backup chunk dir %s: %v", backup, rmErr)
	}
	return nil
}

func (s *SheetBasedChunking) ReadChunks(basePath string) (*models.ExcelDocument, error) {
	chunkDir := resolveChunkDir(basePath)

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

	// Pre-flight: confirm every chunk file in metadata exists before we
	// begin reading. Previously a missing file produced a Warnf + continue,
	// silently returning a partial workbook with fewer sheets than
	// metadata.TotalSheets claimed. That is data-loss masquerading as
	// success. A missing chunk is unambiguous corruption — fail loudly so
	// the caller can refuse to write back a truncated .xlsx.
	for _, chunkFile := range metadata.ChunkFiles {
		path := filepath.Join(chunkDir, filepath.FromSlash(chunkFile))
		if _, err := os.Stat(path); err != nil {
			return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "ReadChunks", path,
				"chunk file listed in metadata is missing or unreadable")
		}
	}

	// Read main workbook file
	mainFile := filepath.Join(chunkDir, metadata.MainFile)
	var doc models.ExcelDocument
	if err := readJSONFileBounded(mainFile, &doc); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "ReadChunks", mainFile, "failed to read main file")
	}

	// Clear sheets array - we'll populate from individual files
	doc.Sheets = []models.Sheet{}

	// Read each sheet file. Any read failure here is fatal — same
	// reasoning as the pre-flight stat: a corrupted chunk file means
	// the rebuilt workbook would be missing data without the user
	// knowing.
	for _, chunkFile := range metadata.ChunkFiles {
		if chunkFile == metadata.MainFile {
			continue // Skip main file
		}

		sheetFile := filepath.Join(chunkDir, filepath.FromSlash(chunkFile))
		var sheetChunk SheetChunk
		if err := readJSONFileBounded(sheetFile, &sheetChunk); err != nil {
			return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "ReadChunks", sheetFile,
				"failed to read sheet chunk")
		}

		doc.Sheets = append(doc.Sheets, sheetChunk.Sheet)
		s.logger.Debugf("Loaded sheet %s with %d cells", sheetChunk.Sheet.Name, len(sheetChunk.Sheet.Cells))
	}

	s.logger.Infof("Successfully read %d sheets from chunks", len(doc.Sheets))
	return &doc, nil
}

func (s *SheetBasedChunking) GetChunkPaths(basePath string) ([]string, error) {
	chunkDir := resolveChunkDir(basePath)

	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		return nil, utils.NewError(utils.ErrorTypeFileSystem, "GetChunkPaths", fmt.Sprintf("chunk directory does not exist: %s", chunkDir))
	}

	metadataFile := filepath.Join(chunkDir, constants.ChunkMetadataFile)
	var metadata ChunkMetadata
	if err := readJSONFileBounded(metadataFile, &metadata); err != nil {
		return nil, utils.WrapFileError(err, utils.ErrorTypeConverter, "GetChunkPaths", metadataFile, "failed to read chunk metadata")
	}

	// Include the metadata file itself. It changes on every conversion
	// (Created timestamp comes from doc.Metadata) and must be committed
	// alongside its content files; if it isn't, git ends up with content
	// chunks whose sibling metadata entry is stale.
	paths := make([]string, 0, len(metadata.ChunkFiles)+1)
	paths = append(paths, metadataFile)
	for _, chunkFile := range metadata.ChunkFiles {
		paths = append(paths, filepath.Join(chunkDir, filepath.FromSlash(chunkFile)))
	}

	return paths, nil
}

// resolveChunkDir returns the on-disk chunk directory for a basePath.
// basePath may point at either:
//   - an existing chunk directory (ends in _chunks/ and contains the
//     metadata file), in which case it is returned as-is, or
//   - the source .xlsx file, in which case the chunk dir is computed
//     relative to the enclosing git root (or the excel dir if not in
//     a repo).
//
// The structural check (suffix + metadata stat) replaces a brittle
// substring match on .gitcells/data, which would not match on Windows
// where paths use backslashes.
func resolveChunkDir(basePath string) string {
	if strings.HasSuffix(basePath, constants.ChunksDirSuffix) {
		if _, err := os.Stat(filepath.Join(basePath, constants.ChunkMetadataFile)); err == nil {
			return basePath
		}
	}

	excelDir := filepath.Dir(basePath)
	excelFile := filepath.Base(basePath)
	excelFile = strings.TrimSuffix(excelFile, ".json")
	excelFile = strings.TrimSuffix(excelFile, ".xlsx")

	gitRoot, err := git.FindRepositoryRoot(excelDir)
	if err != nil {
		gitRoot = excelDir
	}
	relPath, err := filepath.Rel(gitRoot, excelDir)
	if err != nil {
		relPath = ""
	}

	return filepath.Join(gitRoot, constants.GitCellsDataDir, relPath, excelFile+constants.ChunksDirSuffix)
}

// SheetChunk represents a single sheet in a separate file
type SheetChunk struct {
	Version          string       `json:"version"`
	WorkbookChecksum string       `json:"workbook_checksum"`
	Sheet            models.Sheet `json:"sheet"`
}

// Helper methods

// writeJSONFile marshals data to JSON and writes it to path with an
// explicit fsync. The caller is responsible for placing path in a
// staging directory; this function does not rename or move anything.
// Fsync matters because a power loss between write and rename would
// otherwise leave a zero-length file even if the rename appears to
// have succeeded (the directory entry is durable but the file content isn't).
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

	// json.Marshal / MarshalIndent never append a trailing newline. Without
	// one, every `git diff` of a chunk file is flagged "No newline at end
	// of file" and many text tools (cat, less, $EDITOR's syntax highlighter)
	// render the last line oddly. Appending here is cheap and makes chunk
	// files behave like every other text artifact in a repo.
	if len(jsonData) == 0 || jsonData[len(jsonData)-1] != '\n' {
		jsonData = append(jsonData, '\n')
	}

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constants.FilePermissions) // #nosec G304 - path under our staging dir
	if err != nil {
		return err
	}
	if _, err := f.Write(jsonData); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
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

// sheetChunkFilename builds the on-disk JSON filename for a sheet.
//
// A short content-hash suffix is appended when either:
//   - sanitization changed the name (adversarial collision: "A B" and
//     "A_B" both sanitize to "A_B" so the literal one stays bare and the
//     sanitized one carries a hash, keeping them distinct); or
//   - the caller flags a case-insensitive collision with another sheet in
//     the same workbook (forceHash=true). On macOS APFS and Windows NTFS,
//     "Sheet1" and "sheet1" hit the same path on disk; the hash gives both
//     sides a unique filename without depending on filesystem case
//     sensitivity.
//
// The original sheet name is preserved inside the chunk payload
// (SheetChunk.Sheet.Name), so the suffix is purely for filename uniqueness.
func (s *SheetBasedChunking) sheetChunkFilename(sheetName string, forceHash bool) string {
	sanitized := filenameReplacer.Replace(sheetName)
	if sanitized == sheetName && !forceHash {
		return fmt.Sprintf("sheet_%s.json", sanitized)
	}
	sum := sha1.Sum([]byte(sheetName)) // #nosec G401 - collision-resistance only, not auth
	return fmt.Sprintf("sheet_%s_%s.json", sanitized, hex.EncodeToString(sum[:])[:6])
}

// getRelativeChunkFiles returns paths relative to baseDir, normalized to
// forward slashes so the metadata.json is identical regardless of which
// OS wrote it. The reader applies filepath.FromSlash to convert back.
func (s *SheetBasedChunking) getRelativeChunkFiles(baseDir string, fullPaths []string) []string {
	relativePaths := make([]string, 0, len(fullPaths))
	for _, fullPath := range fullPaths {
		relPath, err := filepath.Rel(baseDir, fullPath)
		if err != nil {
			relPath = filepath.Base(fullPath)
		}
		relativePaths = append(relativePaths, filepath.ToSlash(relPath))
	}
	return relativePaths
}

// DefaultOrphanMaxAge is the threshold below which a staging directory
// is assumed to belong to an in-flight write and left alone. One hour
// is comfortably longer than any plausible Excel conversion.
const DefaultOrphanMaxAge = 1 * time.Hour

// SweepOrphans removes leftover staging (`*.tmp.*`) and backup
// (`*.old.*`) directories under root whose mtime is older than maxAge.
// Both patterns are produced by WriteChunks / atomicReplaceDir and
// would otherwise survive a crash or SIGKILL.
//
// This is best-effort hygiene: any failure to scan or remove is logged
// at warn level but never returned, so a permissions glitch on one stale
// dir doesn't block the rest of the pipeline.
func SweepOrphans(root string, maxAge time.Duration, logger Logger) {
	if root == "" {
		return
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return
	}

	cutoff := time.Now().Add(-maxAge)
	walkErr := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return nil // continue walking siblings; one bad subtree shouldn't abort
		}
		if !info.IsDir() || path == root {
			return nil
		}
		base := filepath.Base(path)
		if !isOrphanStagingName(base) {
			return nil
		}
		if info.ModTime().After(cutoff) {
			return filepath.SkipDir
		}
		if rmErr := os.RemoveAll(path); rmErr != nil {
			logger.Warnf("Failed to remove orphan staging dir %s: %v", path, rmErr)
		} else {
			logger.Debugf("Removed orphan staging dir: %s", path)
		}
		return filepath.SkipDir
	})
	if walkErr != nil {
		logger.Warnf("Orphan sweep walk error under %s: %v", root, walkErr)
	}
}

// isOrphanStagingName reports whether basename matches the staging or
// backup naming convention used by WriteChunks and atomicReplaceDir
// (suffixes of the form `.tmp.<pid>.<nanos>` or `.old.<pid>.<nanos>`).
func isOrphanStagingName(base string) bool {
	return strings.Contains(base, ".tmp.") || strings.Contains(base, ".old.")
}
