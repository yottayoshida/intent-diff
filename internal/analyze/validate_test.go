package analyze

import "testing"

func TestValidateResult_HallucinatedFiles(t *testing.T) {
	result := &AnalysisResult{
		Alignment: Alignment{Grade: "C"},
		ImplementationEvidence: []Evidence{
			{File: "real.go", Description: "exists"},
			{File: "hallucinated.go", Description: "does not exist"},
		},
		Mismatches: []Mismatch{
			{
				Category: MismatchScope,
				Evidence: []string{"real.go", "fake.go"},
			},
		},
		AttentionMap: []AttentionItem{
			{File: "real.go", Reason: "important"},
			{File: "ghost.go", Reason: "made up"},
		},
	}

	diffFiles := map[string]bool{"real.go": true}
	issues := ValidateResult(result, diffFiles)

	hallucinatedCount := 0
	for _, issue := range issues {
		if issue.Type == ValidationHallucinatedFile {
			hallucinatedCount++
		}
	}

	if hallucinatedCount != 3 {
		t.Errorf("expected 3 hallucinated file issues, got %d", hallucinatedCount)
	}
}

func TestValidateResult_MissingGrade(t *testing.T) {
	result := &AnalysisResult{
		Alignment: Alignment{Grade: ""},
	}

	issues := ValidateResult(result, map[string]bool{})

	found := false
	for _, issue := range issues {
		if issue.Type == ValidationMissingField {
			found = true
		}
	}
	if !found {
		t.Error("expected missing_field issue for empty grade")
	}
}

func TestValidateResult_Clean(t *testing.T) {
	result := &AnalysisResult{
		Alignment: Alignment{Grade: "A"},
		ImplementationEvidence: []Evidence{
			{File: "main.go", Description: "entry point"},
		},
		AttentionMap: []AttentionItem{
			{File: "main.go", Reason: "check"},
		},
	}

	issues := ValidateResult(result, map[string]bool{"main.go": true})
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(issues))
	}
}
