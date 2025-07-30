package models

import "time"

type ExcelDiff struct {
	Timestamp  time.Time   `json:"timestamp"`
	Summary    DiffSummary `json:"summary"`
	SheetDiffs []SheetDiff `json:"sheet_diffs"`
}

type DiffSummary struct {
	TotalChanges   int `json:"total_changes"`
	AddedSheets    int `json:"added_sheets"`
	ModifiedSheets int `json:"modified_sheets"`
	DeletedSheets  int `json:"deleted_sheets"`
}

type SheetDiff struct {
	SheetName string       `json:"sheet_name"`
	Changes   []CellChange `json:"changes"`
}

type CellChange struct {
	Cell     string      `json:"cell"`
	Type     ChangeType  `json:"type"`
	OldValue interface{} `json:"old_value,omitempty"`
	NewValue interface{} `json:"new_value,omitempty"`
}

type ChangeType string

const (
	ChangeTypeAdd    ChangeType = "add"
	ChangeTypeModify ChangeType = "modify"
	ChangeTypeDelete ChangeType = "delete"
)