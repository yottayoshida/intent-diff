package analyze

import (
	"strings"
	"testing"

	"github.com/yottayoshida/intent-diff/internal/collect"
)

func TestBuildPrompt_ContainsIntent(t *testing.T) {
	cr := &collect.CollectResult{
		Intent: "Refactor auth middleware",
		Files: []collect.ChangedFile{
			{Path: "auth.go", Category: "source", Risk: "auth", Added: 5, Deleted: 2, HunkText: "+new code\n-old code\n"},
		},
	}

	prompt := BuildPrompt(cr)

	if !strings.Contains(prompt, "<pr_description>") {
		t.Error("prompt should contain <pr_description> tag")
	}
	if !strings.Contains(prompt, "Refactor auth middleware") {
		t.Error("prompt should contain the intent text")
	}
	if !strings.Contains(prompt, "<diff>") {
		t.Error("prompt should contain <diff> tag")
	}
	if !strings.Contains(prompt, "auth.go") {
		t.Error("prompt should contain file path")
	}
}

func TestBuildPrompt_EmptyIntent(t *testing.T) {
	cr := &collect.CollectResult{
		Intent: "",
		Files:  []collect.ChangedFile{{Path: "main.go", HunkText: "+code\n"}},
	}

	prompt := BuildPrompt(cr)
	if !strings.Contains(prompt, "No PR description provided") {
		t.Error("empty intent should produce a fallback message")
	}
}

func TestBuildPrompt_TruncationNotice(t *testing.T) {
	cr := &collect.CollectResult{
		Intent:    "Some intent",
		Truncated: true,
		Files:     []collect.ChangedFile{{Path: "big.go", HunkText: "+code\n"}},
	}

	prompt := BuildPrompt(cr)
	if !strings.Contains(prompt, "<truncation_notice>") {
		t.Error("truncated diff should include truncation notice")
	}
}

func TestBuildPrompt_ExcludedFiles(t *testing.T) {
	cr := &collect.CollectResult{
		Intent:        "Update deps",
		ExcludedFiles: []string{"vendor/lib.go", "dist/bundle.js"},
		Files:         []collect.ChangedFile{{Path: "main.go", HunkText: "+code\n"}},
	}

	prompt := BuildPrompt(cr)
	if !strings.Contains(prompt, "<excluded_files>") {
		t.Error("excluded files should be listed")
	}
	if !strings.Contains(prompt, "vendor/lib.go") {
		t.Error("excluded file should appear in prompt")
	}
}
