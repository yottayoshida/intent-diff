package render

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

func goldenPath(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "golden", name)
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(goldenPath(name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRenderChecksSummary_Clean(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version: "0.1",
		Alignment: analyze.Alignment{
			Grade:      "A",
			Score:      1.0,
			Confidence: "high",
		},
	}
	meta := RenderMetadata{FilesAnalyzed: 12, FilesTotal: 12}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	want := readGolden(t, "check_summary_clean.md")
	if buf.String() != want {
		t.Errorf("output mismatch.\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestRenderChecksSummary_Mismatch(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version: "0.1",
		Alignment: analyze.Alignment{
			Grade:               "C",
			Score:               0.45,
			Confidence:          "high",
			HighestRiskCategory: "contract",
		},
		AttentionMap: []analyze.AttentionItem{
			{Priority: "high", File: "auth/session.go", Reason: "Session expiry logic changed"},
			{Priority: "medium", File: "api/handler.go", Reason: "New error code added"},
		},
		Mismatches: []analyze.Mismatch{
			{
				Category:          "contract",
				Severity:          "high",
				Confidence:        "high",
				Claim:             "Auth session expiry changed",
				Observation:       "Session timeout reduced from 24h to 1h",
				Evidence:          []string{"auth/session.go"},
				RecommendedAction: "Update PR description",
			},
		},
	}
	meta := RenderMetadata{FilesAnalyzed: 12, FilesTotal: 12}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	want := readGolden(t, "check_summary_mismatch.md")
	if buf.String() != want {
		t.Errorf("output mismatch.\ngot:\n%s\nwant:\n%s", buf.String(), want)
	}
}

func TestRenderChecksSummary_AttentionMapMax5(t *testing.T) {
	items := make([]analyze.AttentionItem, 8)
	for i := range items {
		items[i] = analyze.AttentionItem{Priority: "medium", File: "file" + string(rune('a'+i)) + ".go", Reason: "reason"}
	}
	result := &analyze.AnalysisResult{
		Alignment: analyze.Alignment{Grade: "D", Score: 0.3, Confidence: "medium"},
		AttentionMap: items,
	}
	meta := RenderMetadata{FilesAnalyzed: 8, FilesTotal: 8}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	tableRows := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "| medium") {
			tableRows++
		}
	}
	if tableRows != 5 {
		t.Errorf("expected max 5 attention rows, got %d", tableRows)
	}
	if !strings.Contains(output, "3 more items not shown") {
		t.Error("expected truncation notice")
	}
}

func TestRenderChecksSummary_WithTruncation(t *testing.T) {
	result := &analyze.AnalysisResult{
		Alignment: analyze.Alignment{Grade: "C", Score: 0.5, Confidence: "medium"},
		Mismatches: []analyze.Mismatch{
			{Category: "scope", Severity: "medium", Claim: "test", Observation: "test obs"},
		},
	}
	meta := RenderMetadata{
		Truncated:     true,
		ExcludedFiles: []string{"big.go", "huge.go"},
		FilesAnalyzed: 10,
		FilesTotal:    12,
		BudgetChars:   100000,
	}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	output := buf.String()
	if !strings.Contains(output, "10 of 12 files analyzed") {
		t.Error("expected truncation info")
	}
	if !strings.Contains(output, "2 excluded") {
		t.Error("expected excluded count")
	}
}

func TestRenderChecksSummary_WithValidationIssues(t *testing.T) {
	result := &analyze.AnalysisResult{
		Alignment: analyze.Alignment{Grade: "C", Score: 0.5, Confidence: "low"},
		Mismatches: []analyze.Mismatch{
			{Category: "scope", Severity: "medium", Claim: "test", Observation: "test obs"},
		},
	}
	issues := []analyze.ValidationIssue{
		{Type: "phantom_file", Message: "nonexistent.go not in diff"},
	}
	meta := RenderMetadata{FilesAnalyzed: 5, FilesTotal: 5}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, issues, meta); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "1 validation issue") {
		t.Error("expected validation issue warning")
	}
}

func TestRenderChecksSummary_GradeB_IsClean(t *testing.T) {
	result := &analyze.AnalysisResult{
		Alignment: analyze.Alignment{Grade: "B", Score: 0.85, Confidence: "high"},
	}
	meta := RenderMetadata{FilesAnalyzed: 5, FilesTotal: 5}

	var buf bytes.Buffer
	if err := RenderChecksSummary(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), "No significant mismatches") {
		t.Error("Grade B should be clean output")
	}
}
