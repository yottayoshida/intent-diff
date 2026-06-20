package analyze

// ValidateResult performs post-hoc validation on the LLM output.
// It checks that referenced file paths actually exist in the diff.
func ValidateResult(result *AnalysisResult, diffFiles map[string]bool) []ValidationIssue {
	var issues []ValidationIssue

	for i, m := range result.Mismatches {
		for _, ev := range m.Evidence {
			if !diffFiles[ev] {
				issues = append(issues, ValidationIssue{
					Type:    ValidationHallucinatedFile,
					Message: ev + " referenced in mismatch evidence but not found in diff",
					Index:   i,
				})
			}
		}
	}

	for i, e := range result.ImplementationEvidence {
		if !diffFiles[e.File] {
			issues = append(issues, ValidationIssue{
				Type:    ValidationHallucinatedFile,
				Message: e.File + " referenced in implementation evidence but not found in diff",
				Index:   i,
			})
		}
	}

	for i, h := range result.BehaviorImpactHypotheses {
		for _, af := range h.AffectedFiles {
			if !diffFiles[af] {
				issues = append(issues, ValidationIssue{
					Type:    ValidationHallucinatedFile,
					Message: af + " referenced in behavior_impact_hypotheses but not found in diff",
					Index:   i,
				})
			}
		}
	}

	for i, a := range result.AttentionMap {
		if !diffFiles[a.File] {
			issues = append(issues, ValidationIssue{
				Type:    ValidationHallucinatedFile,
				Message: a.File + " referenced in attention map but not found in diff",
				Index:   i,
			})
		}
	}

	if result.Alignment.Grade == "" {
		issues = append(issues, ValidationIssue{
			Type:    ValidationMissingField,
			Message: "alignment.grade is empty",
		})
	}

	return issues
}

// ValidationIssueType categorizes validation problems.
type ValidationIssueType string

const (
	ValidationHallucinatedFile ValidationIssueType = "hallucinated_file"
	ValidationMissingField     ValidationIssueType = "missing_field"
)

// ValidationIssue describes a single post-hoc validation problem.
type ValidationIssue struct {
	Type    ValidationIssueType `json:"type"`
	Message string              `json:"message"`
	Index   int                 `json:"index,omitempty"`
}
