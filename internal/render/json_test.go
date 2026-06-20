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
			Grade:              "A",
			Score:              1.0,
			Confidence:         "high",
			HighestRiskCategory: "none",
		},
		ClaimedIntent:          "Test PR",
		SuggestedPRDescription: "Test PR description",
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, result, nil); err != nil {
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
	if err := RenderJSON(&buf, result, issues); err != nil {
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
