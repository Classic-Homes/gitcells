package utils

import (
	"github.com/sirupsen/logrus"
)

// NewLogger creates a new configured logger instance
func NewLogger(verbose bool) *logrus.Logger {
	logger := logrus.New()
	
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		DisableColors:   false,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

// LoggerWithField returns a logger with a specific field set
func LoggerWithField(logger *logrus.Logger, key string, value interface{}) *logrus.Entry {
	return logger.WithField(key, value)
}

// LoggerWithFields returns a logger with multiple fields set
func LoggerWithFields(logger *logrus.Logger, fields logrus.Fields) *logrus.Entry {
	return logger.WithFields(fields)
}