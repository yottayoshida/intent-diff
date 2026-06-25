package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxDiffSize != 100_000 {
		t.Errorf("expected max_diff_size 100000, got %d", cfg.MaxDiffSize)
	}
	if cfg.OutputFormat != "markdown" {
		t.Errorf("expected output_format markdown, got %s", cfg.OutputFormat)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	cfg, err := Load("/nonexistent/.intent-diff.yml")
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxDiffSize != 100_000 {
		t.Error("should return defaults for missing file")
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	content := `
ignore:
  - "**/*.generated.go"
  - "vendor/**"
max_diff_size: 50000
output_format: json
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MaxDiffSize != 50000 {
		t.Errorf("expected 50000, got %d", cfg.MaxDiffSize)
	}
	if cfg.OutputFormat != "json" {
		t.Errorf("expected json, got %s", cfg.OutputFormat)
	}
	if len(cfg.Ignore) != 2 {
		t.Errorf("expected 2 ignore patterns, got %d", len(cfg.Ignore))
	}
}

func TestShouldIgnore(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	content := `
ignore:
  - "vendor/**"
  - "*.lock"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path   string
		ignore bool
	}{
		{"vendor/lib/code.go", true},
		{"package.lock", true},
		{"src/main.go", false},
	}

	for _, tt := range tests {
		if got := cfg.ShouldIgnore(tt.path); got != tt.ignore {
			t.Errorf("ShouldIgnore(%q) = %v, want %v", tt.path, got, tt.ignore)
		}
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	if err := os.WriteFile(path, []byte("invalid: [yaml: {broken"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestResolveTimeout_Default(t *testing.T) {
	cfg := DefaultConfig()
	d, err := cfg.ResolveTimeout()
	if err != nil {
		t.Fatal(err)
	}
	if d != 5*time.Minute {
		t.Errorf("expected 5m, got %s", d)
	}
}

func TestResolveTimeout_Parsing(t *testing.T) {
	cfg := &Config{Timeout: "2m"}
	d, err := cfg.ResolveTimeout()
	if err != nil {
		t.Fatal(err)
	}
	if d != 2*time.Minute {
		t.Errorf("expected 2m, got %s", d)
	}
}

func TestResolveTimeout_Invalid(t *testing.T) {
	cfg := &Config{Timeout: "abc"}
	_, err := cfg.ResolveTimeout()
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
	if !strings.Contains(err.Error(), "invalid timeout") {
		t.Errorf("expected 'invalid timeout' in error, got: %s", err)
	}
}

func TestResolveTimeout_OutOfRange(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
	}{
		{"below min", "10s"},
		{"above max", "1h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Timeout: tt.timeout}
			_, err := cfg.ResolveTimeout()
			if err == nil {
				t.Error("expected error for out-of-range timeout")
			}
			if !strings.Contains(err.Error(), "out of range") {
				t.Errorf("expected 'out of range' in error, got: %s", err)
			}
		})
	}
}

func TestResolveTimeout_BoundaryValues(t *testing.T) {
	tests := []struct {
		timeout string
		want    time.Duration
	}{
		{"30s", 30 * time.Second},
		{"30m", 30 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.timeout, func(t *testing.T) {
			cfg := &Config{Timeout: tt.timeout}
			d, err := cfg.ResolveTimeout()
			if err != nil {
				t.Fatalf("boundary value %s should be valid: %v", tt.timeout, err)
			}
			if d != tt.want {
				t.Errorf("expected %s, got %s", tt.want, d)
			}
		})
	}
}

func TestLoad_WithTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	content := `timeout: "3m"`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != "3m" {
		t.Errorf("expected timeout 3m, got %s", cfg.Timeout)
	}
	d, err := cfg.ResolveTimeout()
	if err != nil {
		t.Fatal(err)
	}
	if d != 3*time.Minute {
		t.Errorf("expected 3m duration, got %s", d)
	}
}
