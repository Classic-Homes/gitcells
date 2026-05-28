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
	"bytes"
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
