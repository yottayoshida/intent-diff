package analyze

import (
	"context"
	"strings"
	"testing"
)

func TestMockClaudeRunner_ReturnsResult(t *testing.T) {
	expected := &AnalysisResult{
		Version:   "0.1",
		Alignment: Alignment{Grade: "A", Score: 0.95},
	}
	runner := &MockClaudeRunner{Result: expected}

	got, err := runner.Run(context.Background(), "test prompt", "{}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Version != expected.Version {
		t.Errorf("version = %s, want %s", got.Version, expected.Version)
	}
	if got.Alignment.Grade != expected.Alignment.Grade {
		t.Errorf("grade = %s, want %s", got.Alignment.Grade, expected.Alignment.Grade)
	}
}

func TestMockClaudeRunner_ReturnsError(t *testing.T) {
	runner := &MockClaudeRunner{Err: context.DeadlineExceeded}

	_, err := runner.Run(context.Background(), "test prompt", "{}")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("error = %v, want DeadlineExceeded", err)
	}
}

func TestPreflightCheck_ErrorMessageContainsCIGuidance(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := PreflightCheck()
	if err == nil {
		t.Fatal("expected error with empty PATH")
	}

	msg := err.Error()

	checks := []struct {
		substr string
		desc   string
	}{
		{"npm install -g @anthropic-ai/claude-code", "installation command"},
		{"ANTHROPIC_API_KEY", "CI secret hint"},
		{"docs.anthropic.com", "documentation URL"},
	}

	for _, c := range checks {
		if !strings.Contains(msg, c.substr) {
			t.Errorf("error message should contain %s (%q), got:\n%s", c.desc, c.substr, msg)
		}
	}
}

func TestPreflightCheck_NoSecretLeakage(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	err := PreflightCheck()
	if err == nil {
		t.Fatal("expected error with empty PATH")
	}

	msg := err.Error()

	if strings.Contains(msg, "sk-ant-") {
		t.Error("error message must not contain API key values")
	}
}
