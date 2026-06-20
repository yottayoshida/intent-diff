# Changelog

All notable changes to this project will be documented in this file.

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

[0.1.0]: https://github.com/yottayoshida/intent-diff/releases/tag/v0.1.0
