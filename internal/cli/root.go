package cli

import (
	"github.com/spf13/cobra"
)

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
