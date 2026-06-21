# Running intent-diff in CI

intent-diff uses Claude Code CLI (`claude`) as its analysis backend. In CI environments like GitHub Actions, you need to install the CLI and provide an API key.

## Prerequisites

- **Node.js 18+** (pre-installed on GitHub Actions runners)
- **An Anthropic API key** with access to Claude
- **Public repository** (for `go install`). For private repos, use `actions/checkout` + `go build` instead

## GitHub Actions setup

### 1. Add the API key as a secret

Go to your repository's **Settings → Secrets and variables → Actions** and add a secret named `ANTHROPIC_API_KEY`.

### 2. Add a workflow step

```yaml
- name: Install intent-diff and Claude Code CLI
  run: |
    npm install -g @anthropic-ai/claude-code
    go install github.com/yottayoshida/intent-diff/cmd/intent-diff@latest

- name: Run intent-diff analysis
  env:
    ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  run: |
    gh pr view ${{ github.event.pull_request.number }} --json title,body \
      | intent-diff analyze --pr-json /dev/stdin
```

### 3. Full workflow example

```yaml
name: PR Intent Check
on:
  pull_request:
    types: [opened, synchronize, edited]

jobs:
  intent-diff:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: read
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      - name: Install tools
        run: |
          npm install -g @anthropic-ai/claude-code
          go install github.com/yottayoshida/intent-diff/cmd/intent-diff@latest

      - name: Analyze PR
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
          GH_TOKEN: ${{ github.token }}
        run: |
          gh pr view ${{ github.event.pull_request.number }} --json title,body \
            | intent-diff analyze --pr-json /dev/stdin --json --out report.json
```

## How it works

intent-diff calls `claude --bare -p --output-format json --json-schema <schema>` as a subprocess. Claude Code CLI reads `ANTHROPIC_API_KEY` from the environment automatically — intent-diff never handles the key directly.

## Failure modes

| Symptom | Cause | Fix |
|---------|-------|-----|
| `claude CLI not found in PATH` | CLI not installed | Add `npm install -g @anthropic-ai/claude-code` before running intent-diff |
| `claude --version failed` | Broken CLI installation | Reinstall with `npm install -g @anthropic-ai/claude-code` |
| Claude CLI returns auth error | Missing or invalid API key | Check that `ANTHROPIC_API_KEY` secret is set and valid |
| Analysis timeout (5 min default) | Large diff or slow response | Reduce diff size with `max_diff_size` in `.intent-diff.yml`, or add `ignore` patterns |

## Security notes

- **Never** pass the API key as a command-line argument. Use the `ANTHROPIC_API_KEY` environment variable.
- intent-diff does not read, log, or propagate the API key. It is consumed solely by Claude Code CLI.
- Use GitHub Actions secrets (not plaintext in workflow files) to store the key.

## See also

- [ADR-0001: CLI runner strategy](adr/0001-cli-runner-strategy.md) — why intent-diff uses CLI instead of a direct API client
- [Claude Code CLI docs](https://docs.anthropic.com/en/docs/claude-code/overview)
