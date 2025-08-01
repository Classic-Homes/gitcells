package constants

import "runtime"

var (
	// Version is the current version of GitCells
	Version = "dev"
	// BuildTime is when the binary was built
	BuildTime = "unknown"
	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// SetVersion sets the version at build time
func SetVersion(v string) {
	Version = v
}

// SetBuildTime sets the build time at build time
func SetBuildTime(t string) {
	BuildTime = t
}