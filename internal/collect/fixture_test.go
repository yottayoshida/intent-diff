package collect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixture_AuthRefactor(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "auth-refactor")
	runFixtureCollectTest(t, fixtureDir, func(t *testing.T, cr *CollectResult) {
		if cr.TotalAdded == 0 && cr.TotalDeleted == 0 {
			t.Error("expected non-zero line changes")
		}

		hasAuthFile := false
		for _, f := range cr.Files {
			if f.Risk == RiskAuth {
				hasAuthFile = true
			}
		}
		if !hasAuthFile {
			t.Error("expected at least one auth-risk file")
		}
	})
}

func TestFixture_DocsOnly(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "docs-only")
	runFixtureCollectTest(t, fixtureDir, func(t *testing.T, cr *CollectResult) {
		hasSource := false
		for _, f := range cr.Files {
			if f.Category == CategorySource {
				hasSource = true
			}
		}
		if !hasSource {
			t.Error("docs-only fixture should have a hidden source file to detect mismatch")
		}
	})
}

func TestFixture_DependencyLockfile(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "dependency-lockfile")
	runFixtureCollectTest(t, fixtureDir, func(t *testing.T, cr *CollectResult) {
		hasExcluded := false
		for _, ex := range cr.ExcludedFiles {
			if ex == "go.sum" {
				hasExcluded = true
			}
		}
		if !hasExcluded {
			t.Error("go.sum should be excluded as lockfile")
		}
	})
}

func TestFixture_TestOnly(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "test-only")
	runFixtureCollectTest(t, fixtureDir, func(t *testing.T, cr *CollectResult) {
		hasSource := false
		hasTest := false
		for _, f := range cr.Files {
			if f.Category == CategorySource {
				hasSource = true
			}
			if f.Category == CategoryTest {
				hasTest = true
			}
		}
		if !hasTest {
			t.Error("expected at least one test file")
		}
		if !hasSource {
			t.Error("test-only fixture should have a hidden source file change to detect mismatch")
		}
	})
}

func TestFixture_VerificationHypotheses(t *testing.T) {
	fixtureDir := filepath.Join("..", "..", "testdata", "fixtures", "verification-hypotheses")
	runFixtureCollectTest(t, fixtureDir, func(t *testing.T, cr *CollectResult) {
		if cr.TotalAdded == 0 {
			t.Error("expected non-zero added lines")
		}
		hasNewFile := false
		for _, f := range cr.Files {
			if f.Path == "internal/cache/redis.go" {
				hasNewFile = true
			}
		}
		if !hasNewFile {
			t.Error("expected internal/cache/redis.go in changed files")
		}
	})
}

func runFixtureCollectTest(t *testing.T, fixtureDir string, check func(*testing.T, *CollectResult)) {
	t.Helper()

	diffPath := filepath.Join(fixtureDir, "diff.patch")
	intentPath := filepath.Join(fixtureDir, "pr.md")

	if _, err := os.Stat(diffPath); os.IsNotExist(err) {
		t.Skipf("fixture diff not found: %s", diffPath)
	}

	parsed, err := ParseDiffFromFile(diffPath)
	if err != nil {
		t.Fatalf("parse diff: %v", err)
	}

	files := FilesToChangedFiles(parsed)
	for i := range files {
		files[i].Category = ClassifyFile(files[i].Path)
		files[i].Risk = TagRisk(files[i].Path)
	}

	included, excluded, truncated := TruncateFiles(files, DefaultBudgetChars)

	intent := ""
	if _, err := os.Stat(intentPath); err == nil {
		data, _ := os.ReadFile(intentPath)
		intent = string(data)
	}

	totalAdded, totalDeleted, totalChars := 0, 0, 0
	for _, f := range included {
		totalAdded += f.Added
		totalDeleted += f.Deleted
		totalChars += len(f.HunkText)
	}

	cr := &CollectResult{
		Intent:        intent,
		IntentSource:  "file:" + intentPath,
		Files:         included,
		TotalAdded:    totalAdded,
		TotalDeleted:  totalDeleted,
		ExcludedFiles: excluded,
		Truncated:     truncated,
		DiffChars:     totalChars,
		BudgetChars:   DefaultBudgetChars,
	}

	check(t, cr)
}
