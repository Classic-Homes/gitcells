package main

import "github.com/Classic-Homes/gitcells/internal/constants"

// Aliases to the shared constants package so the cmd layer uses one canonical
// permission matrix. New code should reference internal/constants directly.
const (
	dirPermissions  = constants.DirPermissions
	filePermissions = constants.FilePermissions
)
