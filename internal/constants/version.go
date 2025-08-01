package constants

import "runtime"

var (
	// Version is the current version of GitCells (set at build time via ldflags)
	Version = "0.1.0"
	// BuildTime is when the binary was built (set at build time via ldflags)
	BuildTime = "unknown"
	// CommitHash is the git commit hash (set at build time via ldflags)
	CommitHash = "unknown"
	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)
