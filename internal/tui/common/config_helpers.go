// Package common provides shared utilities for TUI models
package common

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Classic-Homes/gitcells/internal/config"
)

// SetStringValue sets a string configuration value by key path
func SetStringValue(cfg *config.Config, key, value string) error {
	switch key {
	// Git settings
	case "git.remote":
		cfg.Git.Remote = value
	case "git.branch":
		cfg.Git.Branch = value
	case "git.user_name":
		cfg.Git.UserName = value
	case "git.user_email":
		cfg.Git.UserEmail = value
	case "git.commit_template":
		cfg.Git.CommitTemplate = value
	// Converter settings
	case "converter.chunking_strategy":
		cfg.Converter.ChunkingStrategy = value
	default:
		return fmt.Errorf("unknown string key: %s", key)
	}
	return nil
}

// SetDurationValue sets a duration configuration value by key path
func SetDurationValue(cfg *config.Config, key, value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration format: %w", err)
	}

	switch key {
	case "watcher.debounce_delay":
		cfg.Watcher.DebounceDelay = duration
	case "updates.check_interval":
		cfg.Updates.CheckInterval = duration
	default:
		return fmt.Errorf("unknown duration key: %s", key)
	}
	return nil
}

// SetStringSliceValue sets a string slice configuration value by key path
func SetStringSliceValue(cfg *config.Config, key, value string) error {
	// Parse comma-separated values
	values := strings.Split(value, ",")
	for i, v := range values {
		values[i] = strings.TrimSpace(v)
	}

	// Remove empty values
	var filtered []string
	for _, v := range values {
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	switch key {
	case "watcher.directories":
		cfg.Watcher.Directories = filtered
	case "watcher.ignore_patterns":
		cfg.Watcher.IgnorePatterns = filtered
	case "watcher.file_extensions":
		cfg.Watcher.FileExtensions = filtered
	default:
		return fmt.Errorf("unknown string slice key: %s", key)
	}
	return nil
}

// SetIntValue sets an integer configuration value by key path
func SetIntValue(cfg *config.Config, key, value string) error {
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("invalid integer format: %w", err)
	}

	switch key {
	case "converter.max_cells_per_sheet":
		if intVal < 1 {
			return fmt.Errorf("max_cells_per_sheet must be at least 1")
		}
		cfg.Converter.MaxCellsPerSheet = intVal
	default:
		return fmt.Errorf("unknown int key: %s", key)
	}
	return nil
}

// SetBoolValue sets a boolean configuration value by key path
func SetBoolValue(cfg *config.Config, key string, value bool) error {
	switch key {
	// Git settings
	case "git.auto_push":
		cfg.Git.AutoPush = value
	case "git.auto_pull":
		cfg.Git.AutoPull = value
	// Converter settings
	case "converter.preserve_formulas":
		cfg.Converter.PreserveFormulas = value
	case "converter.preserve_styles":
		cfg.Converter.PreserveStyles = value
	case "converter.preserve_comments":
		cfg.Converter.PreserveComments = value
	case "converter.compact_json":
		cfg.Converter.CompactJSON = value
	case "converter.ignore_empty_cells":
		cfg.Converter.IgnoreEmptyCells = value
	// Feature settings
	case "features.experimental":
		cfg.Features.EnableExperimentalFeatures = value
	case "features.beta_updates":
		cfg.Features.EnableBetaUpdates = value
	case "features.telemetry":
		cfg.Features.EnableTelemetry = value
	// Update settings
	case "updates.auto_check":
		cfg.Updates.AutoCheckUpdates = value
	case "updates.include_prereleases":
		cfg.Updates.IncludePrereleases = value
	case "updates.auto_download":
		cfg.Updates.AutoDownloadUpdates = value
	case "updates.notify_on_update":
		cfg.Updates.NotifyOnUpdate = value
	default:
		return fmt.Errorf("unknown boolean key: %s", key)
	}
	return nil
}

// GetStringValue retrieves a string configuration value by key path
func GetStringValue(cfg *config.Config, key string) (string, error) {
	switch key {
	// Git settings
	case "git.remote":
		return cfg.Git.Remote, nil
	case "git.branch":
		return cfg.Git.Branch, nil
	case "git.user_name":
		return cfg.Git.UserName, nil
	case "git.user_email":
		return cfg.Git.UserEmail, nil
	case "git.commit_template":
		return cfg.Git.CommitTemplate, nil
	// Converter settings
	case "converter.chunking_strategy":
		return cfg.Converter.ChunkingStrategy, nil
	default:
		return "", fmt.Errorf("unknown string key: %s", key)
	}
}

// GetBoolValue retrieves a boolean configuration value by key path
func GetBoolValue(cfg *config.Config, key string) (bool, error) {
	switch key {
	// Git settings
	case "git.auto_push":
		return cfg.Git.AutoPush, nil
	case "git.auto_pull":
		return cfg.Git.AutoPull, nil
	// Converter settings
	case "converter.preserve_formulas":
		return cfg.Converter.PreserveFormulas, nil
	case "converter.preserve_styles":
		return cfg.Converter.PreserveStyles, nil
	case "converter.preserve_comments":
		return cfg.Converter.PreserveComments, nil
	case "converter.compact_json":
		return cfg.Converter.CompactJSON, nil
	case "converter.ignore_empty_cells":
		return cfg.Converter.IgnoreEmptyCells, nil
	// Feature settings
	case "features.experimental":
		return cfg.Features.EnableExperimentalFeatures, nil
	case "features.beta_updates":
		return cfg.Features.EnableBetaUpdates, nil
	case "features.telemetry":
		return cfg.Features.EnableTelemetry, nil
	// Update settings
	case "updates.auto_check":
		return cfg.Updates.AutoCheckUpdates, nil
	case "updates.include_prereleases":
		return cfg.Updates.IncludePrereleases, nil
	case "updates.auto_download":
		return cfg.Updates.AutoDownloadUpdates, nil
	case "updates.notify_on_update":
		return cfg.Updates.NotifyOnUpdate, nil
	default:
		return false, fmt.Errorf("unknown boolean key: %s", key)
	}
}

// ToggleBoolValue toggles a boolean configuration value by key path
func ToggleBoolValue(cfg *config.Config, key string) error {
	current, err := GetBoolValue(cfg, key)
	if err != nil {
		return err
	}
	return SetBoolValue(cfg, key, !current)
}
