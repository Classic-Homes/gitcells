package constants

// File extensions
const (
	// Excel file extensions
	ExtXLSX = ".xlsx"
	ExtXLS  = ".xls"
	ExtXLSM = ".xlsm"

	// JSON file extension
	ExtJSON = ".json"

	// Configuration file extension
	ExtYAML = ".yaml"
)

// Directory names
const (
	// GitCells directories
	GitCellsDir      = ".gitcells"
	GitCellsDataDir  = ".gitcells/data"
	GitCellsLogsDir  = ".gitcells/logs"
	GitCellsCacheDir = ".gitcells.cache"

	// Generic directories
	LogsDir = "logs"

	// Chunking directory suffix
	ChunksDirSuffix = "_chunks"
)

// File permissions
const (
	// Directory permissions (rwxr-xr-x)
	DirPermissions = 0755

	// File permissions (rw-r--r--)
	FilePermissions = 0644

	// Secure directory permissions (rwxr-x---)
	SecureDirPermissions = 0750

	// Secure file permissions (rw-------)
	SecureFilePermissions = 0600
)

// File names
const (
	// Configuration files
	ConfigFileName     = ".gitcells.yaml"
	ConfigFileBaseName = ".gitcells"

	// Log files
	LogFileName = "gitcells.log"

	// Chunking files
	WorkbookFileName  = "workbook.json"
	ChunkMetadataFile = ".gitcells_chunks.json"
	SheetFilePattern  = "sheet_*.json"
)

// File patterns and globs
const (
	// Excel file patterns
	ExcelGlobPattern = "*.xlsx"
	ExcelTempPattern = "~$*.xlsx"

	// Excel temporary file patterns
	ExcelTempPrefix = "~$"
	TempFilePattern = "*.tmp"
	LockFilePattern = ".~lock.*"

	// JSON file pattern
	JSONGlobPattern = "*.json"
)

// Default values
const (
	// Git defaults
	DefaultGitUserName  = "GitCells"
	DefaultGitUserEmail = "gitcells@localhost"
	DefaultGitBranch    = "main"

	// Default commit template
	DefaultCommitTemplate = "GitCells: {action} {filename} at {timestamp}"
)

// File extension lists
var (
	// Excel file extensions for watching
	ExcelExtensions = []string{ExtXLSX, ExtXLS, ExtXLSM}

	// Default ignore patterns
	DefaultIgnorePatterns = []string{ExcelTempPrefix + "*", TempFilePattern, LockFilePattern}
)
