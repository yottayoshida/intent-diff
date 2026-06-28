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

// --- Phase 1 config extension tests ---

func loadFromString(t *testing.T, content string) (*Config, error) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(path)
}

func mustLoad(t *testing.T, content string) *Config {
	t.Helper()
	cfg, err := loadFromString(t, content)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func TestLoad_Phase1Partial(t *testing.T) {
	cfg := mustLoad(t, `
ignore:
  - "vendor/**"
max_diff_size: 100000
output_mode: check_summary
thresholds:
  fail_on_grade: "D"
`)
	if cfg.OutputMode != "check_summary" {
		t.Errorf("expected check_summary, got %s", cfg.OutputMode)
	}
	if cfg.Thresholds.FailOnGrade != "D" {
		t.Errorf("expected D, got %s", cfg.Thresholds.FailOnGrade)
	}
	if cfg.RiskPaths != nil {
		t.Error("expected nil risk_paths")
	}
	if cfg.ProtectedClaims != nil {
		t.Error("expected nil protected_claims")
	}
	if len(cfg.Redaction.Patterns) != 0 {
		t.Error("expected empty redaction patterns")
	}
}

func TestLoad_Phase1Full(t *testing.T) {
	cfg := mustLoad(t, `
ignore:
  - "**/*.generated.go"
  - "vendor/**"
max_diff_size: 80000
output_format: markdown
timeout: "5m"
output_mode: local
risk_paths:
  auth:
    - "internal/auth/**"
    - "pkg/session/**"
  api:
    - "api/v2/**"
protected_claims:
  - "backward compatible"
  - "no breaking changes"
thresholds:
  fail_on_grade: "C"
redaction:
  patterns:
    - "sk-[a-zA-Z0-9]{32,}"
    - "ghp_[a-zA-Z0-9]{36}"
`)
	if cfg.OutputMode != "local" {
		t.Errorf("expected local, got %s", cfg.OutputMode)
	}
	if len(cfg.RiskPaths) != 2 {
		t.Errorf("expected 2 risk categories, got %d", len(cfg.RiskPaths))
	}
	if len(cfg.ProtectedClaims) != 2 {
		t.Errorf("expected 2 claims, got %d", len(cfg.ProtectedClaims))
	}
	if cfg.Thresholds.FailOnGrade != "C" {
		t.Errorf("expected C, got %s", cfg.Thresholds.FailOnGrade)
	}
	if len(cfg.RedactionPatterns()) != 2 {
		t.Errorf("expected 2 redaction patterns, got %d", len(cfg.RedactionPatterns()))
	}
	if cfg.ShouldFlagRisk("internal/auth/login.go") != "auth" {
		t.Error("expected auth risk for internal/auth/login.go")
	}
	if cfg.ShouldFlagRisk("api/v2/handler.go") != "api" {
		t.Error("expected api risk for api/v2/handler.go")
	}
	if cfg.ShouldFlagRisk("main.go") != "" {
		t.Error("expected no risk for main.go")
	}
}

func TestLoad_Phase1Only(t *testing.T) {
	cfg := mustLoad(t, `
output_mode: check_summary
thresholds:
  fail_on_grade: "D"
`)
	if cfg.MaxDiffSize != 100_000 {
		t.Error("expected default max_diff_size")
	}
	if cfg.OutputFormat != "markdown" {
		t.Error("expected default output_format")
	}
	if cfg.OutputMode != "check_summary" {
		t.Errorf("expected check_summary, got %s", cfg.OutputMode)
	}
}

func TestLoad_NestedOnly(t *testing.T) {
	cfg := mustLoad(t, `
thresholds:
  fail_on_grade: "C"
redaction:
  patterns:
    - "AKIA[A-Z0-9]{16}"
`)
	if cfg.Thresholds.FailOnGrade != "C" {
		t.Errorf("expected C, got %s", cfg.Thresholds.FailOnGrade)
	}
	if len(cfg.RedactionPatterns()) != 1 {
		t.Errorf("expected 1 redaction pattern, got %d", len(cfg.RedactionPatterns()))
	}
}

func TestLoad_RiskPathsOnly(t *testing.T) {
	cfg := mustLoad(t, `
risk_paths:
  payment:
    - "pkg/payment/**"
    - "internal/billing/**"
  pii:
    - "internal/user/personal_*"
`)
	if len(cfg.RiskPaths) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cfg.RiskPaths))
	}
	if cfg.ShouldFlagRisk("pkg/payment/stripe.go") != "payment" {
		t.Error("expected payment risk")
	}
	if cfg.ShouldFlagRisk("internal/user/personal_data.go") != "pii" {
		t.Error("expected pii risk")
	}
	if cfg.ShouldFlagRisk("main.go") != "" {
		t.Error("expected no risk")
	}
}

func TestLoad_UnknownKeys(t *testing.T) {
	cfg := mustLoad(t, `
ignore:
  - "vendor/**"
unknown_key: "should be silently ignored"
nested_unknown:
  foo: bar
`)
	if len(cfg.Ignore) != 1 {
		t.Errorf("expected 1 ignore, got %d", len(cfg.Ignore))
	}
}

func TestLoad_EmptyNested(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"null value", "thresholds:\nredaction:\n"},
		{"explicit null", "thresholds: null\nredaction: ~\n"},
		{"empty object", "thresholds: {}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustLoad(t, tt.content)
			if cfg.Thresholds.FailOnGrade != "" {
				t.Errorf("expected empty fail_on_grade, got %s", cfg.Thresholds.FailOnGrade)
			}
			if !cfg.FailOnGrade("E") {
				// empty threshold = never fail
			}
			if cfg.FailOnGrade("A") {
				t.Error("empty threshold should never fail")
			}
		})
	}
}

func TestValidate_BadOutputMode(t *testing.T) {
	_, err := loadFromString(t, `output_mode: "invalid_mode"`)
	if err == nil {
		t.Fatal("expected error for invalid output_mode")
	}
	if !strings.Contains(err.Error(), "invalid output_mode") {
		t.Errorf("expected 'invalid output_mode' in error, got: %s", err)
	}
}

func TestValidate_BadOutputFormat(t *testing.T) {
	_, err := loadFromString(t, `output_format: "xml"`)
	if err == nil {
		t.Fatal("expected error for invalid output_format")
	}
	if !strings.Contains(err.Error(), "invalid output_format") {
		t.Errorf("expected 'invalid output_format' in error, got: %s", err)
	}
}

func TestValidate_BadFailOnGrade(t *testing.T) {
	_, err := loadFromString(t, "thresholds:\n  fail_on_grade: \"F\"\n")
	if err == nil {
		t.Fatal("expected error for invalid fail_on_grade")
	}
	if !strings.Contains(err.Error(), "invalid thresholds.fail_on_grade") {
		t.Errorf("expected 'invalid thresholds.fail_on_grade' in error, got: %s", err)
	}
}

func TestLoad_BadGlob_RiskPaths(t *testing.T) {
	_, err := loadFromString(t, "risk_paths:\n  auth:\n    - \"[invalid\"\n")
	if err == nil {
		t.Fatal("expected error for invalid glob")
	}
	if !strings.Contains(err.Error(), "invalid risk_paths pattern") {
		t.Errorf("expected 'invalid risk_paths pattern' in error, got: %s", err)
	}
}

func TestLoad_BadRegex_Redaction(t *testing.T) {
	_, err := loadFromString(t, "redaction:\n  patterns:\n    - \"(unclosed\"\n")
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid redaction pattern") {
		t.Errorf("expected 'invalid redaction pattern' in error, got: %s", err)
	}
}

func TestLoad_BOM(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".intent-diff.yml")
	bom := []byte{0xEF, 0xBB, 0xBF}
	content := append(bom, []byte("output_mode: check_summary\n")...)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.OutputMode != "check_summary" {
		t.Errorf("BOM should be stripped; expected check_summary, got %q", cfg.OutputMode)
	}
}

func TestValidate_TrailingWhitespace(t *testing.T) {
	cfg := mustLoad(t, `output_mode: "local   "`)
	if cfg.OutputMode != "local" {
		t.Errorf("expected trimmed 'local', got %q", cfg.OutputMode)
	}
}

func TestLoad_EmptyCategoryKey(t *testing.T) {
	_, err := loadFromString(t, "risk_paths:\n  \"\":\n    - \"pkg/**\"\n")
	if err == nil {
		t.Fatal("expected error for empty category key")
	}
	if !strings.Contains(err.Error(), "empty category key") {
		t.Errorf("expected 'empty category key' in error, got: %s", err)
	}
}

func TestValidate_EmptyClaims(t *testing.T) {
	cfg := mustLoad(t, `
protected_claims:
  - ""
  - "   "
  - "backward compatible"
`)
	if len(cfg.ProtectedClaims) != 1 {
		t.Errorf("expected 1 claim after filtering, got %d", len(cfg.ProtectedClaims))
	}
	if cfg.ProtectedClaims[0] != "backward compatible" {
		t.Errorf("expected 'backward compatible', got %q", cfg.ProtectedClaims[0])
	}
}

func TestResolveOutputMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"empty defaults to local", "", "local"},
		{"explicit local", "local", "local"},
		{"check_summary", "check_summary", "check_summary"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{OutputMode: tt.mode}
			if got := cfg.ResolveOutputMode(); got != tt.want {
				t.Errorf("ResolveOutputMode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFailOnGrade(t *testing.T) {
	tests := []struct {
		threshold string
		grade     string
		want      bool
	}{
		{"", "E", false},
		{"", "A", false},
		{"C", "A", false},
		{"C", "B", false},
		{"C", "C", true},
		{"C", "D", true},
		{"C", "E", true},
		{"D", "C", false},
		{"D", "D", true},
		{"D", "E", true},
		{"E", "D", false},
		{"E", "E", true},
	}
	for _, tt := range tests {
		t.Run(tt.threshold+"_"+tt.grade, func(t *testing.T) {
			cfg := &Config{Thresholds: ThresholdConfig{FailOnGrade: tt.threshold}}
			if got := cfg.FailOnGrade(tt.grade); got != tt.want {
				t.Errorf("FailOnGrade(%q) with threshold %q = %v, want %v", tt.grade, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestLoad_BackwardCompat_V01Configs(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"R1 absent", ""},
		{"R2 minimal", "ignore:\n  - \"vendor/**\"\n"},
		{"R3 v0.1 full", "ignore:\n  - \"**/*.generated.go\"\n  - \"vendor/**\"\nmax_diff_size: 50000\noutput_format: json\ntimeout: \"3m\"\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *Config
			var err error
			if tt.content == "" {
				cfg, err = Load("/nonexistent/.intent-diff.yml")
			} else {
				cfg, err = loadFromString(t, tt.content)
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.OutputMode != "" {
				t.Error("Phase 1 OutputMode should be zero")
			}
			if cfg.RiskPaths != nil {
				t.Error("Phase 1 RiskPaths should be nil")
			}
			if cfg.ProtectedClaims != nil {
				t.Error("Phase 1 ProtectedClaims should be nil")
			}
			if cfg.Thresholds.FailOnGrade != "" {
				t.Error("Phase 1 FailOnGrade should be empty")
			}
			if len(cfg.Redaction.Patterns) != 0 {
				t.Error("Phase 1 Redaction.Patterns should be empty")
			}
			if cfg.ResolveOutputMode() != "local" {
				t.Error("zero OutputMode should resolve to local")
			}
			if cfg.FailOnGrade("E") {
				t.Error("zero threshold should never fail")
			}
			if cfg.ShouldFlagRisk("anything.go") != "" {
				t.Error("nil risk paths should flag nothing")
			}
			if len(cfg.RedactionPatterns()) != 0 {
				t.Error("nil redaction should return empty")
			}
		})
	}
}
