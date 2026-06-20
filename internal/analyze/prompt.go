package analyze

import (
	"fmt"
	"strings"

	"github.com/yottayoshida/intent-diff/internal/collect"
)

// BuildPrompt constructs the full prompt for the LLM from the collect result.
func BuildPrompt(cr *collect.CollectResult) string {
	var sb strings.Builder

	sb.WriteString(systemInstructions)
	sb.WriteString("\n\n")

	sb.WriteString("<pr_description>\n")
	if collect.IsEmptyIntent(cr.Intent) {
		sb.WriteString("(No PR description provided. Analyze the diff on its own and flag intent_under_specification.)\n")
	} else {
		sb.WriteString(cr.Intent)
		sb.WriteString("\n")
	}
	sb.WriteString("</pr_description>\n\n")

	sb.WriteString("<diff>\n")
	for _, f := range cr.Files {
		fmt.Fprintf(&sb, "--- %s [category=%s, risk=%s, +%d/-%d]\n",
			f.Path, f.Category, f.Risk, f.Added, f.Deleted)
		if f.Truncated {
			sb.WriteString("(partially truncated)\n")
		}
		sb.WriteString(f.HunkText)
		sb.WriteString("\n")
	}
	sb.WriteString("</diff>\n")

	if len(cr.ExcludedFiles) > 0 {
		sb.WriteString("\n<excluded_files>\n")
		for _, ef := range cr.ExcludedFiles {
			sb.WriteString("- " + ef + "\n")
		}
		sb.WriteString("</excluded_files>\n")
	}

	if cr.Truncated {
		sb.WriteString("\n<truncation_notice>The diff was truncated to fit within the analysis budget. Confidence should be set to medium or low.</truncation_notice>\n")
	}

	return sb.String()
}

const systemInstructions = `You are a PR review triage assistant. Your job is to compare the claimed intent in a PR description with the implementation evidence in the git diff.

IMPORTANT RULES:
1. You are NOT a code reviewer. Do not look for bugs, style issues, or improvements.
2. Focus ONLY on alignment between what the PR claims to do and what the diff actually does.
3. Use the 9-category mismatch taxonomy: scope, contract, risk, test, intent_under_specification, non_code_impact, behavioral_ambiguity, documentation, dependency_risk.
4. Express behavioral observations as HYPOTHESES that need verification, not as facts. Use language like "appears to", "may change", "could affect".
5. Every mismatch must cite specific files from the diff as evidence. Do not reference files not present in the diff.
6. Grade definitions:
   - A: Well-aligned, no material mismatches
   - B: Minor omissions, small gaps but no risk-bearing changes undocumented
   - C: Material omissions, risk-bearing changes not mentioned or scope significantly exceeds claims
   - D: Significant mismatches, explicit claims contradicted by diff evidence
   - E: Critical mismatches, safety/auth/data changes undocumented or explicit non-goals violated
7. If the PR description is empty or too vague, set confidence to "low" and add an intent_under_specification mismatch.
8. The attention_map should be ordered by priority (most important first). The first item should be the single most important thing for the reviewer to check.
9. The suggested_pr_description should be a corrected version that accurately reflects what the diff actually does.

Analyze the PR description and diff below, then produce a structured JSON response.`
