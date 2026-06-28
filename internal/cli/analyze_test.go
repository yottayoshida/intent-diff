package cli

import (
	"errors"
	"testing"

	"github.com/yottayoshida/intent-diff/internal/config"
)

func TestExitError_Code(t *testing.T) {
	err := exitErrorf(ExitConfig, "bad config: %s", "test")
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatal("expected ExitError")
	}
	if exitErr.Code != ExitConfig {
		t.Errorf("expected code %d, got %d", ExitConfig, exitErr.Code)
	}
}

func TestExitError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &ExitError{Code: ExitAnalysis, Err: inner}
	if !errors.Is(err, inner) {
		t.Error("Unwrap should return inner error")
	}
}

func TestResolveOutputMode_EnvAutoDetect(t *testing.T) {
	cfg := &config.Config{}

	t.Setenv("GITHUB_STEP_SUMMARY", "")
	flagOutputMode = ""
	mode := resolveOutputMode(cfg)
	if mode != "local" {
		t.Errorf("expected local without env, got %s", mode)
	}

	t.Setenv("GITHUB_STEP_SUMMARY", "/tmp/summary.md")
	mode = resolveOutputMode(cfg)
	if mode != "check_summary" {
		t.Errorf("expected check_summary with env, got %s", mode)
	}

	flagOutputMode = "local"
	mode = resolveOutputMode(cfg)
	if mode != "local" {
		t.Errorf("flag should override env, got %s", mode)
	}
	flagOutputMode = ""
}

func TestResolveOutputMode_ConfigOverridesEnv(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "/tmp/summary.md")
	cfg := &config.Config{OutputMode: "local"}

	flagOutputMode = ""
	mode := resolveOutputMode(cfg)
	if mode != "local" {
		t.Errorf("config should override env, got %s", mode)
	}
}

func TestResolveOutputMode_FlagOverridesAll(t *testing.T) {
	t.Setenv("GITHUB_STEP_SUMMARY", "/tmp/summary.md")
	cfg := &config.Config{OutputMode: "local"}

	flagOutputMode = "check_summary"
	mode := resolveOutputMode(cfg)
	if mode != "check_summary" {
		t.Errorf("flag should override all, got %s", mode)
	}
	flagOutputMode = ""
}
