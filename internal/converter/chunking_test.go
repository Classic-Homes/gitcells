package converter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSheetBasedChunking(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	chunker := NewSheetBasedChunking(logger)

	// Create test document
	doc := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			AppVersion:   "test-1.0",
			OriginalFile: "test.xlsx",
			FileSize:     1024,
			Checksum:     "test-checksum",
		},
		Sheets: []models.Sheet{
			{
				Name:  "Sheet1",
				Index: 0,
				Cells: map[string]models.Cell{
					"A1": {Value: "Header", Type: models.CellTypeString},
					"A2": {Value: 123.45, Type: models.CellTypeNumber},
					"B1": {Value: "=A2*2", Formula: "=A2*2", Type: models.CellTypeFormula},
				},
			},
			{
				Name:  "Sheet2",
				Index: 1,
				Cells: map[string]models.Cell{
					"A1": {Value: "Data", Type: models.CellTypeString},
					"B2": {Value: true, Type: models.CellTypeBoolean},
				},
			},
		},
		DefinedNames: []models.DefinedName{
			{Name: "TestName", RefersTo: "Sheet1!A1:B2"},
		},
	}

	t.Run("WriteAndReadChunks", func(t *testing.T) {
		// Create temp directory with .git to simulate git repo
		tempDir := t.TempDir()
		gitDir := filepath.Join(tempDir, ".git")
		err := os.Mkdir(gitDir, constants.DirPermissions)
		require.NoError(t, err)

		basePath := filepath.Join(tempDir, "test_workbook.json")

		// Write chunks
		opts := ConvertOptions{
			CompactJSON: false,
		}

		files, err := chunker.WriteChunks(doc, basePath, opts)
		require.NoError(t, err)
		assert.Len(t, files, 3) // workbook.json + 2 sheet files

		// Verify chunk directory exists in .gitcells/data
		expectedChunkDir := filepath.Join(tempDir, ".gitcells", "data", "test_workbook_chunks")
		_, err = os.Stat(expectedChunkDir)
		assert.NoError(t, err)

		// Verify metadata file exists
		metadataPath := filepath.Join(expectedChunkDir, ".gitcells_chunks.json")
		_, err = os.Stat(metadataPath)
		assert.NoError(t, err)

		// Read chunks back
		readDoc, err := chunker.ReadChunks(basePath)
		require.NoError(t, err)

		// Verify document integrity
		assert.Equal(t, doc.Version, readDoc.Version)
		assert.Equal(t, doc.Metadata.AppVersion, readDoc.Metadata.AppVersion)
		assert.Len(t, readDoc.Sheets, 2)

		// Verify sheet data
		assert.Equal(t, "Sheet1", readDoc.Sheets[0].Name)
		assert.Len(t, readDoc.Sheets[0].Cells, 3)
		assert.Equal(t, "Header", readDoc.Sheets[0].Cells["A1"].Value)
		assert.Equal(t, 123.45, readDoc.Sheets[0].Cells["A2"].Value)
		assert.Equal(t, "=A2*2", readDoc.Sheets[0].Cells["B1"].Formula)

		assert.Equal(t, "Sheet2", readDoc.Sheets[1].Name)
		assert.Len(t, readDoc.Sheets[1].Cells, 2)
	})

	t.Run("GetChunkPaths", func(t *testing.T) {
		// Create temp directory with .git
		tempDir := t.TempDir()
		gitDir := filepath.Join(tempDir, ".git")
		err := os.Mkdir(gitDir, constants.DirPermissions)
		require.NoError(t, err)

		basePath := filepath.Join(tempDir, "test_workbook.json")

		// Write chunks first
		opts := ConvertOptions{}
		_, err = chunker.WriteChunks(doc, basePath, opts)
		require.NoError(t, err)

		// Get chunk paths
		paths, err := chunker.GetChunkPaths(basePath)
		require.NoError(t, err)
		// metadata + workbook + 2 sheets. The metadata file is part of
		// the returned set because it must be committed alongside its
		// content files (see GetChunkPaths comment).
		assert.Len(t, paths, 4)

		// Verify all paths exist
		for _, path := range paths {
			_, err := os.Stat(path)
			assert.NoError(t, err, "Path should exist: %s", path)
		}

		// Confirm the metadata file is in the returned list — the
		// pre-fix behavior omitted it, leading to silent commit
		// inconsistency.
		var hasMetadata bool
		for _, p := range paths {
			if strings.HasSuffix(p, constants.ChunkMetadataFile) {
				hasMetadata = true
				break
			}
		}
		assert.True(t, hasMetadata, "GetChunkPaths must include the chunk metadata file")
	})

	t.Run("SheetChunkFilename", func(t *testing.T) {
		c := chunker.(*SheetBasedChunking)

		// Clean names get no hash suffix.
		assert.Equal(t, "sheet_Sheet1.json", c.sheetChunkFilename("Sheet1"))

		// Dirty names get sanitized + 6-hex-char content hash to avoid collisions.
		for _, input := range []string{
			"Sheet/With/Slashes",
			"Sheet:With:Colons",
			"Sheet With Spaces",
			"Sheet*With?Chars",
		} {
			out := c.sheetChunkFilename(input)
			assert.True(t, strings.HasPrefix(out, "sheet_"), "prefix for %q: %s", input, out)
			assert.True(t, strings.HasSuffix(out, ".json"), "suffix for %q: %s", input, out)
			assert.NotContains(t, out, "/", "no slashes for %q: %s", input, out)
			assert.NotContains(t, out, " ", "no spaces for %q: %s", input, out)
		}

		// Adversarial collision: "A B" and "A_B" both sanitize to "A_B" but
		// must produce distinct filenames thanks to the hash suffix.
		a := c.sheetChunkFilename("A B")
		b := c.sheetChunkFilename("A_B")
		assert.NotEqual(t, a, b, "collision: %q vs %q both → %s", "A B", "A_B", a)
	})

	t.Run("AtomicReplaceLeavesNoStagingDirs", func(t *testing.T) {
		// Re-writing a chunk dir must replace it atomically. After two
		// successive WriteChunks, only the final chunk dir should remain in
		// .gitcells/data — no leftover .tmp.* or .old.* siblings.
		tempDir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(tempDir, ".git"), constants.DirPermissions))
		basePath := filepath.Join(tempDir, "atomic.json")
		opts := ConvertOptions{}

		_, err := chunker.WriteChunks(doc, basePath, opts)
		require.NoError(t, err)
		_, err = chunker.WriteChunks(doc, basePath, opts)
		require.NoError(t, err)

		dataDir := filepath.Join(tempDir, ".gitcells", "data")
		entries, err := os.ReadDir(dataDir)
		require.NoError(t, err)
		for _, e := range entries {
			assert.NotContains(t, e.Name(), ".tmp.", "leftover staging dir: %s", e.Name())
			assert.NotContains(t, e.Name(), ".old.", "leftover backup dir: %s", e.Name())
		}
	})

	t.Run("MetadataPathsAreForwardSlash", func(t *testing.T) {
		// Chunk metadata must store forward-slash paths so a repo written
		// on Windows can be opened on macOS/Linux without churn.
		tempDir := t.TempDir()
		require.NoError(t, os.Mkdir(filepath.Join(tempDir, ".git"), constants.DirPermissions))
		basePath := filepath.Join(tempDir, "portable.json")
		opts := ConvertOptions{}
		_, err := chunker.WriteChunks(doc, basePath, opts)
		require.NoError(t, err)

		metadataPath := filepath.Join(tempDir, ".gitcells", "data", "portable_chunks", ".gitcells_chunks.json")
		raw, err := os.ReadFile(metadataPath) // #nosec G304 - test file path
		require.NoError(t, err)
		var meta ChunkMetadata
		require.NoError(t, json.Unmarshal(raw, &meta))
		for _, p := range meta.ChunkFiles {
			assert.NotContains(t, p, "\\", "metadata path must use forward slashes: %s", p)
		}
	})
}

func TestChunkMetadata(t *testing.T) {
	metadata := &ChunkMetadata{
		Version:     "1.0",
		Strategy:    "sheet-based",
		MainFile:    "workbook.json",
		ChunkFiles:  []string{"workbook.json", "sheet_Sheet1.json", "sheet_Sheet2.json"},
		TotalSheets: 2,
		Created:     "2024-01-01T00:00:00Z",
	}

	// Test JSON marshaling
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)

	// Test JSON unmarshaling
	var loaded ChunkMetadata
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, metadata.Version, loaded.Version)
	assert.Equal(t, metadata.Strategy, loaded.Strategy)
	assert.Equal(t, metadata.MainFile, loaded.MainFile)
	assert.Len(t, loaded.ChunkFiles, 3)
	assert.Equal(t, metadata.TotalSheets, loaded.TotalSheets)
}

func TestSheetChunk(t *testing.T) {
	chunk := &SheetChunk{
		Version:          "1.0",
		WorkbookChecksum: "test-checksum",
		Sheet: models.Sheet{
			Name:  "TestSheet",
			Index: 0,
			Cells: map[string]models.Cell{
				"A1": {Value: "Test", Type: models.CellTypeString},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.MarshalIndent(chunk, "", "  ")
	require.NoError(t, err)

	// Test JSON unmarshaling
	var loaded SheetChunk
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, chunk.Version, loaded.Version)
	assert.Equal(t, chunk.WorkbookChecksum, loaded.WorkbookChecksum)
	assert.Equal(t, chunk.Sheet.Name, loaded.Sheet.Name)
	assert.Len(t, loaded.Sheet.Cells, 1)
}

// TestReadChunks_ForwardSlashMetadataIsPortable confirms the cross-OS
// contract documented in CLAUDE.md: chunk metadata always stores paths
// with forward slashes regardless of writer OS, and the reader applies
// filepath.FromSlash before joining. Without this contract, a chunk
// directory authored on Linux would fail to read on Windows.
//
// We hand-build a chunk dir whose .gitcells_chunks.json embeds explicit
// forward-slash paths, then ReadChunks it and assert the sheet payload
// loaded. The test passes identically on Windows, macOS, and Linux.
func TestReadChunks_ForwardSlashMetadataIsPortable(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	chunker := NewSheetBasedChunking(logger)

	tempDir := t.TempDir()
	chunkDir := filepath.Join(tempDir, "fixture"+constants.ChunksDirSuffix)
	require.NoError(t, os.MkdirAll(chunkDir, constants.DirPermissions))

	writeJSON := func(name string, v interface{}) {
		data, err := json.MarshalIndent(v, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(chunkDir, name), data, constants.FilePermissions))
	}

	mainDoc := &models.ExcelDocument{
		Version: "1.0",
		Metadata: models.DocumentMetadata{
			OriginalFile: "fixture.xlsx",
			Checksum:     "cross-os",
		},
		Sheets: []models.Sheet{{Name: "Foo", Index: 0}},
	}
	writeJSON(constants.WorkbookFileName, mainDoc)

	sheetChunk := &SheetChunk{
		Version:          "1.0",
		WorkbookChecksum: "cross-os",
		Sheet: models.Sheet{
			Name:  "Foo",
			Index: 0,
			Cells: map[string]models.Cell{
				"A1": {Value: "portable", Type: models.CellTypeString},
			},
		},
	}
	writeJSON("sheet_Foo.json", sheetChunk)

	// Explicit forward-slash paths in the metadata. On Windows, ReadChunks
	// must apply filepath.FromSlash so the join produces a usable path.
	metadata := &ChunkMetadata{
		Version:     "1.0",
		Strategy:    "sheet-based",
		MainFile:    constants.WorkbookFileName,
		ChunkFiles:  []string{"workbook.json", "sheet_Foo.json"},
		TotalSheets: 1,
		Created:     "2024-01-01T00:00:00Z",
	}
	writeJSON(constants.ChunkMetadataFile, metadata)

	doc, err := chunker.ReadChunks(chunkDir)
	require.NoError(t, err)
	require.Len(t, doc.Sheets, 1)
	assert.Equal(t, "Foo", doc.Sheets[0].Name)
	require.Contains(t, doc.Sheets[0].Cells, "A1")
	assert.Equal(t, "portable", doc.Sheets[0].Cells["A1"].Value)

	// Also assert the metadata on disk has literal forward slashes —
	// guards against a future change that lets backslashes leak in on
	// Windows writes.
	raw, err := os.ReadFile(filepath.Join(chunkDir, constants.ChunkMetadataFile))
	require.NoError(t, err)
	assert.NotContains(t, string(raw), `\\`, "metadata must not embed escaped backslashes")
	assert.Contains(t, string(raw), `"sheet_Foo.json"`)
}
