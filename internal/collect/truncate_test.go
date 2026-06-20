package collect

import (
	"strings"
	"testing"
)

func TestTruncateFiles_ExcludesCategories(t *testing.T) {
	files := []ChangedFile{
		{Path: "src/main.go", Category: CategorySource, Risk: RiskOther, HunkText: "code"},
		{Path: "vendor/lib.go", Category: CategoryVendor, Risk: RiskOther, HunkText: "vendor"},
		{Path: "dist/bundle.js", Category: CategoryGenerated, Risk: RiskOther, HunkText: "gen"},
		{Path: "go.sum", Category: CategoryLockfile, Risk: RiskOther, HunkText: "lock"},
	}

	included, excluded, _ := TruncateFiles(files, 10000)
	if len(included) != 1 {
		t.Errorf("expected 1 included file, got %d", len(included))
	}
	if len(excluded) != 3 {
		t.Errorf("expected 3 excluded files, got %d", len(excluded))
	}
	if included[0].Path != "src/main.go" {
		t.Errorf("expected src/main.go, got %s", included[0].Path)
	}
}

func TestTruncateFiles_SortsByRisk(t *testing.T) {
	files := []ChangedFile{
		{Path: "utils.go", Category: CategorySource, Risk: RiskOther, HunkText: "a", Added: 1},
		{Path: "auth.go", Category: CategorySource, Risk: RiskAuth, HunkText: "b", Added: 1},
		{Path: "api.go", Category: CategorySource, Risk: RiskAPI, HunkText: "c", Added: 1},
	}

	included, _, _ := TruncateFiles(files, 10000)
	if included[0].Risk != RiskAuth {
		t.Errorf("first file should be auth risk, got %s", included[0].Risk)
	}
	if included[1].Risk != RiskAPI {
		t.Errorf("second file should be api risk, got %s", included[1].Risk)
	}
}

func TestTruncateFiles_TruncatesOverBudget(t *testing.T) {
	files := []ChangedFile{
		{Path: "auth.go", Category: CategorySource, Risk: RiskAuth, HunkText: strings.Repeat("a", 50), Added: 10},
		{Path: "utils.go", Category: CategorySource, Risk: RiskOther, HunkText: strings.Repeat("b", 50), Added: 10},
	}

	included, _, truncated := TruncateFiles(files, 60)
	if !truncated {
		t.Error("expected truncation to occur")
	}
	if included[1].Truncated != true {
		t.Error("second (lower risk) file should be truncated")
	}
}

func TestTruncateFiles_NoBudgetExceeded(t *testing.T) {
	files := []ChangedFile{
		{Path: "main.go", Category: CategorySource, Risk: RiskOther, HunkText: "small", Added: 1},
	}

	_, _, truncated := TruncateFiles(files, 10000)
	if truncated {
		t.Error("should not truncate when under budget")
	}
}
