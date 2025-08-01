package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LogLevel represents different log levels
type LogLevel string

const (
	LogLevelTrace LogLevel = "trace"
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"

	// filePermissions is the permission for created files
	filePermissions = 0600
	// loggerCallDepth is the call stack depth to skip for logging
	loggerCallDepth = 8

	// ANSI color codes
	colorMagenta  = "\033[35m"
	colorDarkGray = "\033[90m"
)

// LogFormat represents different log output formats
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// LogConfig configures logger behavior
type LogConfig struct {
	Level     LogLevel
	Format    LogFormat
	Output    io.Writer
	File      string
	MaxSize   int64 // Max file size in bytes
	MaxAge    int   // Max age in days
	AddSource bool  // Add source file and line
	NoColors  bool  // Disable colors in text format
}

// DefaultLogConfig returns sensible defaults
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:     LogLevelInfo,
		Format:    LogFormatText,
		Output:    os.Stdout,
		AddSource: false,
		NoColors:  false,
	}
}

// NewLogger creates a new configured logger instance
func NewLogger(config *LogConfig) *logrus.Logger {
	if config == nil {
		config = DefaultLogConfig()
	}

	logger := logrus.New()

	// Set log level
	switch config.Level {
	case LogLevelTrace:
		logger.SetLevel(logrus.TraceLevel)
	case LogLevelDebug:
		logger.SetLevel(logrus.DebugLevel)
	case LogLevelInfo:
		logger.SetLevel(logrus.InfoLevel)
	case LogLevelWarn:
		logger.SetLevel(logrus.WarnLevel)
	case LogLevelError:
		logger.SetLevel(logrus.ErrorLevel)
	case LogLevelFatal:
		logger.SetLevel(logrus.FatalLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set formatter
	switch config.Format {
	case LogFormatJSON:
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case LogFormatText:
	default:
		logger.SetFormatter(&GitCellsFormatter{
			DisableColors:   config.NoColors,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			AddSource:       config.AddSource,
		})
	}

	// Set output
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, filePermissions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log file %s: %v\n", config.File, err)
			logger.SetOutput(config.Output)
		} else {
			logger.SetOutput(file)
		}
	} else {
		logger.SetOutput(config.Output)
	}

	// Add hooks for error handling
	logger.AddHook(&ErrorContextHook{})

	return logger
}

// NewVerboseLogger creates a logger for verbose mode
func NewVerboseLogger() *logrus.Logger {
	return NewLogger(&LogConfig{
		Level:     LogLevelDebug,
		Format:    LogFormatText,
		Output:    os.Stdout,
		AddSource: true,
		NoColors:  false,
	})
}

// GitCellsFormatter is a custom formatter for GitCells
type GitCellsFormatter struct {
	DisableColors   bool
	FullTimestamp   bool
	TimestampFormat string
	AddSource       bool
}

// Format implements the logrus.Formatter interface
func (f *GitCellsFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b strings.Builder

	// Timestamp
	if f.FullTimestamp {
		timestampFormat := f.TimestampFormat
		if timestampFormat == "" {
			timestampFormat = time.RFC3339
		}
		b.WriteString(entry.Time.Format(timestampFormat))
		b.WriteString(" ")
	}

	// Level
	levelColor := f.getLevelColor(entry.Level)
	if !f.DisableColors && levelColor != "" {
		b.WriteString(levelColor)
	}
	b.WriteString(fmt.Sprintf("[%s]", strings.ToUpper(entry.Level.String())))
	if !f.DisableColors && levelColor != "" {
		b.WriteString("\033[0m") // Reset color
	}
	b.WriteString(" ")

	// Source information
	if f.AddSource && entry.HasCaller() {
		b.WriteString(fmt.Sprintf("%s:%d ", filepath.Base(entry.Caller.File), entry.Caller.Line))
	}

	// Message
	b.WriteString(entry.Message)

	// Fields
	if len(entry.Data) > 0 {
		b.WriteString(" ")
		f.writeFields(&b, entry.Data)
	}

	b.WriteString("\n")
	return []byte(b.String()), nil
}

// getLevelColor returns ANSI color code for the log level
func (f *GitCellsFormatter) getLevelColor(level logrus.Level) string {
	switch level {
	case logrus.DebugLevel:
		return "\033[36m" // Cyan
	case logrus.InfoLevel:
		return "\033[32m" // Green
	case logrus.WarnLevel:
		return "\033[33m" // Yellow
	case logrus.ErrorLevel:
		return "\033[31m" // Red
	case logrus.FatalLevel:
		return colorMagenta
	case logrus.PanicLevel:
		return colorMagenta
	case logrus.TraceLevel:
		return colorDarkGray
	default:
		return ""
	}
}

// writeFields writes structured fields to the log output
func (f *GitCellsFormatter) writeFields(b *strings.Builder, fields logrus.Fields) {
	var keys []string
	for key := range fields {
		keys = append(keys, key)
	}

	for i, key := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		fmt.Fprintf(b, "%s=%v", key, fields[key])
	}
}

// ErrorContextHook adds error context to log entries
type ErrorContextHook struct{}

// Levels returns the log levels this hook should fire for
func (hook *ErrorContextHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel}
}

// Fire is called when a log entry is made
func (hook *ErrorContextHook) Fire(entry *logrus.Entry) error {
	// Add stack trace context for errors
	if entry.Level <= logrus.ErrorLevel {
		if _, ok := entry.Data["stack"]; !ok {
			// Add caller information
			if pc, file, line, ok := runtime.Caller(loggerCallDepth); ok { // Skip logrus frames
				if details := runtime.FuncForPC(pc); details != nil {
					entry.Data["func"] = details.Name()
					entry.Data["file"] = fmt.Sprintf("%s:%d", filepath.Base(file), line)
				}
			}
		}
	}
	return nil
}

// LoggerWithOperation creates a logger with operation context
func LoggerWithOperation(logger *logrus.Logger, operation string) *logrus.Entry {
	return logger.WithField("operation", operation)
}

// LoggerWithFile creates a logger with file context
func LoggerWithFile(logger *logrus.Logger, file string) *logrus.Entry {
	return logger.WithField("file", filepath.Base(file))
}

// LoggerWithFields returns a logger with multiple fields set
func LoggerWithFields(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}

// LogError logs an error with appropriate context
func LogError(logger *logrus.Logger, err error, operation string, context map[string]interface{}) {
	entry := logger.WithField("operation", operation)

	if context != nil {
		entry = entry.WithFields(logrus.Fields(context))
	}

	if ssErr, ok := err.(*GitCellsError); ok {
		entry = entry.WithFields(logrus.Fields{
			"error_type":  ssErr.Type,
			"recoverable": ssErr.IsRecoverable(),
			"file":        ssErr.File,
		})
	}

	entry.Error(err.Error())
}

// LogErrorWithRetry logs an error with retry context
func LogErrorWithRetry(logger *logrus.Logger, err error, attempt, maxAttempts int) {
	logger.WithFields(logrus.Fields{
		"attempt":      attempt,
		"max_attempts": maxAttempts,
		"will_retry":   attempt < maxAttempts,
	}).Warn(fmt.Sprintf("Operation failed, attempt %d/%d: %v", attempt, maxAttempts, err))
}

// Progress represents a progress tracker for logging
type Progress struct {
	logger *logrus.Entry
	total  int
	done   int
	name   string
}

// NewProgress creates a new progress tracker
func NewProgress(logger *logrus.Logger, name string, total int) *Progress {
	return &Progress{
		logger: logger.WithFields(logrus.Fields{
			"progress": name,
			"total":    total,
		}),
		total: total,
		done:  0,
		name:  name,
	}
}

// Update updates the progress and logs if significant progress was made
func (p *Progress) Update(done int) {
	oldPercent := (p.done * 100) / p.total
	newPercent := (done * 100) / p.total

	p.done = done

	// Log every 10% or at completion
	if newPercent-oldPercent >= 10 || done >= p.total {
		p.logger.WithField("done", done).Infof("%s progress: %d%% (%d/%d)",
			p.name, newPercent, done, p.total)
	}
}

// Finish marks the progress as complete
func (p *Progress) Finish() {
	p.logger.WithField("done", p.total).Infof("%s completed: 100%% (%d/%d)",
		p.name, p.total, p.total)
}

// GetLogFilePath returns the path where log files are stored
func GetLogFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./gitcells.log"
	}

	logDir := filepath.Join(homeDir, ".gitcells", "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "./gitcells.log"
	}

	return filepath.Join(logDir, "gitcells.log")
}

// RotateLogFile rotates the log file if it exceeds the maximum size
func RotateLogFile(maxSizeBytes int64) error {
	logFile := GetLogFilePath()

	// Check if file exists and get its size
	info, err := os.Stat(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, nothing to rotate
		}
		return err
	}

	// If file is smaller than max size, no rotation needed
	if info.Size() < maxSizeBytes {
		return nil
	}

	// Create backup filename with timestamp
	backupFile := fmt.Sprintf("%s.%s", logFile, time.Now().Format("2006-01-02-15-04-05"))

	// Rename current log file to backup
	if err := os.Rename(logFile, backupFile); err != nil {
		return fmt.Errorf("failed to rotate log file: %w", err)
	}

	// Clean up old backup files (keep only 5 most recent)
	if err := cleanupOldLogFiles(5); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to cleanup old log files: %v\n", err)
	}

	return nil
}

// cleanupOldLogFiles removes old log backup files, keeping only the specified number
func cleanupOldLogFiles(keepCount int) error {
	logDir := filepath.Dir(GetLogFilePath())

	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	// Filter and sort backup files by modification time
	var backupFiles []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasPrefix(entry.Name(), "gitcells.log.") {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			backupFiles = append(backupFiles, info)
		}
	}

	// Sort by modification time (newest first)
	sort.Slice(backupFiles, func(i, j int) bool {
		return backupFiles[i].ModTime().After(backupFiles[j].ModTime())
	})

	// Remove excess files
	if len(backupFiles) > keepCount {
		for i := keepCount; i < len(backupFiles); i++ {
			filePath := filepath.Join(logDir, backupFiles[i].Name())
			if err := os.Remove(filePath); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to remove old log file %s: %v\n", filePath, err)
			}
		}
	}

	return nil
}

// Global logger instance
var defaultLogger *logrus.Logger

// init initializes the default logger
func init() {
	// Rotate log file if it's too large (5MB)
	if err := RotateLogFile(5 * 1024 * 1024); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
	}

	logFile := GetLogFilePath()
	defaultLogger = NewLogger(&LogConfig{
		Level:     LogLevelInfo,
		Format:    LogFormatText,
		File:      logFile,
		AddSource: true,
		NoColors:  true,
	})
}

// LogErrorDefault logs an error using the default logger
func LogErrorDefault(err error, message string, fields map[string]interface{}) {
	entry := defaultLogger.WithField("operation", "tui")

	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	if ssErr, ok := err.(*GitCellsError); ok {
		entry = entry.WithFields(logrus.Fields{
			"error_type":  ssErr.Type,
			"recoverable": ssErr.IsRecoverable(),
			"file":        ssErr.File,
		})
	}

	entry.WithError(err).Error(message)
}

// LogInfo logs an info message using the default logger
func LogInfo(message string, fields map[string]interface{}) {
	entry := defaultLogger.WithField("operation", "tui")

	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	entry.Info(message)
}

// LogWarn logs a warning using the default logger
func LogWarn(message string, fields map[string]interface{}) {
	entry := defaultLogger.WithField("operation", "tui")

	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	entry.Warn(message)
}

// LogUserAction logs user actions in the TUI
func LogUserAction(action string, fields map[string]interface{}) {
	entry := defaultLogger.WithFields(logrus.Fields{
		"operation": "user_action",
		"action":    action,
	})

	if fields != nil {
		entry = entry.WithFields(logrus.Fields(fields))
	}

	entry.Info("User action performed")
}

// LogModeChange logs when the user changes TUI modes
func LogModeChange(fromMode, toMode string) {
	defaultLogger.WithFields(logrus.Fields{
		"operation": "mode_change",
		"from_mode": fromMode,
		"to_mode":   toMode,
	}).Info("TUI mode changed")
}

// LogValidationError logs validation errors with context
func LogValidationError(field string, value interface{}, reason string) {
	defaultLogger.WithFields(logrus.Fields{
		"operation": "validation",
		"field":     field,
		"value":     value,
		"reason":    reason,
	}).Warn("Validation failed")
}
