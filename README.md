# Intent Diff

A review-triage tool that compares what a PR says it does with what the diff actually changes.

Intent Diff extracts claimed intent from a PR description and compares it with implementation evidence from the git diff, producing a structured mismatch report with a 9-category taxonomy. It helps reviewers decide where to focus attention before reading every changed line.

## Quick Start

```bash
# Install
go install github.com/yottayoshida/intent-diff/cmd/intent-diff@latest

# Analyze current branch against main
intent-diff analyze --intent pr-description.md

# Use gh CLI output as input
gh pr view --json title,body > pr.json
intent-diff analyze --pr-json pr.json

# Pipe PR description from stdin
echo "Refactor auth middleware" | intent-diff analyze

# Use a pre-generated diff file
intent-diff analyze --intent pr.md --diff-file changes.patch

# Output as JSON
intent-diff analyze --intent pr.md --json

# Save to file
intent-diff analyze --intent pr.md --out report.md
```

## How It Works

1. **Collect**: Reads PR description and git diff, classifies files (source/test/config/docs/generated/vendor/lockfile), tags risk levels (auth > api > data > infra > other), and truncates to fit within the analysis budget.
2. **Analyze**: Sends structured prompt to Claude via `claude --bare -p` with JSON schema enforcement. Produces a 9-category mismatch analysis.
3. **Render**: Outputs human-readable Markdown or machine-readable JSON.

## Mismatch Taxonomy

| Category | Description |
|----------|-------------|
| `scope` | Changes exceed or fall short of claimed scope |
| `contract` | API contracts, interfaces, or type signatures changed without mention |
| `risk` | Security, auth, or data-handling changes not documented |
| `test` | Test changes that don't match claimed testing scope |
| `intent_under_specification` | PR description too vague to assess alignment |
| `non_code_impact` | Config, infra, or deployment changes not mentioned |
| `behavioral_ambiguity` | Changes with unclear behavioral impact |
| `documentation` | Documentation doesn't reflect implementation |
| `dependency_risk` | Dependency changes with unmentioned implications |

## Grade Scale

| Grade | Meaning |
|-------|---------|
| A | Well-aligned: no material mismatches |
| B | Minor omissions: small gaps but no risk-bearing changes undocumented |
| C | Material omissions: risk-bearing changes not mentioned |
| D | Significant mismatches: explicit claims contradicted by evidence |
| E | Critical mismatches: safety/auth/data changes undocumented |

## Configuration

Create `.intent-diff.yml` in your project root:

```yaml
# Glob patterns to exclude from analysis
ignore:
  - "**/*.generated.go"
  - "vendor/**"

# Maximum diff size in characters (default: 100000)
max_diff_size: 100000

# Output format: "markdown" or "json" (default: "markdown")
output_format: markdown
```

## Requirements

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code/overview) (`claude` command in PATH)
- Git repository

## CLI Reference

```
intent-diff analyze [flags]

Flags:
  --base string        base ref for git diff (default: merge-base with main)
  --head string        head ref for git diff (default: HEAD)
  --diff-file string   path to a pre-generated diff file
  --intent string      path to a PR description markdown file
  --pr-json string     path to gh pr view --json title,body output
  --out string         output file path (default: stdout)
  --json               output as JSON instead of Markdown
  --force              force LLM analysis even for minimal diffs
```

## Development

```bash
# Build
go build ./cmd/intent-diff

# Test
go test -race ./...

# Lint
golangci-lint run
```

## License

MIT
