# PRD Feedback Decision Log

作成日: 2026-06-20  
対象: `docs/PRD.ja.md` 初稿への批評FB

## Summary

今回のFBは、おおむね正しい。ただし「思想を削る」方向ではなく、「初期ICP・UX・評価・用語」を製品として勝てる形に絞る判断として採用した。

最重要の決定:

- 中心価値を「PRの説明責任」から「レビュー注意配分」に寄せる。
- 初期ICPを「AI-generated PRをレビューする小-中規模チーム」に絞る。
- `Observed behavior` という断定的な語をやめ、`Implementation Evidence` と `Behavior-impact Hypotheses` に分ける。
- v0.1の必須入力をPR descriptionとgit diffに絞る。
- 継続利用UXはPRコメントではなくGitHub Checks summaryを本命にする。

## Decisions

| FB | 判断 | 理由 | PRD反映 |
| --- | --- | --- | --- |
| 思想は強いが購買理由・継続理由が弱い | 採用 | 導入判断は思想ではなく、レビュー時間・見落とし・ノイズ削減で行われる | 市場調査に購買・継続利用仮説を追加 |
| 価値の中心は「PR説明の信用度」ではなく「レビュー注意配分」 | 採用 | ユーザーの実際の問いは「どこから読むべきか」に近い | Executive Summary、Mission、Problemを変更 |
| `Observed behavior` は言い過ぎ | 採用 | diffだけでruntime behaviorは確定できない。高信頼を売るなら断定を避けるべき | `Implementation Evidence` / `Behavior-impact Hypotheses` に変更 |
| mismatch taxonomyをもっと明確に切る | 採用 | 文章生成ではなく分類体系が製品の防衛線になる | 9分類のtaxonomy表を追加 |
| ICPをAI-generated PRレビュー小-中規模チームに絞る | 限定採用 | 初期訴求は絞るべき。一方で人間PR対応は副次価値として残す | Target UsersでICP決定を明記 |
| 競合との差分が概念的すぎる | 採用 | ユーザーは出力面では既存AIコメントと比較する | 実務的競合比較表を追加 |
| v0.1入力が多すぎる | 採用 | PR説明+diffで価値が出ないなら、入力追加で芯を補っても弱い | v0.1必須入力をPR description + git diffに限定 |
| 成功指標が足りない | 採用 | false positive costを測らないと「うるさいツール」になる | useful attention shift、noisy output、PR説明修正率などを追加 |
| Local-firstは強いが継続利用UXとして弱い | 採用 | CLIは試用・評価には強いが、レビュー文脈はGitHub上にある | CLIを実験装置、Checksを本命UXと位置づけ |
| PRコメントよりChecks summary中心がよい | 採用 | 低ノイズ思想と一致する。PRコメントは通知・会話欄汚染リスクが高い | Phase 1でChecks default、comment opt-inに変更 |
| 「説明責任」は重い | 限定採用 | v0.1訴求としては硬いが、Enterprise/監査価値としては残すべき | messagingでattention mapを前面化、説明責任はEnterprise向けに降格 |
| 既存AIレビューに吸収されるリスクが最大 | 採用 | 独立製品の防衛線を明確化する必要がある | 市場リスクと競合比較にmoatを追加 |

## Explicit Non-Decisions

- Alignment Scoreを廃止するかは未決定。A-Eが直感的か、risk labelだけで十分かをv0.1評価で見る。
- 実装言語は未決定。配布性を考えるとRust/Goが有力だが、研究速度を優先するならPythonもあり得る。
- agent session summaryは重要だが、v0.1では後回し。PR description + diffの芯が成立してから扱う。
- GitHub AppはPhase 2。Phase 1ではGitHub Action + Checks summaryで十分に検証できる。

## Revised Product Center

Intent Diff は、PRを読む前に、レビュアーが「何を信じ、何を疑い、どこから読むべきか」を決めるためのレビュー・トリアージ層である。

勝ち筋は、AIコードレビューbotになることではない。mismatch taxonomy、低ノイズChecks UX、チーム固有ルール、評価データセット、privacy postureを組み合わせて、AI-generated PR時代のレビュー注意配分を支えることである。
