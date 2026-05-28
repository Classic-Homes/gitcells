package main

import (
	"fmt"
	"os"

	"github.com/Classic-Homes/gitcells/internal/utils"
)

// resolveDirArg interprets a zero-or-one positional argument as a target
// directory. With no argument, it returns ".". With one argument, it
// stats the path and fails fast unless it points at a real directory.
//
// Previously, status and sync silently fell back to cwd whenever they
// got an arg they didn't understand — e.g. `gitcells status notes.txt`
// would scan the cwd for Excel files and pretend it had honored the
// argument. Surfacing this as an error early matches user intent and
// stops the "why isn't my file showing up?" class of bug report.
func resolveDirArg(args []string) (string, error) {
	if len(args) == 0 {
		return ".", nil
	}
	dir := args[0]
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", utils.NewError(utils.ErrorTypeValidation, "args",
				fmt.Sprintf("directory does not exist: %s", dir))
		}
		return "", utils.WrapFileError(err, utils.ErrorTypeFileSystem, "args", dir, "failed to inspect path")
	}
	if !info.IsDir() {
		return "", utils.NewError(utils.ErrorTypeValidation, "args",
			fmt.Sprintf("not a directory: %s", dir))
	}
	return dir, nil
}
