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
	Truncated        bool                      `json:"truncated,omitempty"`
	TruncatedFiles   []string                  `json:"truncated_files,omitempty"`
	ExcludedFiles    []string                  `json:"excluded_files,omitempty"`
	FilesAnalyzed    int                       `json:"files_analyzed,omitempty"`
	FilesTotal       int                       `json:"files_total,omitempty"`
}

// RenderJSON writes the structured JSON output to w.
func RenderJSON(w io.Writer, result *analyze.AnalysisResult, issues []analyze.ValidationIssue, meta RenderMetadata) error {
	out := JSONOutput{
		AnalysisResult:   result,
		ValidationIssues: issues,
		Truncated:        meta.Truncated,
		TruncatedFiles:   meta.TruncatedFiles,
		ExcludedFiles:    meta.ExcludedFiles,
		FilesAnalyzed:    meta.FilesAnalyzed,
		FilesTotal:       meta.FilesTotal,
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encode JSON output: %w", err)
	}
	return nil
}
