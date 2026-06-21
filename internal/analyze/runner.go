package analyze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ClaudeRunner is the interface for running LLM analysis.
type ClaudeRunner interface {
	Run(ctx context.Context, prompt string, schema string) (*AnalysisResult, error)
}

// ExecClaudeRunner calls `claude --bare -p` as a subprocess.
type ExecClaudeRunner struct{}

func (r *ExecClaudeRunner) Run(ctx context.Context, prompt string, schema string) (*AnalysisResult, error) {
	args := []string{
		"--bare",
		"-p",
		"--output-format", "json",
		"--json-schema", schema,
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Stdin = strings.NewReader(prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("claude CLI failed: %w\nstderr: %s", err, stderr.String())
	}

	var result AnalysisResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return nil, fmt.Errorf("parse claude output: %w\nraw output: %s", err, stdout.String())
	}

	return &result, nil
}

// MockClaudeRunner returns a fixed result for testing.
type MockClaudeRunner struct {
	Result *AnalysisResult
	Err    error
}

func (r *MockClaudeRunner) Run(ctx context.Context, prompt string, schema string) (*AnalysisResult, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Result, nil
}

// PreflightCheck verifies the claude CLI is available and compatible.
func PreflightCheck() error {
	path, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude CLI not found in PATH.\n\n" +
			"intent-diff requires Claude Code CLI to run analysis.\n\n" +
			"  Local install:  npm install -g @anthropic-ai/claude-code\n" +
			"  CI (GitHub Actions): add a step with 'npm install -g @anthropic-ai/claude-code'\n" +
			"    and set the ANTHROPIC_API_KEY secret.\n\n" +
			"Docs: https://docs.anthropic.com/en/docs/claude-code/overview")
	}

	cmd := exec.Command(path, "--version")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("claude CLI found at %s but 'claude --version' failed: %w\n\n"+
			"Try reinstalling: npm install -g @anthropic-ai/claude-code", path, err)
	}

	version := strings.TrimSpace(string(out))
	if version == "" {
		return fmt.Errorf("claude CLI found at %s but --version returned empty output.\n\n"+
			"Try reinstalling: npm install -g @anthropic-ai/claude-code", path)
	}

	return nil
}
