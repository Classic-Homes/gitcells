// scripts/smoke is a portable end-to-end smoke check for a *built*
// gitcells binary. It is invoked from CI on Linux, macOS, and Windows
// to catch the class of bugs that `go test` cannot — path-separator
// handling at the OS boundary, executable startup, exit codes, etc.
//
// Usage: go run scripts/smoke/main.go <path/to/gitcells>
//
// On failure, prints a diagnostic with the binary's stderr and exits
// non-zero. On success, prints "smoke OK" and exits 0.
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: smoke <path/to/gitcells>")
	}
	bin, err := filepath.Abs(os.Args[1])
	if err != nil {
		fail("resolving binary path: %v", err)
	}
	if _, statErr := os.Stat(bin); statErr != nil {
		fail("binary not found at %s: %v", bin, statErr)
	}

	tmp, err := os.MkdirTemp("", "gitcells-smoke-*")
	if err != nil {
		fail("mktemp: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	// 1. --version
	run(bin, tmp, "--version")

	// 2. init in the tempdir
	run(bin, tmp, "init")
	for _, f := range []string{".gitcells.yaml", ".gitignore", ".gitattributes"} {
		if _, err := os.Stat(filepath.Join(tmp, f)); err != nil {
			fail("init did not create %s: %v", f, err)
		}
	}
	// Init must set core.longpaths so Windows users with deeply nested
	// chunk dirs don't trip on MAX_PATH. The setting is a string in
	// .git/config so the check is OS-agnostic.
	gitConfig := filepath.Join(tmp, ".git", "config")
	if cfg, err := os.ReadFile(gitConfig); err != nil { // #nosec G304 - smoke tempdir
		fail("reading .git/config: %v", err)
	} else if !strings.Contains(string(cfg), "longpaths = true") {
		fail("init did not enable core.longpaths in .git/config:\n%s", cfg)
	}

	// 3. Build a small fixture and convert it
	fixture := filepath.Join(tmp, "sample.xlsx")
	if err := writeFixture(fixture); err != nil {
		fail("writing fixture: %v", err)
	}

	run(bin, tmp, "convert", fixture)

	chunkDir := filepath.Join(tmp, ".gitcells", "data", "sample_chunks")
	for _, f := range []string{".gitcells_chunks.json", "workbook.json"} {
		if _, err := os.Stat(filepath.Join(chunkDir, f)); err != nil {
			fail("chunk artifact missing: %s: %v", f, err)
		}
	}

	// 4. Reverse convert and check the resulting xlsx opens cleanly
	roundtrip := filepath.Join(tmp, "roundtrip.xlsx")
	run(bin, tmp, "convert", chunkDir, "--output", roundtrip)

	rt, err := excelize.OpenFile(roundtrip)
	if err != nil {
		fail("opening round-trip xlsx: %v", err)
	}
	sheets := rt.GetSheetList()
	_ = rt.Close()
	if len(sheets) == 0 {
		fail("round-trip xlsx has no sheets")
	}

	// 5. Determinism check on the built binary. Convert the same
	// fixture again in a sibling tempdir and confirm every chunk
	// file is byte-identical to the first conversion. This catches
	// platform-specific encoding drift (e.g. line endings, map
	// iteration order, time-fingerprint regressions) on the actual
	// runner OS, not just in a Go test.
	tmp2, err := os.MkdirTemp("", "gitcells-smoke-det-*")
	if err != nil {
		fail("mktemp determinism: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmp2) }()
	run(bin, tmp2, "init")
	fixture2 := filepath.Join(tmp2, "sample.xlsx")
	if err := copyFilePreservingMtime(fixture, fixture2); err != nil {
		fail("copying fixture for determinism run: %v", err)
	}
	run(bin, tmp2, "convert", fixture2)
	chunkDir2 := filepath.Join(tmp2, ".gitcells", "data", "sample_chunks")
	if err := diffDirsByContent(chunkDir, chunkDir2); err != nil {
		fail("determinism check failed: %v", err)
	}

	// (The case-insensitive sheet-collision protection is exercised by
	// the converter unit test TestWriteChunks_CaseInsensitiveSheetCollision.
	// We can't cover it from this binary-level smoke because Excel
	// itself — and therefore excelize — disallows two sheets whose
	// names differ only in case, so we can't author such a workbook
	// here.)

	// 6. Single-arg diff on unmodified workbook. Pre-fix this returned
	// "not yet implemented"; now it should report no differences and
	// exit 0.
	diffOut, diffErr := runCapture(bin, tmp, "diff", fixture)
	if diffErr != nil {
		fail("single-arg diff: %v\nstdout:\n%s", diffErr, diffOut)
	}
	if !strings.Contains(diffOut, "No differences") {
		fail("single-arg diff did not report no differences:\n%s", diffOut)
	}

	// 7. Bad-sheet policy. Default must fail when a sheet is unreadable;
	// --skip-bad-sheets must succeed and leave the bad sheet out.
	badDir := filepath.Join(tmp, "bad")
	if err := os.Mkdir(badDir, 0o755); err != nil {
		fail("bad dir: %v", err)
	}
	badFixture := filepath.Join(badDir, "bad.xlsx")
	if err := writeFixture(badFixture); err != nil {
		fail("bad fixture: %v", err)
	}
	if err := removeXlsxEntry(badFixture, "xl/worksheets/sheet2.xml"); err != nil {
		fail("corrupt fixture: %v", err)
	}
	if _, err := runCapture(bin, tmp, "convert", badFixture); err == nil {
		fail("default convert must fail on unreadable sheet, got success")
	}
	if _, err := runCapture(bin, tmp, "convert", badFixture, "--skip-bad-sheets"); err != nil {
		fail("--skip-bad-sheets must succeed even with one unreadable sheet: %v", err)
	}

	// 8. --help must not bleed ANSI escapes when stdout is piped
	// (exec.Command gives us a non-TTY pipe). Pre-fix the logger
	// hardcoded DisableColors=false; now TTY detection suppresses them.
	helpOut, helpErr := runCapture(bin, tmp, "--help")
	if helpErr != nil {
		fail("--help failed: %v\n%s", helpErr, helpOut)
	}
	if strings.Contains(helpOut, "\x1b[") {
		fail("--help leaked ANSI escapes when piped:\n%s", helpOut)
	}

	// 9. status --json parses cleanly so scripts can consume the output.
	statusOut, statusErr := runCapture(bin, tmp, "status", "--json")
	if statusErr != nil {
		fail("status --json: %v\n%s", statusErr, statusOut)
	}
	var statusPayload struct {
		Files   []map[string]interface{} `json:"files"`
		Summary map[string]int           `json:"summary"`
	}
	if err := json.Unmarshal([]byte(statusOut), &statusPayload); err != nil {
		fail("status --json is not valid JSON: %v\n%s", err, statusOut)
	}
	if _, ok := statusPayload.Summary["synced"]; !ok {
		fail("status --json missing summary.synced key:\n%s", statusOut)
	}

	fmt.Println("smoke OK")
}

// copyFilePreservingMtime copies src to dst with byte fidelity and
// applies src's mtime/atime to dst. Determinism depends on the source
// mtime since we embed it in metadata.
func copyFilePreservingMtime(src, dst string) error {
	raw, err := os.ReadFile(src) // #nosec G304 - smoke fixture
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, raw, 0o644); err != nil {
		return err
	}
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}

// diffDirsByContent walks both roots and verifies that the set of
// relative file paths is identical and that the bytes match for each.
func diffDirsByContent(a, b string) error {
	collect := func(root string) (map[string][]byte, error) {
		out := make(map[string][]byte)
		err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(p) // #nosec G304 - inside smoke tempdir
			if err != nil {
				return err
			}
			out[filepath.ToSlash(rel)] = data
			return nil
		})
		return out, err
	}
	left, err := collect(a)
	if err != nil {
		return fmt.Errorf("scan %s: %w", a, err)
	}
	right, err := collect(b)
	if err != nil {
		return fmt.Errorf("scan %s: %w", b, err)
	}
	if len(left) != len(right) {
		return fmt.Errorf("file count differs: %d vs %d", len(left), len(right))
	}
	for rel, lBytes := range left {
		rBytes, ok := right[rel]
		if !ok {
			return fmt.Errorf("%s present in %s but not %s", rel, a, b)
		}
		if !bytes.Equal(lBytes, rBytes) {
			return fmt.Errorf("%s differs between conversions", rel)
		}
	}
	return nil
}

func run(bin, cwd string, args ...string) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = cwd
	var stdout, stderr bytes.Buffer
	cmd.Stdout = io.MultiWriter(&stdout, os.Stdout)
	cmd.Stderr = io.MultiWriter(&stderr, os.Stderr)
	if err := cmd.Run(); err != nil {
		fail("`%s %s` failed: %v\nstdout:\n%s\nstderr:\n%s",
			filepath.Base(bin), strings.Join(args, " "), err, stdout.String(), stderr.String())
	}
}

// runCapture is run's quiet sibling — returns combined stdout/stderr and
// the cmd error instead of failing. Tests that want to assert *on* the
// output (or assert that a command failed) use this.
func runCapture(bin, cwd string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	cmd.Dir = cwd
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// removeXlsxEntry rewrites an xlsx zip with the named entry removed.
// Used to simulate an unreadable sheet for the bad-sheet policy check —
// workbook.xml still references the sheet, but the worksheet XML payload
// is gone, so GetRows errors out.
func removeXlsxEntry(xlsxPath, entryName string) error {
	src, err := zip.OpenReader(xlsxPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	matched := false
	for _, f := range src.File {
		if f.Name == entryName {
			matched = true
			continue
		}
		hdr := f.FileHeader
		out, err := w.CreateHeader(&hdr)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			_ = rc.Close()
			return err
		}
		_ = rc.Close()
	}
	if err := w.Close(); err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("entry %q not found in %s", entryName, xlsxPath)
	}
	return os.WriteFile(xlsxPath, buf.Bytes(), 0o644)
}

// writeFixture builds a minimal .xlsx with values, a formula, a merged
// range, and a non-ASCII sheet name — enough surface to exercise the
// converter without depending on any repo-local testdata layout.
func writeFixture(path string) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	if err := f.SetCellValue("Sheet1", "A1", "Item"); err != nil {
		return err
	}
	if err := f.SetCellValue("Sheet1", "B1", "Qty"); err != nil {
		return err
	}
	if err := f.SetCellValue("Sheet1", "A2", "Widget"); err != nil {
		return err
	}
	if err := f.SetCellValue("Sheet1", "B2", 7); err != nil {
		return err
	}
	if err := f.SetCellFormula("Sheet1", "B3", "=B2*2"); err != nil {
		return err
	}
	if err := f.MergeCell("Sheet1", "A4", "B4"); err != nil {
		return err
	}
	if _, err := f.NewSheet("データ"); err != nil {
		return err
	}
	if err := f.SetCellValue("データ", "A1", "値"); err != nil {
		return err
	}
	return f.SaveAs(path)
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "smoke: "+format+"\n", args...)
	os.Exit(1)
}
