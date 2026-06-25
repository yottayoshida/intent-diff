package render

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

func TestRenderJSON_ValidOutput(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version: "0.1",
		Alignment: analyze.Alignment{
			Grade:               "A",
			Score:               1.0,
			Confidence:          "high",
			HighestRiskCategory: "none",
		},
		ClaimedIntent:          "Test PR",
		SuggestedPRDescription: "Test PR description",
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, result, nil, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	var out JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out.Version != "0.1" {
		t.Errorf("expected version 0.1, got %s", out.Version)
	}
	if out.Alignment.Grade != "A" {
		t.Errorf("expected grade A, got %s", out.Alignment.Grade)
	}
}

func TestRenderJSON_WithValidationIssues(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "B"},
	}
	issues := []analyze.ValidationIssue{
		{Type: "hallucinated_file", Message: "fake.go"},
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, result, issues, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	var out JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.ValidationIssues) != 1 {
		t.Errorf("expected 1 validation issue, got %d", len(out.ValidationIssues))
	}
}

func TestRenderJSON_TruncatedFields(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "C", Score: 0.45},
	}
	meta := RenderMetadata{
		Truncated:      true,
		TruncatedFiles: []string{"big.go"},
		ExcludedFiles:  []string{"vendor/x.go"},
		FilesAnalyzed:  10,
		FilesTotal:     15,
		BudgetChars:    100_000,
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, result, nil, meta); err != nil {
		t.Fatal(err)
	}

	var out JSONOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !out.Truncated {
		t.Error("expected truncated=true")
	}
	if len(out.TruncatedFiles) != 1 || out.TruncatedFiles[0] != "big.go" {
		t.Errorf("expected truncated_files=[big.go], got %v", out.TruncatedFiles)
	}
	if len(out.ExcludedFiles) != 1 || out.ExcludedFiles[0] != "vendor/x.go" {
		t.Errorf("expected excluded_files=[vendor/x.go], got %v", out.ExcludedFiles)
	}
	if out.FilesAnalyzed != 10 {
		t.Errorf("expected files_analyzed=10, got %d", out.FilesAnalyzed)
	}
	if out.FilesTotal != 15 {
		t.Errorf("expected files_total=15, got %d", out.FilesTotal)
	}
}

func TestRenderJSON_BackwardCompat(t *testing.T) {
	result := &analyze.AnalysisResult{
		Version:   "0.1",
		Alignment: analyze.Alignment{Grade: "A", Score: 1.0, Confidence: "high"},
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, result, nil, RenderMetadata{}); err != nil {
		t.Fatal(err)
	}

	raw := buf.String()
	if bytes.Contains(buf.Bytes(), []byte(`"truncated"`)) {
		t.Error("zero-value RenderMetadata should not emit truncated field (omitempty)")
	}
	if bytes.Contains(buf.Bytes(), []byte(`"truncated_files"`)) {
		t.Error("zero-value RenderMetadata should not emit truncated_files field")
	}
	if bytes.Contains(buf.Bytes(), []byte(`"excluded_files"`)) {
		t.Error("zero-value RenderMetadata should not emit excluded_files field")
	}

	var out JSONOutput
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatal(err)
	}
	if out.Alignment.Grade != "A" {
		t.Errorf("expected grade A, got %s", out.Alignment.Grade)
	}
}
