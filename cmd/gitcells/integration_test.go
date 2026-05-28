//go:build integration

// Package main integration tests drive the actual cobra commands and
// the watcher in-process so a CI failure on any OS catches a real
// user-visible breakage in the core init → convert → watch → round-trip
// flow. The unit `commands_test.go` only exercises the cobra wiring;
// these tests exercise the wiring AND the underlying packages with a
// real .xlsx fixture.
//
// Run with: go test -tags=integration ./cmd/gitcells -run TestIntegration_
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/converter"
	"github.com/Classic-Homes/gitcells/internal/watcher"
	"github.com/Classic-Homes/gitcells/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// captureStdout swaps os.Stdout for the duration of fn and returns
// whatever fn wrote to it. The diff command prints results via
// fmt.Println, which bypasses cobra's SetOut, so tests that want to
// inspect the rendered output need to redirect at the os level.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()
	require.NoError(t, w.Close())

	captured, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(captured)
}

// realRootCmd wires the actual production commands (not the simplified
// stubs in commands_test.go). Tests drive this just like the user would.
func realRootCmd(t *testing.T) *cobra.Command {
	t.Helper()
	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)

	root := &cobra.Command{Use: "gitcells"}
	root.AddCommand(
		newInitCommand(log),
		newConvertCommand(log),
		newWatchCommand(log),
		newStatusCommand(log),
		newDiffCommand(log),
	)
	root.PersistentFlags().String("config", "", "config file")
	root.PersistentFlags().Bool("verbose", false, "verbose logging")
	return root
}

// chdir cd's into dir for the duration of the test and restores cwd
// on cleanup. cmd/gitcells/init wires the config and git repo relative
// to cwd, so tests must run in a temp dir.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(orig) })
}

// writeFixture builds a small .xlsx with values, a formula, a merged
// range, and a non-ASCII sheet name. Enough to catch a regression in
// the convert/chunk path without depending on test/testdata layout.
func writeFixture(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// Default sheet "Sheet1": values + formula + merged range
	require.NoError(t, f.SetCellValue("Sheet1", "A1", "Item"))
	require.NoError(t, f.SetCellValue("Sheet1", "B1", "Qty"))
	require.NoError(t, f.SetCellValue("Sheet1", "A2", "Widget"))
	require.NoError(t, f.SetCellValue("Sheet1", "B2", 7))
	require.NoError(t, f.SetCellValue("Sheet1", "A3", "Gadget"))
	require.NoError(t, f.SetCellValue("Sheet1", "B3", 3))
	require.NoError(t, f.SetCellFormula("Sheet1", "B4", "=SUM(B2:B3)"))
	require.NoError(t, f.MergeCell("Sheet1", "A5", "B5"))
	require.NoError(t, f.SetCellValue("Sheet1", "A5", "Footer"))

	// Second sheet with unicode name
	idx, err := f.NewSheet("データ")
	require.NoError(t, err)
	require.NoError(t, f.SetCellValue("データ", "A1", "値"))
	require.NoError(t, f.SetCellValue("データ", "A2", 42))
	f.SetActiveSheet(idx)

	require.NoError(t, f.SaveAs(path))
}

func TestIntegration_InitConvertRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	fixture := filepath.Join(tmp, "sample.xlsx")
	writeFixture(t, fixture)

	// 1. init
	root := realRootCmd(t)
	root.SetArgs([]string{"init"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	require.NoError(t, root.Execute(), "init failed: %s", out.String())

	// Config files exist
	for _, f := range []string{".gitcells.yaml", ".gitignore", ".gitattributes"} {
		_, err := os.Stat(f)
		require.NoError(t, err, "%s missing after init", f)
	}

	// gitattributes pins LF for chunk data
	gitattr, err := os.ReadFile(".gitattributes")
	require.NoError(t, err)
	assert.Contains(t, string(gitattr), "eol=lf", ".gitattributes must pin LF")

	// 2. convert sample.xlsx
	root = realRootCmd(t)
	root.SetArgs([]string{"convert", fixture})
	out.Reset()
	root.SetOut(&out)
	root.SetErr(&out)
	require.NoError(t, root.Execute(), "convert failed: %s", out.String())

	chunkDir := filepath.Join(".gitcells", "data", "sample"+constants.ChunksDirSuffix)
	if _, statErr := os.Stat(chunkDir); statErr != nil {
		// Help the failure message: list what's actually under .gitcells/
		listed := listTree(t, ".gitcells")
		t.Fatalf("chunk dir %s not created. .gitcells contents:\n%s\nout: %s", chunkDir, listed, out.String())
	}

	// Metadata + main workbook + per-sheet chunks
	mustExist(t, filepath.Join(chunkDir, constants.ChunkMetadataFile))
	mustExist(t, filepath.Join(chunkDir, constants.WorkbookFileName))

	entries, err := os.ReadDir(chunkDir)
	require.NoError(t, err)
	var sheetChunks int
	for _, e := range entries {
		name := e.Name()
		assert.NotContains(t, name, ".tmp.", "stray staging dir leaked: %s", name)
		assert.NotContains(t, name, ".old.", "stray backup dir leaked: %s", name)
		if strings.HasPrefix(name, "sheet_") && strings.HasSuffix(name, ".json") {
			sheetChunks++
		}
	}
	assert.Equal(t, 2, sheetChunks, "expected 2 sheet chunks (Sheet1 + データ)")

	// 3. round-trip: convert chunks back to a fresh .xlsx and verify
	roundtrip := filepath.Join(tmp, "roundtrip.xlsx")
	root = realRootCmd(t)
	root.SetArgs([]string{"convert", chunkDir, "--output", roundtrip})
	out.Reset()
	root.SetOut(&out)
	root.SetErr(&out)
	require.NoError(t, root.Execute(), "reverse convert failed: %s", out.String())

	rt, err := excelize.OpenFile(roundtrip)
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()

	sheets := rt.GetSheetList()
	assert.Contains(t, sheets, "Sheet1")
	assert.Contains(t, sheets, "データ", "unicode sheet name must round-trip")

	v, err := rt.GetCellValue("Sheet1", "A2")
	require.NoError(t, err)
	assert.Equal(t, "Widget", v)

	formula, err := rt.GetCellFormula("Sheet1", "B4")
	require.NoError(t, err)
	assert.Equal(t, "=SUM(B2:B3)", formula)

	merges, err := rt.GetMergeCells("Sheet1")
	require.NoError(t, err)
	var foundMerge bool
	for _, m := range merges {
		if m.GetStartAxis() == "A5" && m.GetEndAxis() == "B5" {
			foundMerge = true
		}
	}
	assert.True(t, foundMerge, "A5:B5 merge must round-trip")
}

func TestIntegration_WatchDetectsModify(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	// init so .gitcells/ exists and the chunk-dir resolver finds a "git root"
	root := realRootCmd(t)
	root.SetArgs([]string{"init"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	require.NoError(t, root.Execute())

	fixture := filepath.Join(tmp, "watched.xlsx")
	writeFixture(t, fixture)

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(log)

	convOpts := converter.ConvertOptions{
		PreserveFormulas: true,
		PreserveStyles:   true,
		PreserveComments: true,
		ChunkingStrategy: "sheet-based",
	}

	done := make(chan string, 4)
	handler := func(ev watcher.FileEvent) error {
		// Mirror the production watch command's handler.
		if err := conv.ExcelToJSONFile(ev.Path, ev.Path, convOpts); err != nil {
			return err
		}
		done <- ev.Path
		return nil
	}

	cfg := config.GetDefault()
	wcfg := &watcher.Config{
		IgnorePatterns: cfg.Watcher.IgnorePatterns,
		DebounceDelay:  300 * time.Millisecond, // shorter than prod default for test speed
		FileExtensions: cfg.Watcher.FileExtensions,
	}

	fw, err := watcher.NewFileWatcher(wcfg, handler, log)
	require.NoError(t, err)
	require.NoError(t, fw.AddDirectory(tmp))
	require.NoError(t, fw.Start())
	t.Cleanup(func() { _ = fw.Stop(2 * time.Second) })

	// Re-save the workbook to fire a Modify event.
	time.Sleep(100 * time.Millisecond) // let watcher settle
	f, err := excelize.OpenFile(fixture)
	require.NoError(t, err)
	require.NoError(t, f.SetCellValue("Sheet1", "C1", "modified"))
	require.NoError(t, f.Save())
	require.NoError(t, f.Close())

	select {
	case got := <-done:
		assert.Equal(t, fixture, got)
	case <-time.After(8 * time.Second):
		t.Fatalf("watcher did not surface modify within 8s")
	}

	chunkDir := filepath.Join(".gitcells", "data", "watched"+constants.ChunksDirSuffix)
	mustExist(t, filepath.Join(chunkDir, constants.ChunkMetadataFile))
}

// TestIntegration_DiffSingleArg exercises `gitcells diff <file.xlsx>`,
// which compares the live workbook against its stored chunk
// representation. Pre-fix this path returned "not yet implemented" even
// though the help text advertised it as a valid invocation.
func TestIntegration_DiffSingleArg(t *testing.T) {
	tmp := t.TempDir()
	chdir(t, tmp)

	fixture := filepath.Join(tmp, "sample.xlsx")
	writeFixture(t, fixture)

	// 1. init + convert so a chunk tree exists on disk.
	runRoot(t, "init")
	runRoot(t, "convert", fixture)

	// 2. Single-arg diff on the unmodified workbook: format=json so the
	// rendered output is machine-parseable. Result should be an
	// ExcelDiff with no sheet diffs.
	out := captureStdout(t, func() {
		root := realRootCmd(t)
		root.SetArgs([]string{"diff", fixture, "--format", "json"})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		require.NoError(t, root.Execute(), "unmodified diff should succeed")
	})

	var diffResult models.ExcelDiff
	require.NoError(t, json.Unmarshal([]byte(out), &diffResult), "diff output: %q", out)
	assert.False(t, diffResult.HasChanges(), "unmodified workbook must report no changes; got: %+v", diffResult)

	// 3. Mutate the workbook, then single-arg diff must surface the
	// mutated cell.
	f, err := excelize.OpenFile(fixture)
	require.NoError(t, err)
	require.NoError(t, f.SetCellValue("Sheet1", "A2", "Mutated"))
	require.NoError(t, f.Save())
	require.NoError(t, f.Close())

	out = captureStdout(t, func() {
		root := realRootCmd(t)
		root.SetArgs([]string{"diff", fixture, "--format", "json"})
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		require.NoError(t, root.Execute())
	})

	require.NoError(t, json.Unmarshal([]byte(out), &diffResult))
	assert.True(t, diffResult.HasChanges(), "mutated workbook must report changes")
	mentionsA2 := false
	for _, sd := range diffResult.SheetDiffs {
		for _, ch := range sd.Changes {
			if ch.Cell == "A2" {
				mentionsA2 = true
			}
		}
	}
	assert.True(t, mentionsA2, "diff must reference cell A2; got: %+v", diffResult)

	// 4. Remove the stored chunks. Single-arg diff must error with a
	// clear "no stored version" message — not panic, not return success.
	require.NoError(t, os.RemoveAll(filepath.Join(".gitcells", "data")))

	root := realRootCmd(t)
	root.SetArgs([]string{"diff", fixture})
	var errBuf bytes.Buffer
	root.SetOut(io.Discard)
	root.SetErr(&errBuf)
	err = root.Execute()
	require.Error(t, err, "diff must fail when chunks are missing")
	assert.Contains(t, err.Error(), "stored version", "error must explain the problem; got: %v", err)
}

// runRoot is a one-liner for the init+convert prep above; the existing
// pattern of constructing a fresh root, setting buffers, and asserting
// execute() got repetitive.
func runRoot(t *testing.T, args ...string) {
	t.Helper()
	root := realRootCmd(t)
	root.SetArgs(args)
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	require.NoError(t, root.Execute(), "%v failed: %s", args, out.String())
}

func TestIntegration_StagingDirCleanupOnFailure(t *testing.T) {
	// Drive the chunking strategy directly. We force a write failure by
	// making the chunk parent dir read-only after Mkdir succeeds but
	// before the rename phase — easiest is to pre-create the target
	// chunk dir as a *file* so the final rename fails. The staging dir
	// must be cleaned up.
	if isWindows() {
		t.Skip("file-as-dir blocker behaves differently on Windows; covered by other tests")
	}

	tmp := t.TempDir()
	log := logrus.New()
	log.SetLevel(logrus.ErrorLevel)
	conv := converter.NewConverter(log)

	fixture := filepath.Join(tmp, "blocked.xlsx")
	writeFixture(t, fixture)

	// Pre-create the *parent* of the chunk dir as a regular file. This
	// makes MkdirAll on the staging dir's parent fail, and ExcelToJSONFile
	// must surface an error without leaving partial artifacts behind.
	blockerParent := filepath.Join(tmp, ".gitcells")
	require.NoError(t, os.WriteFile(blockerParent, []byte("blocker"), 0o600))

	err := conv.ExcelToJSONFile(fixture, fixture, converter.ConvertOptions{
		PreserveFormulas: true,
		ChunkingStrategy: "sheet-based",
	})
	require.Error(t, err, "expected ExcelToJSONFile to fail when .gitcells is a file")

	// No staging dir should remain in tmp
	entries, readErr := os.ReadDir(tmp)
	require.NoError(t, readErr)
	for _, e := range entries {
		assert.NotContains(t, e.Name(), ".tmp.", "staging dir leaked: %s", e.Name())
	}
}

func TestIntegration_CrossPlatformMetadata(t *testing.T) {
	// Write a chunk tree, then re-serialize its metadata with explicit
	// forward-slash paths (simulating "this tree was authored on Linux,
	// now read on Windows"). The reader must apply filepath.FromSlash
	// to land back on real on-disk filenames regardless of host OS.
	tmp := t.TempDir()
	chdir(t, tmp)

	root := realRootCmd(t)
	root.SetArgs([]string{"init"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	require.NoError(t, root.Execute())

	fixture := filepath.Join(tmp, "cross.xlsx")
	writeFixture(t, fixture)

	log := logrus.New()
	log.SetLevel(logrus.WarnLevel)
	conv := converter.NewConverter(log)
	require.NoError(t, conv.ExcelToJSONFile(fixture, fixture, converter.ConvertOptions{
		PreserveFormulas: true,
		ChunkingStrategy: "sheet-based",
	}))

	chunkDir := filepath.Join(tmp, ".gitcells", "data", "cross"+constants.ChunksDirSuffix)
	metaPath := filepath.Join(chunkDir, constants.ChunkMetadataFile)

	raw, err := os.ReadFile(metaPath)
	require.NoError(t, err)

	var meta map[string]interface{}
	require.NoError(t, json.Unmarshal(raw, &meta))

	// Force every chunk path to forward-slash form, then write back.
	// On Linux this is a no-op (paths are already forward-slash); on
	// Windows it simulates a tree authored on a Unix host.
	files, ok := meta["chunk_files"].([]interface{})
	require.True(t, ok, "metadata chunk_files missing or wrong type")
	for i, f := range files {
		s, ok := f.(string)
		require.True(t, ok)
		files[i] = strings.ReplaceAll(s, "\\", "/")
	}
	meta["chunk_files"] = files

	out, err := json.MarshalIndent(meta, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(metaPath, out, 0o600))

	// Read back via the chunking strategy. If FromSlash is missing on
	// the read path, this would fail on Windows with "no such file"
	// because filepath.Join would receive forward-slash chunk names.
	roundtrip := filepath.Join(tmp, "cross-roundtrip.xlsx")
	require.NoError(t, conv.JSONFileToExcel(chunkDir, roundtrip, converter.ConvertOptions{
		PreserveFormulas: true,
	}))

	rt, err := excelize.OpenFile(roundtrip)
	require.NoError(t, err)
	defer func() { _ = rt.Close() }()
	assert.Contains(t, rt.GetSheetList(), "Sheet1")
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	_, err := os.Stat(path)
	require.NoError(t, err, "expected %s to exist", path)
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

func listTree(t *testing.T, root string) string {
	t.Helper()
	var b strings.Builder
	_ = filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			b.WriteString(p + " ERR: " + err.Error() + "\n")
			return nil
		}
		b.WriteString(p + "\n")
		return nil
	})
	return b.String()
}
