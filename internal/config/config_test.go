package config

import (
	"os"
	"path/filepath"
	"testing"
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
