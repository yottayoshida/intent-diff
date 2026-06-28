# Changelog

All notable changes to this project will be documented in this file.

## [0.1.3] - 2026-06-28

**Summary**: GitHub Action foundation — composite action, Checks summary renderer, output mode routing, exit code routing, and config schema extension for Phase 1 features.

### Added

- `action.yml` composite GitHub Action for PR workflows with `pull_request` event support
- `RenderChecksSummary` renderer with 3-layer compact layout (Grade → Attention Map max 5 → Mismatches with `<details>` folding)
- Output mode routing: `--output-mode` CLI flag > `output_mode` config > `GITHUB_STEP_SUMMARY` env auto-detection > `"local"` default
- `ExitError` type with distinct exit codes: 0 (success), 1 (analysis/threshold), 2 (config error)
- `FailOnGrade` integration: exit 1 when alignment grade meets or exceeds `thresholds.fail_on_grade` threshold
- Config Phase 1 fields: `output_mode`, `risk_paths`, `protected_claims`, `thresholds.fail_on_grade`, `redaction.patterns`
- `Validate()` method with enum checking, glob/regex compilation, BOM strip, trailing whitespace trimming
- `--config` CLI flag for config file path override (errors on missing file when explicitly set)
- `--fail-on-grade` CLI flag (C, D, or E threshold)
- Fork PR graceful skip in action.yml (exit 0 when `ANTHROPIC_API_KEY` not set)
- Soft-fail logic in action.yml: analysis errors exit 0 by default, config errors always exit 2
- `validateRef()` allowlist (alphanumeric + `._/~^-`) replacing simple prefix check
- `verifyRef()` fetch-depth guard for both base and head refs
- `ValidOutputMode()` and `ValidFailGrade()` exported helpers
- Golden file tests for Checks summary (clean + mismatch outputs)
- 71 new tests (111 → 182 total)

### Changed

- Config errors (YAML parse failure, invalid enums, missing `--config` file) now exit 2 instead of 1
- `GradeDescription()` extracted as shared function (was duplicated between renderers)
- `writeToFile()` now respects `--json` / `output_format` instead of always writing JSON

## [0.1.2] - 2026-06-25

**Summary**: Large input resilience — partial-analysis warnings, configurable timeout, CI supply-chain hardening, and expanded documentation.

### Added

- Partial-analysis warning in Markdown output when files are excluded by budget truncation or category filtering, showing separate counts for excluded and truncated files
- `truncated`, `truncated_files`, `excluded_files`, `files_analyzed`, `files_total` fields in JSON output (backward-compatible via `omitempty`)
- `RenderMetadata` struct separating collect-stage facts from LLM schema
- `--timeout` CLI flag with configurable analysis timeout (range: 30s–30m, default: 5m)
- `timeout` field in `.intent-diff.yml` configuration with precedence: CLI flag > config > default
- 3-layer timeout error messages (what happened, current state, recovery action)
- Dependabot configuration for automated dependency updates (github-actions weekly, gomod weekly)
- README sections: How it works, Examples, Configuration (timeout), Troubleshooting, Limitations
- 15 new tests (96 → 111 total)

### Changed

- All GitHub Actions pinned to immutable commit SHAs (eliminates tag-hijack supply-chain risk)
- PRD reconciled with v0.1.1 implementation reality (file count, status, technology confirmation)

### Fixed

- Partial-analysis warning now shown when files excluded by category (vendor/generated) even if budget not exceeded
- Truncated files no longer double-counted as both "analyzed" and "excluded" in warning message
- Slice backing-array aliasing between `truncationExcluded` and `allExcluded` prevented with explicit allocation

## [0.1.1] - 2026-06-21

**Summary**: CI runner strategy established — Claude Code CLI as the sole runner for both local and CI environments.

### Changed

- JSON schema version field from `const` to `enum` for broader structured-output tool compatibility
- PreflightCheck error messages now include 3-layer guidance: what happened, current state, and what to do (with CI-specific install instructions)

### Added

- ADR-0001 documenting the CLI runner strategy decision (no Anthropic Go SDK, no API billing)
- CI setup documentation (`docs/ci.md`) with GitHub Actions workflow example, failure modes, and security notes
- 8 new tests for schema validation and PreflightCheck error messages

## [0.1.0] - 2026-06-20

**Summary**: Initial release — PR description vs git diff structured comparison CLI.

### Added

- `intent-diff analyze` command with 3-stage pipeline (Collect → Analyze → Render)
- 9-category mismatch taxonomy: scope, contract, risk, test, intent_under_specification, non_code_impact, behavioral_ambiguity, documentation, dependency_risk
- Grade A–E alignment scale with confidence scoring
- Attention map with prioritized file-level review guidance
- PR description input via `--intent` (markdown file), `--pr-json` (gh CLI output), or stdin
- Diff input via git refs (`--base`/`--head`) or pre-generated patch file (`--diff-file`)
- File classification (source, test, config, docs, generated, vendor, binary, lockfile)
- Risk-based file prioritization (auth > api > data > infra > other)
- Diff truncation with configurable character budget (default 100K chars)
- Markdown and JSON output formats
- `.intent-diff.yml` configuration file support (ignore patterns, max diff size, output format)
- LLM integration via `claude --bare -p` with `--json-schema` enforcement
- Prompt injection defense: XML data boundaries, JSON schema constraint, post-hoc file path validation
- Pre-flight check for Claude Code CLI availability
- Empty/copy-paste PR description detection with diff-only fallback
- Minimal diff shortcut (≤5 lines → Grade A without LLM call)
- goreleaser configuration for cross-platform binary distribution

[0.1.3]: https://github.com/yottayoshida/intent-diff/releases/tag/v0.1.3
[0.1.2]: https://github.com/yottayoshida/intent-diff/releases/tag/v0.1.2
[0.1.1]: https://github.com/yottayoshida/intent-diff/releases/tag/v0.1.1
[0.1.0]: https://github.com/yottayoshida/intent-diff/releases/tag/v0.1.0
