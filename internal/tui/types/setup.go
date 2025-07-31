package types

// SetupConfig contains the configuration gathered from the setup wizard
type SetupConfig struct {
	Directory      string
	Pattern        string
	AutoCommit     bool
	AutoPush       bool
	CommitTemplate string
}