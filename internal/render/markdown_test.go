package render

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

func TestRenderMarkdown_GradeA(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version: "0.1",
		Alignment: analyze.Alignment{
			Grade:               "A",
			Score:               0.95,
			Confidence:          "high",
			HighestRiskCategory: "none",
		},
		ClaimedIntent:          "Refactor auth middleware",
		SuggestedPRDescription: "Refactor auth middleware without behavior changes",
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, nil, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Grade: A") {
		t.Error("should contain grade")
	}
	if !strings.Contains(out, "Well-aligned") {
		t.Error("should contain grade description")
	}
	if !strings.Contains(out, "No mismatches detected") {
		t.Error("should indicate no mismatches")
	}
}

func TestRenderMarkdown_WithMismatches(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version: "0.1",
		Alignment: analyze.Alignment{
			Grade:               "C",
			Score:               0.5,
			Confidence:          "high",
			HighestRiskCategory: "scope",
		},
		ClaimedIntent: "Docs-only update",
		Mismatches: []analyze.Mismatch{
			{
				Category:          "scope",
				Severity:          "high",
				Confidence:        "high",
				Claim:             "Documentation only",
				Observation:       "Source code was also modified",
				Evidence:          []string{"handler.go"},
				RecommendedAction: "Update PR description to reflect code changes",
			},
		},
		AttentionMap: []analyze.AttentionItem{
			{File: "handler.go", Reason: "Unexpected source change", Priority: "high"},
		},
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, nil, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Mismatches (1)") {
		t.Error("should show mismatch count")
	}
	if !strings.Contains(out, "scope") {
		t.Error("should show mismatch category")
	}
	if !strings.Contains(out, "Attention Map") {
		t.Error("should contain attention map")
	}
}

func TestRenderMarkdown_ValidationIssues(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "B", Score: 0.7, Confidence: "medium"},
	}
	issues := []analyze.ValidationIssue{
		{Type: "hallucinated_file", Message: "ghost.go not in diff"},
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, issues, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Warning") {
		t.Error("should show warning for validation issues")
	}
	if !strings.Contains(out, "Validation Issues") {
		t.Error("should contain validation issues section")
	}
}

func TestRenderMarkdown_PartialWarning(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "C", Score: 0.45, Confidence: "low"},
	}
	meta := RenderMetadata{
		Truncated:      true,
		TruncatedFiles: []string{"large_file.go"},
		ExcludedFiles:  []string{"vendor/lib.go", "generated.go"},
		FilesAnalyzed:  42,
		FilesTotal:     58,
		BudgetChars:    100_000,
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Partial analysis") {
		t.Error("should contain partial analysis warning")
	}
	if !strings.Contains(out, "42 of 58 files") {
		t.Error("should contain file counts")
	}
	if !strings.Contains(out, "100000 chars") {
		t.Error("should contain budget chars")
	}
	if !strings.Contains(out, "max_diff_size") {
		t.Error("should contain config hint")
	}
}

func TestRenderMarkdown_PartialWarning_ExcludedOnly(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "B", Score: 0.8, Confidence: "medium"},
	}
	meta := RenderMetadata{
		Truncated:     false,
		ExcludedFiles: []string{"vendor/lib.go"},
		FilesAnalyzed: 5,
		FilesTotal:    6,
		BudgetChars:   100_000,
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "Partial analysis") {
		t.Error("should show partial analysis warning when files are excluded by category")
	}
	if !strings.Contains(out, "5 of 6 files") {
		t.Error("should show correct file counts")
	}
	if !strings.Contains(out, "1 file(s) were excluded") {
		t.Error("should show excluded file count")
	}
}

func TestRenderMarkdown_NoPartialWarning(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "A", Score: 1.0, Confidence: "high"},
	}

	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, result, nil, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if strings.Contains(out, "Partial analysis") {
		t.Error("should not contain partial analysis warning when not truncated")
	}
}
