package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	ExitSuccess     = 0
	ExitAnalysis    = 1
	ExitConfig      = 2
)

// ExitError wraps an error with a specific exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }

func exitErrorf(code int, format string, args ...any) *ExitError {
	return &ExitError{Code: code, Err: fmt.Errorf(format, args...)}
}

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "intent-diff",
	Short:   "Compare PR description intent with git diff evidence",
	Long:    "intent-diff extracts claimed intent from a PR description and compares it with implementation evidence from git diff, producing a structured mismatch report to help reviewers allocate attention.",
	Version: version,
}

func Execute() error {
	return rootCmd.Execute()
}
