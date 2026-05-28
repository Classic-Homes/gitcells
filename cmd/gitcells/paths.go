package main

import (
	"path/filepath"

	"github.com/Classic-Homes/gitcells/internal/constants"
	"github.com/Classic-Homes/gitcells/internal/git"
)

// resolveDataDir returns the on-disk path where GitCells stores its
// chunked JSON data, anchored at the git repository containing `start`
// (or at `start` itself when it isn't inside a repo).
//
// Use this instead of the literal `constants.GitCellsDataDir` whenever
// you need a path the caller can actually pass to os.Stat, filepath.Walk,
// SweepOrphans, etc. The literal constant is just `.gitcells/data` —
// relative to whatever the cwd happens to be — and earlier code that
// passed it directly would silently target the wrong directory when
// invoked from a subdir of a repo.
//
// The resolution intentionally mirrors converter/chunking.go's WriteChunks
// so every command (convert, watch, sweep, status, sync) computes the
// same target path for the same source file.
func resolveDataDir(start string) string {
	gitRoot, err := git.FindRepositoryRoot(start)
	if err != nil {
		gitRoot = start
	}
	return filepath.Join(gitRoot, constants.GitCellsDataDir)
}
