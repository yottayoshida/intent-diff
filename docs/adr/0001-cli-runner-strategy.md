# ADR-0001: CLI runner strategy for CI environments

- **Status**: Accepted
- **Date**: 2026-06-20

## Context

intent-diff v0.1.0 uses `ExecClaudeRunner` to call `claude --bare -p --output-format json --json-schema` as a subprocess. This works locally when Claude Code CLI is installed and authenticated via a Max subscription. Issue #8 asks how intent-diff should run in CI (GitHub Actions) where there is no interactive session.

Two approaches were considered: (1) add an `APIRunner` using the Anthropic Go SDK (`anthropic-sdk-go`) to call the Messages API directly with an API key, or (2) reuse the existing CLI runner in CI by having users install Claude Code CLI and authenticate via `ANTHROPIC_API_KEY`.

## Decision

Use **Claude Code CLI in CI** as the sole runner strategy. Do not add an `APIRunner` or the Anthropic Go SDK as a dependency.

CI users install Claude Code CLI (`npm install -g @anthropic-ai/claude-code`) in their workflow and set `ANTHROPIC_API_KEY` as a GitHub Actions secret. The existing `ExecClaudeRunner` handles both local and CI execution without code changes.

## Alternatives Considered

| Alternative | Reason for rejection |
|-------------|---------------------|
| Anthropic Go SDK (`APIRunner`) | Requires API billing separate from Max plan. Adds a dependency and second code path to maintain. No user benefit over CLI runner. |
| `--api-key` CLI flag | Security risk: API keys visible in `ps` output and shell history. Claude Code CLI already reads `ANTHROPIC_API_KEY` from the environment. |
| Dual runner with auto-detection | Unnecessary complexity. One runner covers both environments. |

## Consequences

- **No new Go dependency**: `go.mod` stays minimal (cobra, go-gitdiff, glob, yaml).
- **Single code path**: `ExecClaudeRunner` is the only runner implementation, reducing maintenance burden.
- **CI users need Node.js**: Claude Code CLI requires `npm` to install, which is available by default on GitHub Actions runners.
- **Future flexibility preserved**: The `ClaudeRunner` interface allows adding an API runner later if requirements change.
- **No API billing**: Users run intent-diff under their existing Claude Code subscription or API key without a separate billing relationship.
