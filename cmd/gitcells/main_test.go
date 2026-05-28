package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestShouldDisableColors locks in the decision matrix for ANSI color
// suppression. The function is pure so the truth table is exhaustive in
// eight rows.
//
// Precedence per CLI convention and https://no-color.org:
//
//	flag == true      ⇒ disable
//	NO_COLOR set      ⇒ disable
//	stderr not a TTY  ⇒ disable
//	otherwise         ⇒ keep colors
//
// All inputs combined as truth-table rows.
func TestShouldDisableColors(t *testing.T) {
	cases := []struct {
		flag    bool
		envVal  string
		isTTY   bool
		expect  bool
		comment string
	}{
		{false, "", true, false, "interactive terminal with no flag/env → keep color"},
		{false, "", false, true, "piped output → strip color"},
		{false, "1", true, true, "NO_COLOR=1 → strip color even in TTY"},
		{false, "1", false, true, "NO_COLOR=1 → strip color when piped"},
		{true, "", true, true, "--no-color in TTY → strip"},
		{true, "", false, true, "--no-color when piped → strip"},
		{true, "1", true, true, "--no-color + NO_COLOR → strip"},
		{true, "1", false, true, "--no-color + NO_COLOR + piped → strip"},
	}

	for _, tc := range cases {
		got := shouldDisableColors(tc.flag, tc.envVal, tc.isTTY)
		assert.Equal(t, tc.expect, got,
			"flag=%v env=%q tty=%v → %s", tc.flag, tc.envVal, tc.isTTY, tc.comment)
	}
}
