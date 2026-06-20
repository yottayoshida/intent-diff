package render

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/yottayoshida/intent-diff/internal/analyze"
)

// JSONOutput wraps the analysis result with validation metadata for JSON output.
type JSONOutput struct {
	*analyze.AnalysisResult
	ValidationIssues []analyze.ValidationIssue `json:"validation_issues,omitempty"`
}

// RenderJSON writes the structured JSON output to w.
func RenderJSON(w io.Writer, result *analyze.AnalysisResult, issues []analyze.ValidationIssue) error {
	out := JSONOutput{
		AnalysisResult:   result,
		ValidationIssues: issues,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode JSON output: %w", err)
	}
	return nil
}
