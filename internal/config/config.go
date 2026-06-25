package config

import (
	"fmt"
	"os"
	"time"

	"github.com/gobwas/glob"
	"go.yaml.in/yaml/v3"
)

const (
	DefaultTimeout = 5 * time.Minute
	MinTimeout     = 30 * time.Second
	MaxTimeout     = 30 * time.Minute
)

// Config represents the .intent-diff.yml configuration file.
type Config struct {
	Ignore       []string `yaml:"ignore"`
	MaxDiffSize  int      `yaml:"max_diff_size"`
	OutputFormat string   `yaml:"output_format"`
	Timeout      string   `yaml:"timeout"`

	ignoreGlobs []glob.Glob
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

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MaxDiffSize <= 0 {
		cfg.MaxDiffSize = 100_000
	}

	if err := cfg.compileIgnores(); err != nil {
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
