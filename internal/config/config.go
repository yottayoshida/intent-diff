package config

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultTimeout = 5 * time.Minute
	MinTimeout     = 30 * time.Second
	MaxTimeout     = 30 * time.Minute
)

// ThresholdConfig controls when the action should fail.
type ThresholdConfig struct {
	FailOnGrade string `yaml:"fail_on_grade"`
}

// RedactionConfig controls what patterns to redact before LLM analysis.
type RedactionConfig struct {
	Patterns []string `yaml:"patterns"`
}

// Config represents the .intent-diff.yml configuration file.
type Config struct {
	Ignore       []string `yaml:"ignore"`
	MaxDiffSize  int      `yaml:"max_diff_size"`
	OutputFormat string   `yaml:"output_format"`
	Timeout      string   `yaml:"timeout"`

	OutputMode      string              `yaml:"output_mode"`
	RiskPaths       map[string][]string `yaml:"risk_paths"`
	ProtectedClaims []string            `yaml:"protected_claims"`
	Thresholds      ThresholdConfig     `yaml:"thresholds"`
	Redaction       RedactionConfig     `yaml:"redaction"`

	ignoreGlobs    []glob.Glob
	riskPathGlobs  map[string][]glob.Glob
	redactionRegex []*regexp.Regexp
}

// ResolveTimeout parses the Timeout string and validates bounds.
// Returns DefaultTimeout when Timeout is empty.
func (c *Config) ResolveTimeout() (time.Duration, error) {
	if c.Timeout == "" {
		return DefaultTimeout, nil
	}
	d, err := time.ParseDuration(c.Timeout)
	if err != nil {
		return 0, fmt.Errorf("invalid timeout %q: %w", c.Timeout, err)
	}
	if d < MinTimeout || d > MaxTimeout {
		return 0, fmt.Errorf("timeout %s out of range: must be between %s and %s", d, MinTimeout, MaxTimeout)
	}
	return d, nil
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MaxDiffSize:  100_000,
		OutputFormat: "markdown",
	}
}

// Load reads the config from the given path, falling back to defaults for missing fields.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MaxDiffSize <= 0 {
		cfg.MaxDiffSize = 100_000
	}

	if err := cfg.compileIgnores(); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	if err := cfg.compileRiskPaths(); err != nil {
		return nil, err
	}

	if err := cfg.compileRedaction(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) compileIgnores() error {
	c.ignoreGlobs = make([]glob.Glob, 0, len(c.Ignore))
	for _, pattern := range c.Ignore {
		g, err := glob.Compile(pattern, '/')
		if err != nil {
			return fmt.Errorf("invalid ignore pattern %q: %w", pattern, err)
		}
		c.ignoreGlobs = append(c.ignoreGlobs, g)
	}
	return nil
}

// ShouldIgnore returns true if the given path matches any ignore pattern.
func (c *Config) ShouldIgnore(path string) bool {
	for _, g := range c.ignoreGlobs {
		if g.Match(path) {
			return true
		}
	}
	return false
}

// ResolveOutputMode returns the effective output mode.
// Priority: explicit value > "local" default.
func (c *Config) ResolveOutputMode() string {
	if c.OutputMode != "" {
		return c.OutputMode
	}
	return "local"
}

var gradeRank = map[string]int{"A": 1, "B": 2, "C": 3, "D": 4, "E": 5}

// FailOnGrade returns true if the given grade meets or exceeds the configured threshold.
func (c *Config) FailOnGrade(grade string) bool {
	threshold := c.Thresholds.FailOnGrade
	if threshold == "" {
		return false
	}
	return gradeRank[grade] >= gradeRank[threshold]
}

// ShouldFlagRisk returns the custom risk category for a path, or empty string.
func (c *Config) ShouldFlagRisk(path string) string {
	for category, globs := range c.riskPathGlobs {
		for _, g := range globs {
			if g.Match(path) {
				return category
			}
		}
	}
	return ""
}

// RedactionPatterns returns compiled redaction regexps.
func (c *Config) RedactionPatterns() []*regexp.Regexp {
	return c.redactionRegex
}

var validOutputModes = map[string]bool{"": true, "local": true, "check_summary": true}
var validOutputFormats = map[string]bool{"markdown": true, "json": true, "": true}
var validFailGrades = map[string]bool{"": true, "C": true, "D": true, "E": true}

func ValidOutputMode(v string) bool { return validOutputModes[v] }
func ValidFailGrade(v string) bool  { return validFailGrades[v] }

func (c *Config) validate() error {
	c.OutputMode = strings.TrimSpace(c.OutputMode)
	if !validOutputModes[c.OutputMode] {
		return fmt.Errorf("invalid output_mode %q: must be one of: local, check_summary", c.OutputMode)
	}

	c.OutputFormat = strings.TrimSpace(c.OutputFormat)
	if !validOutputFormats[c.OutputFormat] {
		return fmt.Errorf("invalid output_format %q: must be one of: markdown, json", c.OutputFormat)
	}

	c.Thresholds.FailOnGrade = strings.TrimSpace(c.Thresholds.FailOnGrade)
	if !validFailGrades[c.Thresholds.FailOnGrade] {
		return fmt.Errorf("invalid thresholds.fail_on_grade %q: must be one of: C, D, E", c.Thresholds.FailOnGrade)
	}

	for category := range c.RiskPaths {
		if strings.TrimSpace(category) == "" {
			return fmt.Errorf("risk_paths contains empty category key")
		}
	}

	filtered := c.ProtectedClaims[:0]
	for _, claim := range c.ProtectedClaims {
		if strings.TrimSpace(claim) != "" {
			filtered = append(filtered, claim)
		}
	}
	c.ProtectedClaims = filtered

	return nil
}

func (c *Config) compileRiskPaths() error {
	if len(c.RiskPaths) == 0 {
		return nil
	}
	c.riskPathGlobs = make(map[string][]glob.Glob, len(c.RiskPaths))
	for category, patterns := range c.RiskPaths {
		compiled := make([]glob.Glob, 0, len(patterns))
		for _, pattern := range patterns {
			g, err := glob.Compile(pattern, '/')
			if err != nil {
				return fmt.Errorf("invalid risk_paths pattern %q in category %q: %w", pattern, category, err)
			}
			compiled = append(compiled, g)
		}
		c.riskPathGlobs[category] = compiled
	}
	return nil
}

func (c *Config) compileRedaction() error {
	if len(c.Redaction.Patterns) == 0 {
		return nil
	}
	c.redactionRegex = make([]*regexp.Regexp, 0, len(c.Redaction.Patterns))
	for _, pattern := range c.Redaction.Patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid redaction pattern %q: %w", pattern, err)
		}
		c.redactionRegex = append(c.redactionRegex, re)
	}
	return nil
}
