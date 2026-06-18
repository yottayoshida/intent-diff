# Intent Diff

A review-orientation tool that compares what a PR says it is doing with what the diff actually changes.

## Core Idea

Traditional diffs show file changes. AI PR reviewers look for bugs, risks, and suggested improvements. Intent Diff sits one step earlier: it extracts the claimed intent from the PR description, linked issue, original prompt, commit messages, or agent session summary, then compares that intent with the actual behavioral changes implied by the diff.

For example, it should catch cases where a PR claims to be "refactor only" but changes authentication behavior, or where a "UI copy update" also changes an API contract. The goal is to help reviewers notice mismatches before they spend time reading every changed line.

## Why Now

- As AI agents produce PRs faster, reviewers need better orientation before diving into the diff.
- Existing AI PR reviewers mostly focus on defect detection, security issues, or improvement suggestions.
- The missing layer is alignment between the intended work and the implemented result.

## Differentiation

- Not a code review bot: a review-orientation tool.
- Not a semantic diff: an intent-vs-behavior diff.
- Not primarily a security warning system: a daily attention-allocation aid for reviewers.
- Useful for both AI-generated and human-authored PRs.

## Inputs

- PR title and description.
- Linked issue or task.
- Agent prompt or session summary when available.
- Commit messages.
- Git diff.
- Optional local repo context.

## Outputs

- Claimed intent.
- Observed behavior changes.
- Alignment score.
- Suspicious mismatches.
- Reviewer focus areas.
- Suggested PR description corrections.

## Example

Claimed intent:

- Refactor auth middleware without behavior changes.

Observed changes:

- Token expiry handling changed.
- Logout path now clears session earlier.
- Error response status changed from 401 to 403 in one branch.

Potential mismatch:

- This is not a pure refactor. Review auth edge cases and update PR description.

## First Prototype

1. CLI reads current git diff and PR description markdown.
2. LLM produces intent, observed behavior, mismatch list.
3. Output as local markdown.
4. Later: GitHub Action that comments on PRs.

## Open Questions

- Can AST diff tools such as Difftastic provide structured signals before calling an LLM?
- Should the tool require an explicit PR template, or infer intent from messy text?
- What mismatch categories are useful enough for daily review?
- Can it be made cheap enough to run on every local diff before opening a PR?
