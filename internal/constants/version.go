package constants

import "runtime"

// These are populated at build time via ldflags (see Makefile). When the
// binary is run from `go run` or built without ldflags, Version remains
// "dev" and the updater treats it as upgrade-eligible. Do not hardcode a
// numeric version here — the VERSION file is the single source of truth.
var (
	Version    = "dev"
	BuildTime  = "unknown"
	CommitHash = "unknown"
	GoVersion  = runtime.Version()
)
