# Intent Diff

> Compare what a PR says it does with what the diff actually changes.

Intent Diff extracts claimed intent from a PR description and compares it with implementation evidence from the git diff. It produces a structured mismatch report with a Grade A–E scale and an attention map, helping reviewers decide where to focus before reading every changed line.

## Installing / Getting started

```shell
go install github.com/yottayoshida/intent-diff/cmd/intent-diff@latest
```

Requires [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code/overview) (`claude` command in PATH) and a git repository.

```shell
# Analyze current branch against main
gh pr view --json title,body | intent-diff analyze --pr-json /dev/stdin

# Or use a markdown file as PR description
intent-diff analyze --intent pr-description.md

# Output as JSON
intent-diff analyze --pr-json pr.json --json --out report.json
```

## Features

* **9-category mismatch taxonomy** — scope, contract, risk, test, intent_under_specification, non_code_impact, behavioral_ambiguity, documentation, dependency_risk
* **Grade A–E** — from well-aligned to critical mismatches
* **Attention map** — prioritized list of files to review first
* **Structured output** — Markdown for humans, JSON for CI
* **Smart diff handling** — file classification, risk-based prioritization, truncation for large diffs
* **Prompt injection defense** — XML data boundaries, `--json-schema` enforcement, post-hoc file path validation

## Configuration

Create `.intent-diff.yml` in your project root:

```yaml
ignore:
  - "**/*.generated.go"
  - "vendor/**"
max_diff_size: 100000
output_format: markdown
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `ignore` | `[]string` | `[]` | Glob patterns to exclude from analysis |
| `max_diff_size` | `int` | `100000` | Maximum diff size in characters |
| `output_format` | `string` | `"markdown"` | `"markdown"` or `"json"` |

## Developing

```shell
git clone https://github.com/yottayoshida/intent-diff.git
cd intent-diff
go build ./cmd/intent-diff
go test -race ./...
```

## Contributing

If you'd like to contribute, please fork the repository and use a feature branch. Pull requests are warmly welcome.

## Links

- Repository: https://github.com/yottayoshida/intent-diff
- Issue tracker: https://github.com/yottayoshida/intent-diff/issues

## Licensing

The code in this project is licensed under [MIT license](LICENSE).
