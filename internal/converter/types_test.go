package converter

import (
	"testing"
	"github.com/Classic-Homes/sheetsync/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/sirupsen/logrus"
)

func TestDetectCellType(t *testing.T) {
	logger := logrus.New()
	c := &converter{logger: logger}

	tests := []struct {
		name     string
		value    interface{}
		expected models.CellType
	}{
		{
			name:     "string value",
			value:    "hello",
			expected: models.CellTypeString,
		},
		{
			name:     "number as string",
			value:    "123.45",
			expected: models.CellTypeNumber,
		},
		{
			name:     "boolean true as string",
			value:    "true",
			expected: models.CellTypeBoolean,
		},
		{
			name:     "boolean false as string",
			value:    "FALSE",
			expected: models.CellTypeBoolean,
		},
		{
			name:     "formula",
			value:    "=SUM(A1:A10)",
			expected: models.CellTypeFormula,
		},
		{
			name:     "float64 number",
			value:    123.45,
			expected: models.CellTypeNumber,
		},
		{
			name:     "int number",
			value:    42,
			expected: models.CellTypeNumber,
		},
		{
			name:     "boolean value",
			value:    true,
			expected: models.CellTypeBoolean,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: models.CellTypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.detectCellType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCellReference(t *testing.T) {
	tests := []struct {
		name      string
		col       int
		row       int
		expected  string
		expectErr bool
	}{
		{
			name:      "A1",
			col:       1,
			row:       1,
			expected:  "A1",
			expectErr: false,
		},
		{
			name:      "Z99",
			col:       26,
			row:       99,
			expected:  "Z99",
			expectErr: false,
		},
		{
			name:      "AA1",
			col:       27,
			row:       1,
			expected:  "AA1",
			expectErr: false,
		},
		{
			name:      "invalid column",
			col:       0,
			row:       1,
			expected:  "",
			expectErr: true,
		},
		{
			name:      "invalid row",
			col:       1,
			row:       0,
			expected:  "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cellReference(tt.col, tt.row)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatCommitMessage(t *testing.T) {
	template := "SheetSync: {action} {filename} at {timestamp}"
	replacements := map[string]string{
		"action":    "updated",
		"filename":  "test.xlsx",
		"timestamp": "2025-07-30T12:00:00Z",
	}

	expected := "SheetSync: updated test.xlsx at 2025-07-30T12:00:00Z"
	result := formatCommitMessage(template, replacements)
	
	assert.Equal(t, expected, result)
}