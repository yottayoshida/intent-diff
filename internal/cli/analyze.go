package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yottayoshida/intent-diff/internal/analyze"
	"github.com/yottayoshida/intent-diff/internal/collect"
	"github.com/yottayoshida/intent-diff/internal/config"
	"github.com/yottayoshida/intent-diff/internal/render"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze PR description against git diff for mismatches",
	Long:  "Collect PR description and git diff, send to LLM for structured comparison, and output a mismatch report with attention map.",
	RunE:  runAnalyze,
}

var (
	flagBase        string
	flagHead        string
	flagDiffFile    string
	flagIntent      string
	flagPRJSON      string
	flagOut         string
	flagJSON        bool
	flagForce       bool
	flagDumpCollect bool
)

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().StringVar(&flagBase, "base", "", "base ref for git diff (default: merge-base with main)")
	analyzeCmd.Flags().StringVar(&flagHead, "head", "HEAD", "head ref for git diff")
	analyzeCmd.Flags().StringVar(&flagDiffFile, "diff-file", "", "path to a pre-generated diff file (instead of running git diff)")
	analyzeCmd.Flags().StringVar(&flagIntent, "intent", "", "path to a PR description markdown file")
	analyzeCmd.Flags().StringVar(&flagPRJSON, "pr-json", "", "path to gh pr view --json title,body output")
	analyzeCmd.Flags().StringVar(&flagOut, "out", "", "output file path (default: stdout)")
	analyzeCmd.Flags().BoolVar(&flagJSON, "json", false, "output as JSON instead of Markdown")
	analyzeCmd.Flags().BoolVar(&flagForce, "force", false, "force LLM analysis even for minimal diffs")
	analyzeCmd.Flags().BoolVar(&flagDumpCollect, "dump-collect", false, "dump intermediate collect-stage JSON for debugging")
	_ = analyzeCmd.Flags().MarkHidden("dump-collect")
}

const minimalDiffThreshold = 5

func runAnalyze(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(".intent-diff.yml")
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// --- Collect: Diff ---
	var files []collect.ChangedFile
	if flagDiffFile != "" {
		parsed, err := collect.ParseDiffFromFile(flagDiffFile)
		if err != nil {
			return err
		}
		files = collect.FilesToChangedFiles(parsed)
	} else {
		parsed, err := collect.ParseDiffFromGit(flagBase, flagHead)
		if err != nil {
			return err
		}
		files = collect.FilesToChangedFiles(parsed)
	}

	// --- Collect: Intent ---
	var intent, intentSource string
	switch {
	case flagIntent != "":
		intent, intentSource, err = collect.ReadIntentFromFile(flagIntent)
	case flagPRJSON != "":
		intent, intentSource, err = collect.ReadIntentFromPRJSON(flagPRJSON)
	default:
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			intent, intentSource, err = collect.ReadIntentFromStdin(os.Stdin)
		} else {
			intent = ""
			intentSource = "none"
		}
	}
	if err != nil {
		return fmt.Errorf("read intent: %w", err)
	}

	// --- Collect: Classify + Risk ---
	for i := range files {
		files[i].Category = collect.ClassifyFile(files[i].Path)
		files[i].Risk = collect.TagRisk(files[i].Path)
	}

	// Apply ignore patterns from config
	var filtered []collect.ChangedFile
	var ignoredFiles []string
	for _, f := range files {
		if cfg.ShouldIgnore(f.Path) {
			ignoredFiles = append(ignoredFiles, f.Path)
		} else {
			filtered = append(filtered, f)
		}
	}
	files = filtered

	// --- Collect: Truncate ---
	included, excluded, truncated := collect.TruncateFiles(files, cfg.MaxDiffSize)
	excluded = append(excluded, ignoredFiles...)

	totalAdded, totalDeleted := 0, 0
	totalDiffChars := 0
	for _, f := range included {
		totalAdded += f.Added
		totalDeleted += f.Deleted
		totalDiffChars += len(f.HunkText)
	}

	cr := &collect.CollectResult{
		Intent:        intent,
		IntentSource:  intentSource,
		Files:         included,
		TotalAdded:    totalAdded,
		TotalDeleted:  totalDeleted,
		ExcludedFiles: excluded,
		Truncated:     truncated,
		DiffChars:     totalDiffChars,
		BudgetChars:   cfg.MaxDiffSize,
	}

	// --- Dump collect (debug) ---
	if flagDumpCollect {
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		return enc.Encode(cr)
	}

	// --- Minimal diff shortcut ---
	if totalAdded+totalDeleted <= minimalDiffThreshold && !flagForce {
		result := &analyze.AnalysisResult{
			Version: "0.1",
			Alignment: analyze.Alignment{
				Grade:              "A",
				Score:              1.0,
				Confidence:         "high",
				HighestRiskCategory: "none",
			},
			ClaimedIntent:          intent,
			SuggestedPRDescription: intent,
		}
		fmt.Fprintf(os.Stderr, "Changes are minimal (%d lines). Skipping LLM analysis. Use --force to override.\n", totalAdded+totalDeleted)
		return writeOutput(result, nil, cfg)
	}

	// --- Pre-flight ---
	if err := analyze.PreflightCheck(); err != nil {
		return err
	}

	// --- Intent quality warnings ---
	if collect.IsEmptyIntent(intent) {
		fmt.Fprintf(os.Stderr, "Warning: PR description is empty or too short. Analysis will proceed with diff-only mode.\n")
	} else {
		diffText := collectDiffText(included)
		if collect.IsCopyPasteIntent(intent, diffText) {
			fmt.Fprintf(os.Stderr, "Warning: PR description appears to be a copy of the diff. Treating as empty intent.\n")
			cr.Intent = ""
		}
	}

	// --- Analyze ---
	prompt := analyze.BuildPrompt(cr)
	schema := analyze.JSONSchema()

	runner := &analyze.ExecClaudeRunner{}
	result, err := runner.Run(context.Background(), prompt, schema)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	// --- Validate ---
	diffFiles := make(map[string]bool)
	for _, f := range included {
		diffFiles[f.Path] = true
	}
	issues := analyze.ValidateResult(result, diffFiles)

	return writeOutput(result, issues, cfg)
}

func writeOutput(result *analyze.AnalysisResult, issues []analyze.ValidationIssue, cfg *config.Config) error {
	var w io.Writer = os.Stdout
	var outFile *os.File
	if flagOut != "" {
		f, err := os.Create(flagOut)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		outFile = f
		w = f
	}

	useJSON := flagJSON || cfg.OutputFormat == "json"
	var renderErr error
	if useJSON {
		renderErr = render.RenderJSON(w, result, issues)
	} else {
		renderErr = render.RenderMarkdown(w, result, issues)
	}

	if outFile != nil {
		if closeErr := outFile.Close(); closeErr != nil {
			if renderErr == nil {
				return fmt.Errorf("close output file: %w", closeErr)
			}
		}
	}
	return renderErr
}

func collectDiffText(files []collect.ChangedFile) string {
	var sb strings.Builder
	for _, f := range files {
		sb.WriteString(f.HunkText)
	}
	return sb.String()
}
