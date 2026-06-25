package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

// RenderMarkdown writes a human-readable Markdown report to w.
func RenderMarkdown(w io.Writer, result *analyze.AnalysisResult, issues []analyze.ValidationIssue, meta RenderMetadata) error {
	fmt.Fprintf(w, "# Intent Diff Report\n\n")

	renderGradeBadge(w, result.Alignment)
	fmt.Fprintf(w, "\n")

	if meta.Truncated {
		excluded := len(meta.TruncatedFiles) + len(meta.ExcludedFiles)
		fmt.Fprintf(w, "> **Partial analysis**: The diff exceeded the analysis budget (%d chars).\n", meta.BudgetChars)
		fmt.Fprintf(w, "> %d of %d files were analyzed; %d file(s) were excluded or truncated.\n", meta.FilesAnalyzed, meta.FilesTotal, excluded)
		fmt.Fprintf(w, "> To analyze the full diff, increase `max_diff_size` in `.intent-diff.yml`.\n\n")
	}

	if len(issues) > 0 {
		fmt.Fprintf(w, "> **Warning**: %d validation issue(s) detected (possible hallucination). See details below.\n\n", len(issues))
	}

	fmt.Fprintf(w, "## Claimed Intent\n\n%s\n\n", result.ClaimedIntent)

	if len(result.AttentionMap) > 0 {
		renderAttentionMap(w, result.AttentionMap)
	}

	if len(result.Mismatches) > 0 {
		renderMismatches(w, result.Mismatches)
	} else {
		fmt.Fprintf(w, "## Mismatches\n\nNo mismatches detected.\n\n")
	}

	if len(result.ImplementationEvidence) > 0 {
		renderEvidence(w, result.ImplementationEvidence)
	}

	if len(result.BehaviorImpactHypotheses) > 0 {
		renderHypotheses(w, result.BehaviorImpactHypotheses)
	}

	if result.SuggestedPRDescription != "" {
		fmt.Fprintf(w, "## Suggested PR Description\n\n%s\n\n", result.SuggestedPRDescription)
	}

	if len(issues) > 0 {
		renderValidationIssues(w, issues)
	}

	return nil
}

func renderGradeBadge(w io.Writer, a analyze.Alignment) {
	gradeDesc := map[string]string{
		"A": "Well-aligned",
		"B": "Minor omissions",
		"C": "Material omissions",
		"D": "Significant mismatches",
		"E": "Critical mismatches",
	}
	desc := gradeDesc[a.Grade]
	if desc == "" {
		desc = "Unknown"
	}

	fmt.Fprintf(w, "**Grade: %s** — %s (confidence: %s, score: %.2f)\n",
		a.Grade, desc, a.Confidence, a.Score)

	if a.HighestRiskCategory != "" && a.HighestRiskCategory != "none" {
		fmt.Fprintf(w, "\nHighest risk category: `%s`\n", a.HighestRiskCategory)
	}
}

func renderAttentionMap(w io.Writer, items []analyze.AttentionItem) {
	fmt.Fprintf(w, "## Attention Map\n\n")
	fmt.Fprintf(w, "| Priority | File | Reason |\n")
	fmt.Fprintf(w, "|----------|------|--------|\n")
	for _, item := range items {
		fmt.Fprintf(w, "| %s | `%s` | %s |\n", escPipe(item.Priority), item.File, escPipe(item.Reason))
	}
	fmt.Fprintf(w, "\n")
}

func escPipe(s string) string {
	return strings.ReplaceAll(s, "|", "\\|")
}

func renderMismatches(w io.Writer, mismatches []analyze.Mismatch) {
	fmt.Fprintf(w, "## Mismatches (%d)\n\n", len(mismatches))
	for i, m := range mismatches {
		fmt.Fprintf(w, "### %d. [%s] %s (severity: %s, confidence: %s)\n\n",
			i+1, escPipe(string(m.Category)), escPipe(m.Claim), escPipe(string(m.Severity)), escPipe(string(m.Confidence)))
		fmt.Fprintf(w, "**Observation**: %s\n\n", escPipe(m.Observation))
		if len(m.Evidence) > 0 {
			fmt.Fprintf(w, "**Evidence**: %s\n\n", strings.Join(m.Evidence, ", "))
		}
		fmt.Fprintf(w, "**Recommended action**: %s\n\n", escPipe(m.RecommendedAction))
	}
}

func renderEvidence(w io.Writer, evidence []analyze.Evidence) {
	fmt.Fprintf(w, "## Implementation Evidence\n\n")
	for _, e := range evidence {
		fmt.Fprintf(w, "- `%s` (%s): %s\n", e.File, e.Category, e.Description)
	}
	fmt.Fprintf(w, "\n")
}

func renderHypotheses(w io.Writer, hypotheses []analyze.Hypothesis) {
	fmt.Fprintf(w, "## Behavior-Impact Hypotheses\n\n")
	fmt.Fprintf(w, "> These are hypotheses that need verification, not confirmed facts.\n\n")
	for i, h := range hypotheses {
		fmt.Fprintf(w, "%d. %s\n", i+1, h.Description)
		if len(h.AffectedFiles) > 0 {
			fmt.Fprintf(w, "   - Files: %s\n", strings.Join(h.AffectedFiles, ", "))
		}
		fmt.Fprintf(w, "   - Verify: %s\n", h.VerificationHint)
	}
	fmt.Fprintf(w, "\n")
}

func renderValidationIssues(w io.Writer, issues []analyze.ValidationIssue) {
	fmt.Fprintf(w, "## Validation Issues\n\n")
	fmt.Fprintf(w, "> The following issues were detected during post-hoc validation.\n\n")
	for _, issue := range issues {
		fmt.Fprintf(w, "- [%s] %s\n", issue.Type, issue.Message)
	}
	fmt.Fprintf(w, "\n")
}
