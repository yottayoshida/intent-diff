# Intent Diff

PRやAIエージェントの作業について、「言っていること」と「実際に変えたこと」のズレを見つけるレビュー補助ツール。

## Core Idea

普通のdiffはファイル差分を見る。AI PR reviewerはバグや改善点を見る。Intent Diffはその前に、PR本文、issue、依頼プロンプト、エージェント作業ログから「意図」を抽出し、実際の差分がその意図に沿っているかを見る。

たとえば、PRが「リファクタのみ」と言っているのに認証挙動を変えている、または「UI文言修正」と言っているのにAPI契約を変えている、といったズレを日常レビューで早く見つける。

## Why Now

- AIエージェントがPRを高速に作るほど、レビュー担当者は「何を見ればいいか」を失いやすい。
- 既存のAI PR reviewerは多いが、多くは欠陥検出や改善提案に寄っている。
- これから必要になるのは、差分そのものより「作業意図と実装結果の整合性」を見るレイヤ。

## Differentiation

- Code review botではなく、review orientation tool。
- Semantic diffではなく、intent-vs-behavior diff。
- セキュリティ警告ではなく、日常レビューの注意配分を助ける。
- AI生成PRだけでなく、人間のPRにも使える。

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

