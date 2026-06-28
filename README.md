# Intent Diff

> Compare what a PR says it does with what the diff actually changes.

Intent Diff extracts claimed intent from a PR description and compares it with implementation evidence from the git diff. It produces a structured mismatch report with a Grade A–E scale and an attention map, helping reviewers decide where to focus before reading every changed line.

## How it works

Intent Diff runs a 3-stage pipeline:

1. **Collect** — Parse the git diff and PR description. Classify files by role (source, test, config, docs, etc.) and tag risk levels (auth, api, data, infra). Truncate large diffs by risk priority.
2. **Analyze** — Send the structured prompt to Claude CLI (`claude --bare -p`) with a JSON schema. The LLM compares claimed intent against implementation evidence and returns a structured analysis.
3. **Render** — Format the analysis as Markdown (human-readable) or JSON (machine-readable) with Grade badge, Attention Map, Mismatch taxonomy, and validation warnings.

## Installing / Getting started

```shell
go install github.com/yottayoshida/intent-diff/cmd/intent-diff@latest
```

Requires [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code/overview) (`claude` command in PATH) and a git repository.

## Examples

### Basic usage

```shell
# Analyze current branch against main using gh CLI
gh pr view --json title,body | intent-diff analyze --pr-json /dev/stdin

# Use a markdown file as PR description
intent-diff analyze --intent pr-description.md

# Output as JSON
intent-diff analyze --pr-json pr.json --json --out report.json

# Custom base/head refs
intent-diff analyze --base origin/develop --head feature/auth --intent desc.md

# Use a pre-generated diff file
intent-diff analyze --diff-file changes.diff --intent desc.md
```

### Sample output

```
# Intent Diff Report

**Grade: C** — Material omissions (confidence: high, score: 0.45)

> **Partial analysis**: The diff exceeded the analysis budget (100000 chars).
> 42 of 58 files were analyzed; 16 file(s) were excluded or truncated.
> To analyze the full diff, increase `max_diff_size` in `.intent-diff.yml`.

## Claimed Intent

Refactor auth middleware for readability

## Attention Map

| Priority | File | Reason |
|----------|------|--------|
| high | `auth/session.go` | Session expiry logic changed |
| medium | `api/handler.go` | New error code added |

## Mismatches (1)

### 1. [contract] Auth session expiry changed (severity: high, confidence: high)

**Observation**: Session timeout reduced from 24h to 1h
**Evidence**: auth/session.go
**Recommended action**: Update PR description to disclose session expiry change
```

## Features

* **9-category mismatch taxonomy** — scope, contract, risk, test, intent_under_specification, non_code_impact, behavioral_ambiguity, documentation, dependency_risk
* **Grade A–E** — from well-aligned to critical mismatches
* **Attention map** — prioritized list of files to review first
* **Structured output** — Markdown for humans, JSON for CI
* **Smart diff handling** — file classification (8 categories), risk-based prioritization (5 levels), truncation for large diffs with partial analysis warnings
* **Prompt injection defense** — XML data boundaries, `--json-schema` enforcement, post-hoc file path validation
* **Configurable timeout** — `--timeout` flag and config file support with 30s–30min bounds

## GitHub Action

Add intent-diff to your PR workflow in two steps:

1. Add `ANTHROPIC_API_KEY` to your repository secrets
2. Add the workflow:

```yaml
name: Intent Diff
on:
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: read

jobs:
  intent-diff:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for intent-diff

      - uses: yottayoshida/intent-diff@v0
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
```

### Action inputs

| Input | Default | Description |
|-------|---------|-------------|
| `config-path` | `.intent-diff.yml` | Path to config file |
| `version` | `latest` | intent-diff version to install (e.g. `v0.1.3`) |
| `claude-code-version` | `1` | Claude Code CLI version |
| `soft-fail` | `true` | Exit 0 on analysis errors. Grade threshold failures always exit 1 |

### Action outputs

| Output | Description |
|--------|-------------|
| `grade` | Alignment grade (A–E) |
| `score` | Alignment score (0–1) |
| `has-mismatches` | `"true"` if mismatches were detected |
| `report-json` | Path to JSON report file |

### Fork PRs

Fork PRs do not have access to repository secrets. When `ANTHROPIC_API_KEY` is not set, the action skips analysis and exits 0. Use the `pull_request` event only — `pull_request_target` is not recommended as it exposes secrets to fork code.

## Configuration

Create `.intent-diff.yml` in your project root:

```yaml
ignore:
  - "**/*.generated.go"
  - "vendor/**"
max_diff_size: 100000
output_format: markdown
timeout: "5m"
output_mode: check_summary
risk_paths:
  auth:
    - "internal/auth/**"
  api:
    - "api/v2/**"
protected_claims:
  - "backward compatible"
thresholds:
  fail_on_grade: "D"
redaction:
  patterns:
    - "sk-[a-zA-Z0-9]{32,}"
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ignore` | `[]string` | `[]` | Glob patterns to exclude from analysis |
| `max_diff_size` | `int` | `100000` | Maximum diff size in characters |
| `output_format` | `string` | `"markdown"` | `"markdown"` or `"json"` |
| `timeout` | `string` | `"5m"` | Analysis timeout (e.g. `"2m"`, `"10m"`); range: 30s–30m |
| `output_mode` | `string` | `"local"` | Output destination: `"local"` or `"check_summary"` |
| `risk_paths` | `map[string][]string` | `{}` | Custom risk categories with glob patterns |
| `protected_claims` | `[]string` | `[]` | PR description claims to flag when contradicted |
| `thresholds.fail_on_grade` | `string` | `""` | Exit 1 if grade is at or below threshold (`"C"`, `"D"`, or `"E"`) |
| `redaction.patterns` | `[]string` | `[]` | Regex patterns to redact from diff before LLM analysis |

All config fields are optional. Existing `.intent-diff.yml` files with only v0.1 fields continue to work unchanged.

CLI flags `--config`, `--output-mode`, and `--fail-on-grade` override the corresponding config file values.

## Troubleshooting

**Analysis times out**: Large diffs take longer to analyze. Increase the timeout with `--timeout 10m` or reduce diff size with `ignore` patterns in `.intent-diff.yml`.

**Partial analysis warning**: When the diff exceeds `max_diff_size`, lower-risk files are excluded. The report shows how many files were analyzed. Increase `max_diff_size` or add `ignore` patterns for generated/vendor files.

**`claude` command not found**: Intent Diff requires [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code/overview) installed and available in PATH. See [docs/ci.md](docs/ci.md) for CI environment setup.

**Low confidence grades**: The LLM analysis is probabilistic. Low confidence often means the PR description is vague or the diff is very large. Consider improving the PR description or using `--force` for minimal diffs.

## Limitations

- **Large diffs are truncated**: Diffs exceeding `max_diff_size` are truncated by risk priority. High-risk files (auth, API) are kept; low-risk files (docs, generated) are excluded first. The report discloses when truncation occurred.
- **LLM-dependent**: Results are hypotheses based on LLM analysis, not verified facts. The post-hoc validator catches some hallucinations (e.g., file paths not in the diff), but human judgment is still required.
- **No incremental analysis**: Each run analyzes the full diff. There is no caching or delta analysis between runs.
- **Single LLM backend**: Currently requires Claude Code CLI. No API key mode or alternative LLM backends (see [ADR-0001](docs/adr/0001-cli-runner-strategy.md)).

## Developing

```shell
git clone https://github.com/yottayoshida/intent-diff.git
cd intent-diff
go build ./cmd/intent-diff
go test -race ./...
```

## Links

- Repository: https://github.com/yottayoshida/intent-diff
- Issue tracker: https://github.com/yottayoshida/intent-diff/issues

## License

Licensed under either of

- Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
- MIT license ([LICENSE-MIT](LICENSE-MIT) or http://opensource.org/licenses/MIT)

at your option.
