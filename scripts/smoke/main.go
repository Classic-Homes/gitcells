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

	fmt.Println("smoke OK")
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
