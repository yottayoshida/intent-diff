package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

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
	flagConfig      string
	flagOutputMode  string
	flagFailOnGrade string
	flagTimeout     time.Duration
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
	analyzeCmd.Flags().StringVar(&flagConfig, "config", ".intent-diff.yml", "path to config file")
	analyzeCmd.Flags().StringVar(&flagOutputMode, "output-mode", "", "output destination: local, check_summary (default: auto-detect)")
	analyzeCmd.Flags().StringVar(&flagFailOnGrade, "fail-on-grade", "", "exit 1 if alignment grade is at or below this threshold (C, D, or E)")
	analyzeCmd.Flags().BoolVar(&flagJSON, "json", false, "output as JSON instead of Markdown")
	analyzeCmd.Flags().BoolVar(&flagForce, "force", false, "force LLM analysis even for minimal diffs")
	analyzeCmd.Flags().DurationVar(&flagTimeout, "timeout", config.DefaultTimeout, "analysis timeout (e.g. 2m, 10m); range: 30s-30m")
	analyzeCmd.Flags().BoolVar(&flagDumpCollect, "dump-collect", false, "dump intermediate collect-stage JSON for debugging")
	_ = analyzeCmd.Flags().MarkHidden("dump-collect")
}

const minimalDiffThreshold = 5

func runAnalyze(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("config") {
		if _, err := os.Stat(flagConfig); os.IsNotExist(err) {
			return fmt.Errorf("config file not found: %s", flagConfig)
		}
	}

	cfg, err := config.Load(flagConfig)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if cmd.Flags().Changed("output-mode") {
		if !config.ValidOutputMode(flagOutputMode) {
			return fmt.Errorf("invalid --output-mode %q: must be one of: local, check_summary", flagOutputMode)
		}
		cfg.OutputMode = flagOutputMode
	}
	if cmd.Flags().Changed("fail-on-grade") {
		if !config.ValidFailGrade(flagFailOnGrade) {
			return fmt.Errorf("invalid --fail-on-grade %q: must be one of: C, D, E", flagFailOnGrade)
		}
		cfg.Thresholds.FailOnGrade = flagFailOnGrade
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
		stat, statErr := os.Stdin.Stat()
		if statErr == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
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
	included, truncationExcluded, truncated := collect.TruncateFiles(files, cfg.MaxDiffSize)
	allExcluded := make([]string, 0, len(truncationExcluded)+len(ignoredFiles))
	allExcluded = append(allExcluded, truncationExcluded...)
	allExcluded = append(allExcluded, ignoredFiles...)

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
		ExcludedFiles: allExcluded,
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

	meta := render.RenderMetadata{
		Truncated:      truncated,
		TruncatedFiles: collectTruncatedFileNames(included),
		ExcludedFiles:  truncationExcluded,
		FilesAnalyzed:  len(included),
		FilesTotal:     len(included) + len(truncationExcluded),
		BudgetChars:    cfg.MaxDiffSize,
	}

	// --- Minimal diff shortcut ---
	if totalAdded+totalDeleted <= minimalDiffThreshold && !flagForce {
		result := &analyze.AnalysisResult{
			Version: "0.1",
			Alignment: analyze.Alignment{
				Grade:               "A",
				Score:               1.0,
				Confidence:          "high",
				HighestRiskCategory: "none",
			},
			ClaimedIntent:          intent,
			SuggestedPRDescription: intent,
		}
		fmt.Fprintf(os.Stderr, "Changes are minimal (%d lines). Skipping LLM analysis. Use --force to override.\n", totalAdded+totalDeleted)
		return writeOutput(result, nil, cfg, meta)
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

	// --- Resolve timeout ---
	timeout, err := resolveTimeout(cmd, cfg)
	if err != nil {
		return err
	}

	// --- Analyze ---
	prompt := analyze.BuildPrompt(cr)
	schema := analyze.JSONSchema()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	runner := &analyze.ExecClaudeRunner{}
	result, err := runner.Run(ctx, prompt, schema)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("analysis timed out after %s\nThe diff contained %d files (%d chars).\nTry: increase timeout with --timeout 10m, or reduce diff size with ignore patterns in .intent-diff.yml",
				timeout, len(included), totalDiffChars)
		}
		return fmt.Errorf("analysis failed: %w", err)
	}

	// --- Validate ---
	diffFiles := make(map[string]bool)
	for _, f := range included {
		diffFiles[f.Path] = true
	}
	issues := analyze.ValidateResult(result, diffFiles)

	return writeOutput(result, issues, cfg, meta)
}

func resolveOutputMode(cfg *config.Config) string {
	if flagOutputMode != "" {
		return flagOutputMode
	}
	if cfg.OutputMode != "" {
		return cfg.OutputMode
	}
	if os.Getenv("GITHUB_STEP_SUMMARY") != "" {
		return "check_summary"
	}
	return "local"
}

func writeOutput(result *analyze.AnalysisResult, issues []analyze.ValidationIssue, cfg *config.Config, meta render.RenderMetadata) error {
	mode := resolveOutputMode(cfg)

	if mode == "check_summary" {
		if err := writeChecksSummary(result, issues, meta); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to write checks summary: %v\n", err)
		}
	}

	if flagOut != "" {
		return writeToFile(result, issues, meta)
	}

	if mode == "check_summary" && !flagJSON {
		return nil
	}

	var w io.Writer = os.Stdout
	useJSON := flagJSON || cfg.OutputFormat == "json"
	if useJSON {
		return render.RenderJSON(w, result, issues, meta)
	}
	return render.RenderMarkdown(w, result, issues, meta)
}

func writeChecksSummary(result *analyze.AnalysisResult, issues []analyze.ValidationIssue, meta render.RenderMetadata) error {
	summaryPath := os.Getenv("GITHUB_STEP_SUMMARY")
	if summaryPath == "" {
		return fmt.Errorf("GITHUB_STEP_SUMMARY not set")
	}
	f, err := os.OpenFile(summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open step summary: %w", err)
	}
	defer f.Close()
	return render.RenderChecksSummary(f, result, issues, meta)
}

func writeToFile(result *analyze.AnalysisResult, issues []analyze.ValidationIssue, meta render.RenderMetadata) error {
	f, err := os.Create(flagOut)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()
	return render.RenderJSON(f, result, issues, meta)
}

func collectDiffText(files []collect.ChangedFile) string {
	var sb strings.Builder
	for _, f := range files {
		sb.WriteString(f.HunkText)
	}
	return sb.String()
}

func resolveTimeout(cmd *cobra.Command, cfg *config.Config) (time.Duration, error) {
	if cmd.Flags().Changed("timeout") {
		if flagTimeout < config.MinTimeout || flagTimeout > config.MaxTimeout {
			return 0, fmt.Errorf("invalid timeout %s: must be between %s and %s", flagTimeout, config.MinTimeout, config.MaxTimeout)
		}
		return flagTimeout, nil
	}
	return cfg.ResolveTimeout()
}

func collectTruncatedFileNames(files []collect.ChangedFile) []string {
	var names []string
	for _, f := range files {
		if f.Truncated {
			names = append(names, f.Path)
		}
	}
	return names
}
