package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

const maxAttentionItems = 5

// RenderChecksSummary writes a compact 3-layer Checks summary to w.
// Layer 1: Grade + confidence + score (always)
// Layer 2: Attention Map (max 5 items, only for C-E)
// Layer 3: Mismatches with <details> folding (only for C-E)
func RenderChecksSummary(w io.Writer, result *analyze.AnalysisResult, issues []analyze.ValidationIssue, meta RenderMetadata) error {
	grade := result.Alignment.Grade
	gradeDesc := gradeDescription(grade)

	fmt.Fprintf(w, "## Intent Diff: Grade %s — %s\n", grade, gradeDesc)

	if isCleanGrade(grade) {
		fmt.Fprintf(w, "No significant mismatches detected. (confidence: %s, %d files analyzed)\n",
			result.Alignment.Confidence, meta.FilesAnalyzed)
		return nil
	}

	fmt.Fprintf(w, "**Confidence**: %s | **Score**: %.2f",
		result.Alignment.Confidence, result.Alignment.Score)
	if result.Alignment.HighestRiskCategory != "" && result.Alignment.HighestRiskCategory != "none" {
		fmt.Fprintf(w, " | **Highest risk**: %s", result.Alignment.HighestRiskCategory)
	}
	fmt.Fprintf(w, "\n\n")

	if len(result.AttentionMap) > 0 {
		renderAttentionMapCompact(w, result.AttentionMap)
	}

	if len(result.Mismatches) > 0 {
		renderMismatchesCompact(w, result.Mismatches)
	}

	if meta.Truncated || len(meta.ExcludedFiles) > 0 {
		fmt.Fprintf(w, "\n> %d of %d files analyzed.", meta.FilesAnalyzed, meta.FilesTotal)
		if len(meta.ExcludedFiles) > 0 {
			fmt.Fprintf(w, " %d excluded.", len(meta.ExcludedFiles))
		}
		fmt.Fprintf(w, "\n")
	}

	if len(issues) > 0 {
		fmt.Fprintf(w, "\n> ⚠ %d validation issue(s) detected.\n", len(issues))
	}

	return nil
}

func isCleanGrade(grade string) bool {
	return grade == "A" || grade == "B"
}

func gradeDescription(grade string) string {
	desc := map[string]string{
		"A": "Well-aligned",
		"B": "Minor omissions",
		"C": "Material omissions",
		"D": "Significant mismatches",
		"E": "Critical mismatches",
	}
	if d, ok := desc[grade]; ok {
		return d
	}
	return "Unknown"
}

func renderAttentionMapCompact(w io.Writer, items []analyze.AttentionItem) {
	fmt.Fprintf(w, "### Attention Map\n")
	fmt.Fprintf(w, "| Priority | File | Reason |\n")
	fmt.Fprintf(w, "|----------|------|--------|\n")
	limit := len(items)
	if limit > maxAttentionItems {
		limit = maxAttentionItems
	}
	for _, item := range items[:limit] {
		fmt.Fprintf(w, "| %s | `%s` | %s |\n", escPipe(item.Priority), item.File, escPipe(item.Reason))
	}
	if len(items) > maxAttentionItems {
		fmt.Fprintf(w, "\n*(%d more items not shown)*\n", len(items)-maxAttentionItems)
	}
	fmt.Fprintf(w, "\n")
}

func renderMismatchesCompact(w io.Writer, mismatches []analyze.Mismatch) {
	fmt.Fprintf(w, "### Mismatches (%d)\n", len(mismatches))
	for _, m := range mismatches {
		fmt.Fprintf(w, "**[%s]** %s — severity: %s\n",
			escPipe(string(m.Category)), escPipe(m.Claim), escPipe(string(m.Severity)))
		fmt.Fprintf(w, "> %s\n\n", escPipe(m.Observation))
	}

	fmt.Fprintf(w, "<details><summary>Full analysis</summary>\n\n")
	for _, m := range mismatches {
		if len(m.Evidence) > 0 {
			fmt.Fprintf(w, "**Evidence** (%s): %s\n\n", escPipe(m.Claim), strings.Join(m.Evidence, ", "))
		}
	}
	fmt.Fprintf(w, "</details>\n")
}
