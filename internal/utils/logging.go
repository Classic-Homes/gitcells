package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
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
	default:
		logger.SetFormatter(&SheetSyncFormatter{
			DisableColors:   config.NoColors,
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			AddSource:       config.AddSource,
		})
	}

	// Set output
	if config.File != "" {
		file, err := os.OpenFile(config.File, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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

// SheetSyncFormatter is a custom formatter for SheetSync
type SheetSyncFormatter struct {
	DisableColors   bool
	FullTimestamp   bool
	TimestampFormat string
	AddSource       bool
}

// Format implements the logrus.Formatter interface
func (f *SheetSyncFormatter) Format(entry *logrus.Entry) ([]byte, error) {
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
func (f *SheetSyncFormatter) getLevelColor(level logrus.Level) string {
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
		return "\033[35m" // Magenta
	default:
		return ""
	}
}

// writeFields writes structured fields to the log output
func (f *SheetSyncFormatter) writeFields(b *strings.Builder, fields logrus.Fields) {
	var keys []string
	for key := range fields {
		keys = append(keys, key)
	}

	for i, key := range keys {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(fmt.Sprintf("%s=%v", key, fields[key]))
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
			if pc, file, line, ok := runtime.Caller(8); ok { // Skip logrus frames
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

	if ssErr, ok := err.(*SheetSyncError); ok {
		entry = entry.WithFields(logrus.Fields{
			"error_type":  ssErr.Type,
			"recoverable": ssErr.IsRecoverable(),
			"file":        ssErr.File,
		})
	}

	entry.Error(err.Error())
}

// LogErrorWithRetry logs an error with retry context
func LogErrorWithRetry(logger *logrus.Logger, err error, attempt int, maxAttempts int) {
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
