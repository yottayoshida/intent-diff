# Intent Diff PRD

作成日: 2026-06-20  
対象リポジトリ: <https://github.com/yottayoshida/intent-diff>  
ステータス: v0.1.1 released — Phase 1 in progress
改訂メモ: 2026-06-20 の批評FBを受け、中心価値を「説明責任」から「レビュー注意配分」に寄せて再整理

## 0. エグゼクティブサマリー

Intent Diff は、Pull Request を読む前に、レビュアーが「何を信じ、何を疑い、どこから読むべきか」を決めるためのレビュー・トリアージツールである。

既存の差分ツールは「どの行が変わったか」を示す。既存のAIコードレビューは「バグ・セキュリティ・改善点」を探す。Intent Diff はその一段手前で、PRが主張している意図と、diffから得られる実装上の証拠を照合し、レビュー注意資源の配分を助ける。

中核仮説:

- AI coding agent によりPR作成量とPR説明の流暢さが増えるほど、レビューのボトルネックは「コードを読む時間」だけでなく「どの前提を信じてよいか」「どこを深く読むべきか」に移る。
- 初期の最重要ユーザーは、AI-generated PRを日常的にレビューする小-中規模チームのTech Lead / senior reviewerである。人間PRにも使えるが、v0.1の緊急性はAI-generated PRに寄せる。
- ユーザーが継続利用する条件は、検出数ではない。レビュー開始前の理解時間が短くなり、説明にない危険な変更を早く見つけ、不要な通知やコメントが少ないことである。
- 「意図と実装証拠のズレ」は、セキュリティスキャナ、静的解析、AIレビュー、PR summary、構文diffのどれとも重なるが、完全には代替されない独立した製品面である。

最初の勝ち筋:

- v0.1: ローカルCLIで、必須入力をPR説明とgit diffに絞り、「Claimed intent / Implementation evidence / Mismatch taxonomy / Attention map」をMarkdownとJSONで出す。CLIは本命UXではなく、出力品質と評価データ収集のための実験装置と位置づける。
- Phase 1: GitHub Action と設定ファイルを追加し、デフォルトはGitHub Checks summaryとして、レビュアーの読む順序と確認ポイントを提示する。PRコメントはhigh-confidence mismatchのみopt-inにする。
- Phase 2 / v1.0: GitHub App、組織ルール、履歴学習、評価基盤、セルフホスト対応を備えた、チームのレビュー信頼性レイヤーにする。

製品焦点:

| 問い | 決定 |
| --- | --- |
| 誰が使うか | AI-generated PRをレビューする小-中規模チームのTech Lead / senior reviewer |
| どの瞬間に使うか | PRを詳細に読む前、レビュー開始時、またはPR作成前のpreflight |
| 何の代替か | PR説明を信じてdiffを読み始めること、またはAIレビューbotの大量コメントを読むこと |
| 何が改善されたら継続利用するか | 読む順序の判断が早くなる、説明にない高リスク変更に気づく、PR説明修正や追加確認が早まる、通知ノイズが増えない |

## 1. 現状コンテキスト

v0.1.1 がリリース済み。Go 1.26 CLI として53ファイル、96テストで実装されている。3-stage pipeline (Collect → Analyze → Render) が動作し、Markdown/JSON出力、ファイル分類、リスクタグ、トランケーション、バリデーションが機能している。

- 説明: Intent-vs-behavior diff tool for reviewing AI and human pull requests
- コアアイデア: PR説明、linked issue、元プロンプト、commit message、agent session summary から主張された意図を抽出し、git diff が示す振る舞いの変化と比較する。
- 例: 「refactor only」と言いながら認証挙動が変わる、「UI copy update」と言いながらAPI契約が変わる。
- 初期プロトタイプ案: CLIがgit diffとPR description markdownを読み、LLMがintent、implementation evidence、mismatch listを生成し、Markdown出力する。

このPRDでは、READMEの種をそのまま拡張するだけでなく、以下の製品判断を加える。

- 「Observed behavior」という断定的な語は避け、diffからの根拠付き推定として扱う。
- 初期ICPは「AI-generated PRをレビューする小-中規模チーム」に絞る。
- v0.1はPR説明とgit diffだけで価値が出るかを検証する。agent prompt/session summaryは後回しにする。
- 継続利用の本命UXはGitHub Checks summaryであり、PRコメントは慎重に使う。
- 勝ち筋はLLM文章生成ではなく、mismatch taxonomy、低ノイズUX、チームルール、評価データセットに置く。

## 2. 製品哲学

### 2.1 ミッション

PRを読む前に、レビュアーの注意配分を最適化する。

Intent Diff の目的は「AIにレビューを任せること」ではない。人間のレビュアーが、限られた注意を最も危険なズレに使えるようにすることである。「説明責任の可視化」はv1.0以降の組織・監査向け価値として残すが、v0.1の採用メッセージでは「どこから読むべきか」を前面に出す。

### 2.2 プロダクトの立ち位置

Intent Diff はコードレビューbotではない。レビューの前処理である。

- コードレビューbot: 「この行にバグがあります」と言う。
- 静的解析: 「このルールに違反しています」と言う。
- 構文diff: 「この式や構造が変わりました」と言う。
- Intent Diff: 「このPRが主張している作業範囲と、diffから見える実装証拠がズレている可能性があります。ここから読んでください」と言う。

### 2.3 世界観

1. Intent is an interface  
   PR description、issue、タスク、agent prompt は、実装とレビュアーの間のインターフェースである。このインターフェースが壊れると、レビューは深く読む前から危険になる。

2. Diff is not behavior, but it is evidence  
   git diff は振る舞いそのものではない。テスト、実行、型、依存関係、runtime contextなしに挙動は確定できない。しかし、振る舞い・契約・運用・セキュリティ境界の変化を推測する強い証拠である。したがってIntent Diffは断定ではなく、根拠付きレビュー仮説を出す。

3. The reviewer remains sovereign  
   Intent Diff は合否判定者ではない。レビュアーに「どこから読むべきか」「説明を修正すべきか」「追加テストが必要そうか」を渡す。

4. Low noise is a product feature  
   AIレビュー市場ではコメント過多が疲労を生む。Intent Diff は、少ない出力、根拠付き、PR全体サマリー中心、必要時のみ詳細、を基本とする。

5. Local-first trust  
   初期ユーザーは、未公開コードや顧客データを扱う。最初からローカル実行、JSON出力、モデル選択、除外パターン、ログ抑制を設計に含める。ただしCLIは継続利用の本命ではなく、GitHub Checks UXへ進むための検証装置である。

6. Team intent matters more than generic best practice  
   「refactor」「bug fix」「copy update」「migration」「security hardening」の意味はチームごとに違う。将来的には組織ルールと過去PRからチーム固有の意図基準を学ぶ。

## 3. 解くべき問題

### 3.1 一次問題

PRの説明と実装差分の不一致により、レビュアーが誤った前提でレビューを始め、注意を置く場所を間違える。

ユーザーの実際の問いは「このPR説明は信用できるか」だけではない。より現場に近い問いは以下である。

- このPRは軽く見てよいのか、時間を取るべきか。
- 説明にない危ない変更はあるか。
- テスト、API、認証、DB、設定、infraのどこを見るべきか。
- PR説明を直してから通常レビューに進むべきか。

典型例:

- 「純粋なリファクタ」と書いているが、認証、例外処理、データ保存、APIレスポンスが変わっている。
- 「小さなUI変更」と書いているが、状態管理、API schema、DB query、feature flag が変わっている。
- 「テスト追加」と書いているが、実装コードも変わっている。
- 「バグ修正」と書いているが、既存の互換性を壊している。
- AIエージェントがPR説明を自動生成し、もっともらしいが不完全な説明を残している。

### 3.2 二次問題

- レビュアーはPRを読む順序を自分で再構築している。
- PR作成者は自分の変更範囲を過小申告しがちである。
- AI生成PRでは、説明文の流暢さが実装理解の正しさを保証しない。
- セキュリティ・QA・SRE観点のレビュー担当者は、どのPRを深く見るべきか判断しにくい。
- チームは「なぜこのPRが危なかったか」をレビュー後に学習しにくい。

### 3.3 やらないこと

v1.0までのIntent Diffは、以下を主目的にしない。

- 汎用バグ検出
- 脆弱性スキャン
- コードスタイル指摘
- 自動修正PRの作成
- PRのマージ可否を単独で決めるゲート
- すべてのファイルを深く読む大規模AIレビュー

## 4. 市場調査と外部環境

### 4.1 市場ドライバー

AI開発支援はすでに一般化している。Stack Overflow Developer Survey 2025 では、回答者の84%がAIツールを開発プロセスで利用中または利用予定、プロ開発者の51%が毎日利用している。一方で、AI出力の正確性を信頼する開発者より、信頼しない開発者の方が多いと報告されている。これは「AIで書く」だけでなく「AI出力をどう確認するか」の市場を作っている。

DORA 2024 は、AI採用が個人の生産性・フロー・仕事満足に利益をもたらす一方で、ソフトウェア delivery stability と throughput に負の影響もあり得ると整理している。つまり、AI導入の価値は「コード生成量」ではなく、レビュー、テスト、小さなバッチ、安定した優先順位などの運用能力に依存する。

Google Cloud / DORA のAI開発ROI資料は、AI導入の投資対効果を「人員削減」ではなく、不要な手戻りを減らしてエンジニアリングキャパシティを取り戻すことに結びつけている。Intent Diff はまさに、レビュー初期の誤読・説明修正・手戻りを減らす領域にある。

### 4.2 既存カテゴリ

#### AIコードレビュー

代表例:

- GitHub Copilot code review
- CodeRabbit
- Qodo
- Graphite AI Reviews
- Claude Code Review などのエージェント系レビュー

特徴:

- PRや未コミット変更をレビューし、コメントやsuggestionを出す。
- GitHub UI、IDE、PRコメント、GitHub Actionsなどに統合される。
- バグ、セキュリティ、品質、パフォーマンス、style、edge case を探す。
- カスタムルールや組織コンテキスト対応が進んでいる。

Intent Diff との差分:

- 既存AIレビューの主語は「コードの良し悪し」。
- Intent Diff の主語は「説明された意図と実装された変化の整合性」。
- 既存AIレビューはレビュアーの代替・補助になりやすい。Intent Diff はレビュアーがレビューを始める向きを整える。

#### 静的解析・セキュリティスキャン

代表例:

- Semgrep
- CodeQL
- Snyk
- SonarQube

特徴:

- ルールベースまたは解析ベースで脆弱性、品質、依存関係、secret、policy違反を検出する。
- 差分ベーススキャン、PRコメント、CI/CD統合が成熟している。

Intent Diff との差分:

- 静的解析は「既知ルールへの違反」に強い。
- Intent Diff は「このPRがそういう変更をするとは言っていない」に強い。
- 例えば、APIレスポンスの型変更がルール違反でなくても、「copy update」のPRでは重要なズレになる。

#### 構文diff・semantic diff

代表例:

- Difftastic
- AST diff / RefactoringMiner 系
- IDEやGitHubの差分ビュー

特徴:

- 行単位ではなく構文構造やリファクタリングを意識して差分を読みやすくする。
- フォーマット変更やwrapper追加などを視覚的に理解しやすい。

Intent Diff との差分:

- 構文diffは「何が変わったか」をより正確に示す。
- Intent Diff は「それがPRの意図と合っているか」を判断する。
- Phase 1以降、Intent Diff は構文diffをLLM前の構造シグナルとして使う余地がある。

#### PRサマリー・PR説明生成

代表例:

- GitHub Copilot PR summary
- CodeRabbit PR summary
- 各種AI agentのPR description生成

特徴:

- 変更内容を自然言語で要約する。
- レビュアーの導入を助けるが、PRの元意図と実装差分の照合までは主目的にしていない。

Intent Diff との差分:

- PRサマリーは「差分から説明を作る」。
- Intent Diff は「既にある説明・issue・promptと差分を照合する」。
- 説明を更新する提案は出すが、説明生成そのものが主役ではない。

#### 実務的な競合比較

| 既存手段 | ユーザーが得るもの | Intent Diffの差分 | 防衛線 |
| --- | --- | --- | --- |
| GitHub diff | 変更行、変更ファイル | PRの主張と変更範囲の整合性、読む順序 | risk taxonomy、attention map |
| Copilot PR summary | 差分から生成された変更要約 | 既に主張された意図との差分、説明不足の指摘 | intent taxonomy、protected claims |
| CodeRabbit / Qodo / Graphite等 | バグ・改善・セキュリティコメント | レビュー前の注意配分、PR説明修正、scope drift検出 | low-noise Checks UX、評価データセット |
| Semgrep / Snyk / CodeQL | 既知ルール違反、脆弱性、依存リスク | ルール化されていないスコープ逸脱や契約変更の説明漏れ | team rules、semantic claim mapping |
| Danger / GitHub Actions | 手続き的・構成的チェック | 意味的な不一致、曖昧なPR主張の検出 | LLM + structured taxonomy |
| Difftastic / AST diff | 構文上の変化を読みやすくする | 構文変化がPR意図と合うかを判断する材料化 | structured evidence pipeline |

競合が後から「PR説明とdiffのズレ」を機能追加する可能性は高い。Intent Diff の独立性は、単なるLLMプロンプトでは守れない。守るべき資産は、mismatch taxonomy、risk taxonomy、チームルール、評価データセット、GitHub Checks中心の低ノイズUX、self-host/privacy postureである。

### 4.3 市場ギャップ

混雑しているのは「AIがレビューコメントを書く」市場である。空いているのは「レビュー前にPRの意図を検査し、読むべき場所を決める」市場である。

Intent Diff が狙うギャップ:

- PR説明の信頼度を測るツールが少ない。
- 「refactor only」「no behavior change」「small copy update」などの危険な主張を構造的に検査する製品が少ない。
- AI-generated PRがもっともらしい説明を持つほど、レビュアーはPR説明ではなく実装証拠から注意配分を決める必要がある。
- AIエージェントの作業ログ・プロンプト・session summaryとPR diffを照合する層は未成熟。ただしこれはv0.1の芯ではなく、v1.0以降の拡張価値である。
- コードレビューの成果指標は「バグ検出」に偏り、「レビュアーの注意配分」や「説明修正による手戻り削減」は未開拓。

### 4.4 購買・継続利用の仮説

ユーザーは「意図と実装のズレ」という思想だけでは導入しない。導入判断は、次の実務的改善にかかる。

購買理由:

- AI-generated PRが増え、レビュー待ちとレビュー疲れが増えている。
- PR説明が流暢でも、実装範囲の説明漏れが起きる。
- 既存AIレビューはコメントが多く、レビュー前に読む順序を決める用途では重い。
- `refactor only`、`no behavior change`、`docs only`のような危険な主張を軽量に検査したい。

継続理由:

- レビュアーが最初に読むファイルを早く決められる。
- PR説明修正、追加確認、追加テスト依頼がレビュー前半で出る。
- 不要なPRコメントが少なく、Checks summaryだけで使える。
- チーム固有のrisk pathやprotected claimを設定できる。

### 4.5 研究からの示唆

AIコードレビューの有効性はまだばらつきが大きい。2025年のGitHub Actions上のAIコードレビュー研究では、22,000件以上のAIレビューコメントを分析し、簡潔でコードスニペットを含み、手動トリガーされ、hunk-levelのものほどコード変更につながりやすいと報告されている。これは、Intent Diff も「常時大量コメント」より「少数・明確・レビュアーが求めた時に強い」設計にすべきことを示唆する。

SWE-PRBench 2026 は、AIコードレビューが人間の専門レビュアーにまだ届かないこと、diff-only条件での人間指摘検出率が限定的であること、長いコンテキストが必ずしも改善しないことを示している。Intent Diff はこの知見を踏まえ、巨大コンテキスト投入ではなく、diff summary、構造シグナル、意図抽出を小さく鋭く行うべきである。

PRレビュー順序の研究では、レビュアーは必ずしもファイル名順に読まず、diffサイズ順、PRタイトル・説明との類似度順、test-firstなどの戦略を取る。Intent Diff は「レビュアーの読む順序」をプロダクト機能として扱える。

AI coding agent のPR説明研究では、PR説明の構造やスタイルがレビュアーの反応、応答時間、merge outcome と関連することが示されている。Intent Diff は、AI agentが生成したPR説明を、流暢さではなく整合性で評価する基盤になれる。

## 5. ターゲットユーザー

### 5.1 ICPの決定

v0.1からPhase 1の初期ICPは、AI coding agentを日常利用している小-中規模開発チームである。

この絞り込みは採用する。理由は、Intent Diff の「なぜ今必要か」が最も強く出るのが、AI-generated PRが増え、流暢なPR説明と実装差分のズレがレビュー負荷に直結するチームだからである。

優先順位:

1. AI-generated PRをレビューする小-中規模チームのTech Lead / senior reviewer
2. AI coding agentを使ってPRを作るauthor / agent operator
3. レビュアー負荷が高いOSS maintainer
4. SaaS企業のplatform / DevExチーム
5. security / compliance感度の高いチーム

人間PRにも使えることは重要だが、初期のメッセージでは主役にしない。人間PR対応は「同じ仕組みで自然に効く副次価値」として扱う。

### 5.2 Primary persona: AI-generated PRをレビューするTech Lead

状況:

- 1日に複数のAI-generated PRまたはAI-assisted PRをレビューする。
- PR説明は流暢だが、agentがどの範囲まで触ったかを毎回深く読む必要がある。
- 全部を深く読む時間はないが、見落としの責任は重い。

痛み:

- PR説明をどこまで信じてよいか分からない。
- どのPRを軽く見てよく、どのPRに時間を取るべきかを判断しにくい。
- 「小さい変更」と言われても、実際には認証・DB・APIが触られている。
- AIレビューbotのコメントが多く、かえって判断が遅れる。

成功:

- 30秒で「何を信じ、何を疑い、どこから読むべきか」が分かる。
- PR説明修正、追加テスト確認、API/認証/DB確認をレビュー前半で依頼できる。
- Checks summaryだけでレビュー開始時の地図が手に入る。

### 5.3 Secondary persona: PR author / AI agent operator

状況:

- 自分、またはAI agentがPRを作る。
- PR説明を自動生成または手早く書く。
- レビュー前に差分の説明漏れを直したい。

痛み:

- 変更内容を自分でも過小評価してしまう。
- 「refactor」と書いたが、実は挙動変更が入ったことに後で気づく。
- レビューで「PR説明が違う」と言われるのが遅い。

成功:

- PRを出す前に説明漏れを検出できる。
- 変更のスコープを正しく言語化できる。
- レビュアーの信頼を失わない。

### 5.4 Secondary persona: Engineering Manager / DevEx owner

状況:

- AI開発支援の導入によりPR量が増えている。
- レビュー待ち、手戻り、品質懸念を管理したい。

痛み:

- AI導入のROIを「生成量」ではなく「品質・手戻り・レビュー時間」で説明したい。
- レビュー疲れとbot疲れを避けたい。
- チーム規範に合ったレビュー補助を導入したい。

成功:

- チームで「意図ズレが多いPRの種類」を把握できる。
- PRテンプレートやagent promptを改善できる。
- レビューリードタイムと手戻りを測れる。

### 5.5 Secondary persona: Security / Compliance reviewer

状況:

- すべてのPRを深く見ることはできない。
- 触られた領域ではなく「言っていないのに変えた領域」を知りたい。

痛み:

- 「UI変更」のPRで権限・個人情報・ログ出力が変わる。
- 「依存更新」のPRでruntime behaviorが変わる。
- 規制領域では、説明と実装の不一致そのものが監査リスクになる。

成功:

- セキュリティレビューが必要なPRを早く拾える。
- 意図ズレの証跡を保存できる。

## 6. ユーザーのペインポイント詳細

### 6.1 レビュアーのペイン

- PR説明が抽象的で、何を保証しているのか曖昧。
- 「no behavior change」を信じて読むと、重要な変更を見落とす。
- AI生成の長いPR説明が、実装差分と微妙にズレる。
- 差分が大きいと、どこから読むべきか分からない。
- Botコメントが多すぎて、人間が本当に考えるべき点が埋もれる。
- 説明修正、追加テスト依頼、スコープ分割依頼がレビュー後半に出て手戻りになる。

### 6.2 PR authorのペイン

- PR descriptionを書くのが面倒。
- issueやpromptの意図を、実装後の差分と照らし直す習慣がない。
- 変更の副作用に気づかず、レビューで信頼を落とす。
- AI agentが実装した変更を、作成者自身が完全に説明できない。

### 6.3 チームのペイン

- PRテンプレートが形骸化する。
- 「refactor」と「behavior change」の境界が人により違う。
- レビューで同じ種類のズレが繰り返される。
- AIツール導入後、生成速度は上がったがレビュー品質の管理が追いつかない。
- コードレビューの改善が、個人の経験則に閉じる。

### 6.4 組織・監査のペイン

- PR説明が実装内容の監査証跡として弱い。
- AI agentが何を依頼され、何を変えたかを追跡しにくい。
- セキュリティ・プライバシー・データ保持・課金・権限の変更が小さなPRに紛れ込む。

## 7. コア概念

### 7.1 Claimed Intent

PRが主張している目的・範囲・非目的・保証。

入力候補:

- PR title
- PR body
- linked issue / Linear / Jira
- commit messages
- branch name
- PR template checkbox
- changelog / design doc reference

v1.0以降の拡張入力:

- original agent prompt
- agent session summary
- agent task transcript

出力例:

- Primary goal: 認証middlewareの内部構造を整理する
- Claimed scope: auth middlewareのみ
- Explicit non-goal: 挙動変更なし
- Implied guarantee: token expiry、logout、error codeは変わらない
- Confidence: medium

### 7.2 Implementation Evidence

diffから得られる実装上の証拠。これは「実際の振る舞い」を断定するものではない。テスト実行、型検査、runtime context、依存関係の解決なしに挙動は確定できないため、Intent Diff はImplementation EvidenceからBehavior-impact Hypothesesを作る。

入力候補:

- git diff
- file path
- symbol/function/class name
- tests
- API route / schema / migration / config / dependency file
- lockfile
- feature flag
- IaC / CI workflow
- README/docs/changelog

出力例:

- `src/auth/session.ts` のtoken expiry分岐が変更されている
- `src/auth/logout.ts` でsession clearの順序がaudit loggingより前に移動している
- 一部のfailure pathでHTTP statusが401から403に変わっている可能性がある
- auth middleware testsはhappy path中心で、expiry/error branchの検証が薄い

### 7.3 Behavior-impact Hypotheses

Implementation Evidenceから作るレビュー仮説。出力は「変わった」と断定せず、「変わった可能性がある」「確認すべき」と表現する。

例:

- Token expiryの扱いがPR説明にない形で変わっている可能性がある。
- Logout時のaudit logging順序が変わる可能性がある。
- API clientが401前提の場合、403変更の影響確認が必要。

### 7.4 Intent-Evidence Mismatch

Claimed Intent と Implementation Evidence / Behavior-impact Hypotheses の差。

Intent Diff の価値は、LLMがそれっぽい文章を生成することではなく、ズレの分類体系にある。v0.1では最低限、次のtaxonomyを固定する。

| 種別 | 内容 | 例 | 初期優先度 |
| --- | --- | --- | --- |
| Scope mismatch | 説明より変更範囲が広い | copy updateと言いつつstate管理変更 | high |
| Contract mismatch | API、DB、イベント、型、CLI、設定キーなど外部契約が変わる可能性 | response schema、HTTP status、migration変更 | high |
| Risk mismatch | 低リスクPRに見えるが高リスク領域を触る | auth、payment、permission、PII、logging変更 | high |
| Test mismatch | 主張とテスト内容が合わない | bug fixなのに回帰テストなし、test-onlyなのにsource変更 | medium |
| Intent under-specification | 説明が曖昧すぎてレビュー仮説を解消できない | "minor fix"、"cleanup"のみ | medium |
| Non-code impact mismatch | config、infra、ops影響が説明されていない | env var、migration、cron、CI、cache、queue変更 | high |
| Behavioral ambiguity | diffから見える影響が説明で解消されていない | fallback条件、error handling、retry条件変更 | medium |
| Documentation mismatch | docs/changelogと実装証拠が合わない | docsだけ更新、または実装変更にdocsなし | low |
| Dependency risk mismatch | dependency/lockfile/runtime変更の影響が説明されていない | major version bump、runtime image変更 | medium |

AI agentのpromptやsession summaryと最終diffの不一致は重要だが、v0.1では後回しにする。v1.0以降の `AI provenance mismatch` として扱う。

### 7.5 Alignment Score

単純な合否ではなく、レビュー前の信頼度として扱う。

推奨表現:

- A: Intent matches implementation evidence
- B: Minor omissions
- C: Material omissions, reviewer should inspect named areas
- D: Claimed scope is misleading
- E: High-risk mismatch; PR description or scope should be corrected before normal review

併記:

- Confidence: low / medium / high
- Evidence count
- Highest-risk category
- Suggested next action

注意:

- スコアだけをCI gateにしない。
- v0.1ではMarkdownサマリー中心。JSONには数値スコアを入れて自動評価可能にする。

## 8. 製品原則

1. Orientation before inspection  
   最初に読むべき場所を示す。レビューの代替を名乗らない。

2. Evidence over assertion  
   すべてのmismatchに、該当ファイル、diff hunk、関数名、入力文を添える。

3. Fewer comments, higher signal  
   PRに大量のinline commentを付けない。Phase 1のデフォルトはChecks summaryであり、PR commentはhigh-confidence mismatchのみopt-inにする。

4. Author-first when possible  
   PRを出す前にローカルで修正できる体験を重視する。

5. Cheap by default  
   大規模コンテキスト投入を避け、diff summary、path heuristics、構造シグナルでLLMコストを抑える。

6. Human-readable and machine-readable  
   MarkdownとJSONの両方を出す。

7. Local and private by design  
   ソースコード送信を制御できる。除外、redaction、self-host modelを前提にする。

8. Team-customizable without becoming brittle  
   intent categories、risk files、PR templates、ignore rules、model providerを設定可能にする。

## 9. 成功指標

North Star:

- Useful attention shift rate: Intent Diffの出力により、レビュアーの読む順序、確認観点、PR説明修正依頼、追加テスト確認のいずれかが実際に変わった割合。

補助的な原則:

- 検出数は成功指標にしない。
- false positiveは「間違い」だけでなく、レビュアーの注意を奪った時間として測る。
- "No significant mismatch" が静かで信頼できることも価値として扱う。

### 9.1 ユーザー価値指標

- レビュアーが「最初に読むべきファイル」を判断する時間が短くなる。
- PR説明修正コメントがレビュー後半ではなくレビュー前半で出る。
- 「refactor/no behavior change」PRで、説明にない実装影響の確認が早くなる。
- Intent Diffのsummaryが、レビュアーの実際のレビュー順序に影響する。
- レビュアーが「この出力は注意を奪う価値があった」と判断する割合が高い。

### 9.2 プロダクト指標

v0.1:

- CLI実行成功率
- 1PRあたり実行時間
- 1PRあたり推定コスト
- JSON schema安定性
- 手動評価でのuseful attention shift率
- 不要・うるさいと判断された出力率
- PR description修正につながった件数
- 追加確認・追加テストにつながった件数
- 1PRあたりの出力文字量

Phase 1:

- GitHub Action導入リポジトリ数
- PRあたりコメント数
- Checks summary閲覧・リアクション
- AuthorがPR説明を修正した割合
- Re-run率
- False positive feedback率

v1.0:

- 週次アクティブリポジトリ
- チーム内での平均review lead time変化
- 意図ズレカテゴリ別トレンド
- Enterprise/self-host導入数
- 重大mismatchの人間確認precision

### 9.3 ガードレール指標

- PRコメントが多すぎないこと
- 「誤った強い断定」の割合
- high-risk mismatchのprecision
- false positive cost: レビュアーが不要に確認した時間
- Checks summaryで完結した割合
- 機密情報の出力漏れ
- 大規模PRでのtimeout率
- モデルコストの上限超過率

## 10. 機能要求の全体像

### 10.1 入力取得

Must:

- ローカルgit diffを読む
- base/head指定
- PR description Markdownを読む
- stdin入力

Should:

- GitHub PR URLからtitle/body/diffを取得
- linked issue本文を取得
- commit messagesを読む
- branch nameを使う
- `.intent-diff.yml`を読む

Could:

- Linear/Jira/Notion連携
- Claude/Codex/Cursor agent log importer
- PR template checkboxの構造解析
- OpenAPI/GraphQL schema比較

### 10.2 意図抽出

Must:

- Primary goal
- Claimed scope
- Explicit non-goals
- Implied guarantees
- Risk-sensitive keywords
- Ambiguity warnings

Should:

- PR説明の不足箇所を指摘
- issueとPR descriptionのズレを抽出
- commit messagesとのズレを抽出
- intent confidenceを出す

Could:

- チームの過去PRからintent taxonomyを学習
- PR templateごとの抽出ルール

### 10.3 差分理解

Must:

- changed files
- additions/deletions
- language/file type
- tests/docs/source/config分類
- high-risk path分類
- diff hunk summary

Should:

- function/class/symbol名抽出
- public APIらしき変更検出
- dependency/lockfile変更検出
- migrations/env/config/CI変更検出
- test coverage relationの粗い推定

Could:

- Difftastic/tree-sitterによる構文diff
- language-specific analyzers
- OpenAPI/GraphQL/DB schema diff
- call graph / import graph
- package managerごとのtransitive risk annotation

### 10.4 照合

Must:

- claimed scope vs changed files
- non-goal vs implementation evidence
- primary goal vs unrelated changes
- risk-sensitive changes vs PR説明
- mismatch category
- evidence

Should:

- severity
- confidence
- reviewer focus areas
- suggested PR description correction
- suggested split recommendation
- suggested tests to inspect/add

Could:

- org policyとの照合
- CODEOWNERSとの照合
- reviewer assignment recommendation
- historical false positive learning

### 10.5 出力

Must:

- Markdown report
- JSON report
- exit code policy: always 0 by default, threshold指定時のみnon-zero
- concise summary

Should:

- GitHub Checks summary
- high-confidence mismatchのみPR comment
- SARIF-like structured outputは検討
- local HTML report

Could:

- VS Code extension panel
- Slack/Linear summary
- Dashboard
- API

### 10.6 設定

Must:

- model provider
- include/exclude files
- max diff size
- output format
- risk path patterns

Should:

- custom intent categories
- protected claims: `no behavior change`, `refactor only`, `docs only`
- output mode: local / check_summary / comment_on_high_confidence
- language
- redaction rules

Could:

- org-level policy packs
- per-directory intent rules
- team-specific prompt templates
- historical learning toggle

## 11. MVP v0.1

### 11.1 目的

「PR説明とgit diffだけで、レビュー前の注意配分に役立つ出力が作れるか」を最短で検証する。

v0.1のCLIはプロダクト本体ではなく、評価データ収集と出力品質検証のための実験装置である。継続利用の本命はPhase 1のGitHub Action / Checks summaryである。

v0.1は、GitHub Appや高度な静的解析よりも、以下を優先する。

- 1コマンドで動く
- 出力がレビュアーにとって即読める
- JSONで後から評価できる
- 意図ズレの分類が妥当
- コストとレイテンシが小さい

v0.1で検証しないこと:

- agent prompt / session summary があれば精度が上がるか
- 組織履歴やチーム学習が効くか
- PRコメントで継続利用されるか

### 11.2 想定UX

ローカルPR作成前:

```bash
intent-diff analyze \
  --base main \
  --intent pr.md \
  --out intent-diff.md \
  --json intent-diff.json
```

PRレビュー前:

```bash
gh pr view 123 --json title,body > pr.json
intent-diff analyze --base origin/main --pr-json pr.json
```

標準出力:

```text
Alignment: C - Material omissions
Highest risk: Risk mismatch

Claimed intent:
- Refactor auth middleware without behavior changes.

Implementation evidence:
- Token expiry branch changed in src/auth/session.ts.
- Logout appears to clear session before audit logging.
- One failure path may return 403 instead of 401.

Behavior-impact hypotheses:
- This may not be a pure refactor.
- Auth edge-case behavior may have changed.

Attention map:
1. Review auth edge cases around token expiry.
2. Confirm 401/403 behavior is intended.
3. Ask author to update PR description if behavior change is intentional.
```

### 11.3 MVP機能

Must have:

- CLI command: `intent-diff analyze`
- Inputs:
  - `--base`
  - `--head`
  - `--diff-file`
  - `--intent`
  - stdin
- git diff収集
- PR/intent Markdown parsing
- simple file classification
- LLM prompt for:
  - claimed intent
  - implementation evidence
  - behavior-impact hypotheses
  - mismatch list
  - attention map
  - PR description correction
- Markdown output
- JSON output
- basic config file: `.intent-diff.yml`
- max diff size handling
- ignore patterns
- deterministic-ish prompt structure
- no network unless model provider is configured

Should have:

- `--pr-json`
- commit message collection
- linked issue text input
- `--no-llm` heuristic mode
- OpenAI-compatible API endpoint設定
- redaction patterns
- Japanese/English output language
- golden sample tests
- fixture PRs for mismatch categories
- README quickstart

Non-goals:

- GitHub comments
- GitHub App
- agent prompt/session summary importer
- dashboard
- multi-repo context
- AST parsing必須化
- 自動修正
- 自動merge gate

### 11.4 MVP出力仕様

Markdown sections:

1. Summary
2. Claimed Intent
3. Implementation Evidence
4. Behavior-impact Hypotheses
5. Potential Mismatches
6. Attention Map
7. Suggested PR Description Correction
8. Evidence
9. Confidence and Limitations

JSON schema概略:

```json
{
  "version": "0.1",
  "alignment": {
    "grade": "C",
    "score": 0.62,
    "confidence": "medium",
    "highest_risk_category": "risk_mismatch"
  },
  "claimed_intent": {
    "primary_goal": "...",
    "claimed_scope": ["..."],
    "non_goals": ["..."],
    "implied_guarantees": ["..."]
  },
  "implementation_evidence": [
    {
      "summary": "...",
      "files": ["src/auth/session.ts"],
      "change_type": "control_flow",
      "risk_tags": ["auth", "status_code"]
    }
  ],
  "behavior_impact_hypotheses": [
    {
      "summary": "...",
      "confidence": "medium",
      "requires_verification": true
    }
  ],
  "mismatches": [
    {
      "category": "risk_mismatch",
      "severity": "high",
      "confidence": "medium",
      "claim": "...",
      "observation": "...",
      "evidence": [
        {
          "file": "src/auth/session.ts",
          "lines": "approximate or hunk id",
          "note": "..."
        }
      ],
      "recommended_action": "..."
    }
  ],
  "attention_map": ["..."],
  "suggested_pr_description": "..."
}
```

### 11.5 MVP受け入れ条件

- READMEのexample相当のauth refactorケースで、hidden behavior changeを検出できる。
- docs-only PRでsource code変更が混入した場合に検出できる。
- test-only PRでproduction code変更が混入した場合に検出できる。
- dependency/lockfile変更を「説明が必要な変更」として扱える。
- 1000行程度のdiffで実用的な時間内に完了する。
- Markdownだけ読んでもレビューの初手が分かる。
- JSONを使ってfixtureテストが書ける。
- 挙動を断定せず、verification-neededな仮説として表現できる。

### 11.6 MVP成功判定

少数の実PRまたは過去PR 30-50件で手動評価する。

評価項目:

- Useful: レビュアーの読む順序や質問が変わる
- Accurate: 指摘されたmismatchが妥当
- Not noisy: 重要でない指摘が少ない
- Actionable: 次に何をすべきか明確
- Attention-worthy: レビュアーの注意を奪う価値があった
- Cheap: 1PRあたりコストが継続利用可能

目標:

- useful attention shift率 60%以上
- high severity mismatch precision 70%以上
- noisy output率 20%以下
- PR description修正または追加確認につながる率 30%以上
- 1PRあたりsummary出力 1件以内
- false alarmが多いカテゴリを特定できる

## 12. Phase 1

### 12.1 目的

ローカル便利ツールから、チームのPRワークフローに自然に入るツールへ移行する。

Phase 1のテーマ:

- GitHub PR上で使える
- チームごとのルールを設定できる
- コメントがうるさくない
- 評価と改善サイクルが回る
- LLMだけに頼らない構造シグナルを増やす

### 12.2 機能

#### GitHub Action

Must:

- `pull_request` trigger
- GitHub Checks summary出力
- PR commentはデフォルト無効
- `workflow_dispatch`
- rerun support
- permissions最小化
- private repo対応

Should:

- changed filesのみ処理
- large diff fallback
- previous runとの差分
- high-confidence mismatchのみbot comment update
- labelsによるmanual trigger
- reviewdog/SARIF風連携は検討
- inline commentは原則しない

#### 設定ファイル拡張

`.intent-diff.yml`:

```yaml
version: 1
language: ja
intent_sources:
  - pr_title
  - pr_body
  - commits
  - linked_issues
risk_paths:
  auth:
    - "src/auth/**"
    - "app/**/middleware*"
  api_contract:
    - "openapi/**"
    - "src/routes/**"
  data:
    - "migrations/**"
    - "schema/**"
protected_claims:
  - "no behavior change"
  - "refactor only"
  - "docs only"
output_mode: check_summary
comment_policy:
  enabled: false
  only_high_confidence: true
thresholds:
  fail_on_grade: null
redaction:
  patterns:
    - "sk-[A-Za-z0-9]+"
```

#### 構造シグナル

Should:

- tree-sitterまたはDifftastic連携の検証
- language別symbol extraction
- API route変更検出
- DB migration検出
- dependency file分類
- tests/source対応関係の粗い推定
- CODEOWNERS参照

#### 出力改善

Must:

- 「1画面で分かる」summary
- evidence links to files
- severity/confidence separation
- suggested PR body patch
- attention map: read first / verify / ask author

Should:

- "Reviewer reading order"
- "Questions to ask author"
- "Tests to inspect"
- "No mismatch found"時の静かな出力

#### 評価基盤

Must:

- fixture corpus
- expected JSON snapshots
- mismatch taxonomy tests
- prompt regression tests

Should:

- human label UIまたはCSV
- feedback command:

```bash
intent-diff feedback --run-id ... --useful yes --false-positive risk_mismatch
```

### 12.3 Phase 1 非機能要件

- PR 1件あたりの標準実行時間: 60秒以内を目標
- 大規模PRではpartial analysisを明示
- 失敗時はレビューをブロックしない
- 機密ファイルは除外可能
- 結果は再現可能な入力サマリーを保持
- モデル呼び出しログに生コードを残さない設定

### 12.4 Phase 1 成功判定

- 3-5の実リポジトリで継続実行できる。
- PRコメント過多にならない。
- authorがPR説明を更新する事例が出る。
- tech leadが「最初に見る場所の判断」に使う。
- high-risk protected claimでの有用指摘が複数出る。
- Checks summaryだけで価値を感じる利用者が出る。

## 13. Phase 2 / v1.0 リリースレベル

### 13.1 目的

組織が安心して導入できる、レビュー信頼性レイヤーにする。

v1.0のテーマ:

- GitHub Appとして導入が簡単
- チーム/組織ルールに対応
- 評価可能で改善可能
- privacy/security postureが明確
- OSS/Pro/Enterpriseの製品境界がある

### 13.2 v1.0機能

#### GitHub App

Must:

- repo install
- PR Checks integration
- PR comment update
- per-repo config discovery
- org-level defaults
- installation permissions最小化
- UIでのrun history

Should:

- PR label/manual trigger
- review request integration
- CODEOWNERS-aware reviewer focus
- branch protectionとは独立したsoft gate
- PR commentはhigh-confidenceまたはpolicy violationのみ
- GitHub Enterprise Server検討

#### チームルール / Policy Packs

Must:

- protected claims
- risk path taxonomy
- mismatch taxonomy
- required explanation rules
- file type category rules
- output language

Should:

- org templates
- repo inheritance
- rule dry-run
- false positive suppression
- "refactor policy", "API policy", "migration policy", "security policy"

#### 履歴と学習

Must:

- run history
- mismatch category trend
- PR description correction tracking
- feedback collection

Should:

- 過去PRからチーム固有のintent vocabularyを抽出
- author別ではなくPR pattern別の改善提案
- PR template改善提案
- agent prompt改善提案
- noisy rule suppressionの学習

#### エンタープライズ・セキュリティ

Must:

- data retention設定
- self-host / private model provider
- audit logs
- SSO/SAMLはEnterpriseで検討
- source code storage policy明示
- secrets redaction

Should:

- VPC/self-host runner対応
- no-code-retention model provider guide
- compliance export
- on-prem GitHub Enterprise対応

#### API / Integrations

Should:

- REST API for analysis
- webhook output
- Linear/Jira issue source
- Slack summary
- Codex/Claude/Cursor session summary importer
- GitLab supportはv1.0後半またはv1.1候補

#### Dashboard

Should:

- repo overview
- mismatch categories
- protected claim failure trends
- review lead time correlation
- cost/usage
- top noisy rules

DashboardはMVPでは不要。ただしv1.0ではDevEx ownerとEM向けの価値になる。

### 13.3 v1.0の品質基準

- high-risk mismatch precisionが実運用で十分高い
- low/medium severityを大量に出さない
- 大規模PRでもpartial summaryとして破綻しない
- 組織が設定・監査・費用管理できる
- 主要な失敗モードが明示される
- "No significant mismatch"が信用できる
- Checks summary中心で、PR会話欄を汚さずに価値が出る

### 13.4 v1.0 非目標

- 完全自動レビュー代替
- 自動マージ判断
- 汎用SAST競合
- 全言語で深い意味解析
- すべてのagent toolログの完全対応

## 14. 製品仕様詳細

### 14.1 Analysis pipeline

1. Collect inputs
   - PR metadata
   - intent text
   - commits
   - diff
   - optional linked task

2. Normalize
   - file classification
   - diff truncation
   - redaction
   - hunk summaries
   - path risk tags

3. Extract intent
   - goal
   - scope
   - non-goals
   - guarantees
   - ambiguity

4. Extract implementation evidence
   - behavior-impact hypotheses
   - contracts
   - tests/docs/config/infra
   - risk tags

5. Compare
   - map claims to implementation evidence
   - detect unsupported claims
   - detect unclaimed changes
   - assign category/severity/confidence

6. Render
   - human summary
   - evidence
   - reviewer focus
   - suggested PR description correction
   - JSON

7. Feedback
   - useful/not useful
   - false positive category
   - ignored/no action

### 14.2 Severity model

Severityは「ズレの危険度」であり、「バグ確率」ではない。

High:

- explicit non-goalに反する
- auth/security/privacy/data/API/infraを説明なく変更
- public contract drift
- migration/deployment impact drift

Medium:

- claimed scope外のsource変更
- test/source不一致
- dependency/lockfile変更の説明不足
- docs/changelog不足

Low:

- PR説明の曖昧さ
- minor omitted files
- naming/copy程度の説明漏れ

### 14.3 Confidence model

Confidenceは、モデルの断定度ではなく、証拠の強さ。

High:

- explicit claimあり
- direct diff evidenceあり
- path risk tagあり
- testsまたはpublic contractに関連

Medium:

- claimは暗黙
- diff evidenceはあるが挙動推測が必要

Low:

- diff context不足
- truncated
- generated/minified/vendor
- language unsupported

### 14.4 コメント方針

デフォルト:

- GitHub Checks summaryのみ
- PR commentはhigh-confidence mismatchまたは明示policy違反のみopt-in
- inline commentはv1.0までは原則しない
- "No mismatch found"の場合はChecks summaryのみ

コメント例:

```markdown
## Intent Diff: Alignment C

This PR claims a refactor with no behavior change, but the diff contains auth-related implementation evidence that should be verified.

Potential mismatches:
- `src/auth/session.ts`: token expiry branch changed.
- `src/auth/logout.ts`: logout appears to clear session before audit logging.

Attention map:
1. Confirm 401/403 behavior.
2. Review token expiry edge cases.
3. Ask author to update the PR description if behavior change is intended.
```

## 15. 技術アーキテクチャ案

### 15.1 コンポーネント

- CLI
  - command parsing
  - local git integration
  - output rendering

- Core analyzer
  - input normalization
  - file classification
  - risk tagging
  - intent extraction
  - implementation evidence extraction
  - behavior-impact hypothesis generation
  - comparison

- Model adapter
  - OpenAI-compatible
  - local/self-host endpoint
  - mock provider for tests

- Integrations
  - GitHub Action
  - GitHub App
  - issue trackers
  - agent logs

- Evaluation
  - fixtures
  - golden outputs
  - feedback store

### 15.2 言語・実装候補

**確定: Go 1.26** (ADR-0001)。

選定理由: GitHub Action/CLI配布が容易、goreleaser によるクロスプラットフォームバイナリ配布、cobra/go-gitdiff エコシステム。Rust は serde_yaml deprecated (RUSTSEC) + git2-rs の C deps が障壁。Python は配布摩擦。

却下候補:

- Rust: serde_yaml deprecated (RUSTSEC-2024-0005)、git2-rs の C deps がビルド複雑化。
- TypeScript/Node: 単一バイナリ配布に工夫が必要。
- Python: CLI製品配布と速度で不利。

### 15.3 LLM依存を下げる設計

LLMに直接巨大diffを渡さない。

前処理:

- file list
- diff stats
- risk path tags
- hunk summaries
- protected claim extraction
- file type classification
- dependency/config/schema flags

LLMの役割:

- 自然言語意図の抽出
- diff summaryの意味づけ
- claim-observation比較
- reviewer向け文章化

LLM以外の役割:

- ファイル分類
- risk tag
- schema検出
- protected claim detection
- output validation
- JSON schema validation

### 15.4 大規模PR対応

方針:

- Large PRは完全解析を装わない。
- まずfile-level risk triageを行う。
- high-risk filesを優先してLLMに送る。
- skipped/truncated evidenceを明示する。

例:

```text
Partial analysis: 42 of 219 changed files analyzed.
Skipped generated/vendor files: 130.
High-risk files analyzed: auth, migration, API schema.
Confidence reduced to medium due to truncation.
```

## 16. GTM / 市場投入戦略

### 16.1 Beachhead

最初に狙うべきユーザー:

- AI coding agentを日常利用している小-中規模開発チーム
- その中でレビュー責任を持つTech Lead / senior reviewer
- レビュアー負荷が高いOSS maintainer
- Platform/DevExチーム
- レビュー負荷が高いスタートアップ
- regulated SaaSのアプリケーションチーム

避けるべき初期市場:

- AIレビューを完全自動化したいだけの層
- 既存SASTの置き換えを期待する層
- 大企業の全社導入を最初から狙う案件

### 16.2 OSS戦略

OSSで広げるべき範囲:

- CLI
- GitHub Action
- 基本taxonomy
- JSON schema
- provider adapter
- local config

商用化余地:

- GitHub App hosted service
- org policy packs
- dashboard
- run history
- feedback analytics
- enterprise privacy controls
- support
- self-host package

### 16.3 メッセージング

避ける表現:

- "AI code reviewer"
- "Finds all bugs"
- "Replaces human review"
- "Semantic diff"

推奨表現:

- "Review attention map for AI-generated pull requests"
- "Decide what to trust, what to verify, and where to read first"
- "Review orientation for AI-era pull requests"
- "Compare what a PR claims with implementation evidence from the diff"
- "Catch scope drift before review starts"
- "A trust layer for agent-generated PRs"

Enterprise向けでは使えるが、v0.1の前面には出さない表現:

- "Make no-behavior-change claims auditable"
- "PR explanation accountability"

### 16.4 価格仮説

OSS:

- local CLI
- GitHub Action self-managed

Pro:

- hosted GitHub App
- private repos
- monthly PR quota
- basic history

Team:

- org rules
- dashboard
- feedback analytics
- shared policy packs
- priority model routing

Enterprise:

- self-host
- SSO
- audit logs
- data retention
- custom model provider
- GitHub Enterprise Server
- support/SLA

価格はAIレビューbotより安く見せるべきである。Intent Diff は深い全量レビューではなく、レビュー前のattention allocationであり、PRごとの低コスト実行が価値の前提になる。

## 17. リスクと対策

### 17.1 技術リスク

LLM hallucination:

- 対策: evidence必須、confidence分離、JSON schema validation、low confidence明示。

Behavior overclaiming:

- 対策: "observed behavior"を使わず、Implementation EvidenceとBehavior-impact Hypothesesに分離する。出力は「確認すべき可能性」として書く。

Large diff failure:

- 対策: partial analysis、risk-first sampling、truncation明示。

False positives:

- 対策: feedback、suppression、protected claim中心、severityを控えめにする。

False negatives:

- 対策: high-risk path heuristics、評価コーパス、カテゴリ別recall測定。

Model cost:

- 対策: pre-summarization、cache、small model option、manual trigger。

### 17.2 UXリスク

Bot fatigue:

- 対策: Checks summaryをデフォルトにし、No mismatchは静かに、inline commentは原則しない。

Score misuse:

- 対策: CI gate非推奨、soft signal、confidence併記。

Author defensiveness:

- 対策: blameしない文体、"if intentional, update PR description"を基本にする。

### 17.3 市場リスク

既存AIレビューに吸収される:

- 対策: mismatch taxonomy、risk taxonomy、protected claims、低ノイズChecks UX、評価データセット、team rules、PR description correctionに集中。

GitHub/Copilotが同機能を追加する:

- 対策: OSS trust、local-first、multi-provider、team-specific policy、self-host/privacy postureで差別化。

価値がニッチすぎる:

- 対策: AI agent PR、refactor claims、regulated teams、review bottleneckで強いuse caseを検証する。

単なるLLM wrapperになる:

- 対策: v0.1からtaxonomy、JSON schema、fixture corpus、feedback loopを製品コアとして扱う。プロンプト文面ではなく、分類・評価・低ノイズUXを資産化する。

### 17.4 セキュリティ・プライバシーリスク

ソースコード外部送信:

- 対策: 明示設定、redaction、self-host provider、no-store mode。

機密情報のreport出力:

- 対策: secret scanning/redaction、evidence snippet制御。

監査ログの過剰保持:

- 対策: retention policy、local-only mode。

## 18. 検証計画

### 18.1 Discovery questions

- AI-generated PRをレビューするチームで、PRの意図ズレをどの頻度で経験するか。
- どの主張が最も危険か: refactor only / docs only / no behavior change / tests only / dependency update。
- PR説明修正提案はauthorに受け入れられるか。
- レビュアーはAlignment Scoreを好むか、それともカテゴリとattention mapだけでよいか。
- Checks summaryだけで継続利用に足るか。PRコメントはどの閾値なら許容されるか。
- GitHub Actionは自動実行か手動トリガーか。
- どの出力が「注意を奪う価値があった」と判断されるか。

### 18.2 手動評価コーパス

初期は以下のPRを集める。

- refactor claimed, behavior changed
- docs/copy claimed, code changed
- test claimed, source changed
- bug fix claimed, API contract changed
- dependency update with behavior risk
- migration/config/CI changes
- accurate PR description
- intentionally broad PR
- AI-generated PR description

各PRに人間ラベル:

- claimed intent
- implementation evidence
- behavior-impact hypothesis
- mismatch categories
- useful attention map
- acceptable output
- noisy output
- PR description correction happened / did not happen

### 18.3 ユーザーインタビュー

対象:

- 5人のシニアレビュアー
- 3人のAI agent heavy user
- 2人のEM/DevEx owner
- 1-2人のsecurity reviewer

聞くこと:

- 最近の「説明と実装が違ったPR」
- レビュー前に知りたい情報
- AIレビューbotへの不満
- どのコメントなら邪魔ではないか
- PR説明修正をいつ行いたいか

### 18.4 ベータ導入

ステップ:

1. ローカルCLIを作者自身のPRで使う。
2. 2-3 OSS repoで過去PRに対してオフライン評価。
3. GitHub Actionをmanual triggerで導入。
4. Checks summaryのみで運用。
5. Summary commentをopt-in。
6. feedbackを収集。

v0.1での最小評価レポート:

- analyzed PR count
- useful attention shift rate
- noisy output rate
- high-severity precision
- PR description correction count
- added verification/test request count
- median output length
- median runtime/cost

## 19. ロードマップ詳細

### v0.1 MVP

期間目安: 2-4週間

Deliverables:

- CLI experiment harness
- basic config
- Markdown/JSON output
- OpenAI-compatible model adapter
- fixture tests
- README quickstart
- sample reports
- v0.1 evaluation report template

Milestone:

- 自分の実PRに対して毎回使える。
- READMEのexampleを再現できる。
- 10個以上のfixtureで分類が崩れない。
- PR説明+diffだけでuseful attention shiftが起きるか判断できる。

### Phase 1

期間目安: 1-2か月

Deliverables:

- GitHub Action
- Checks summary default UX
- high-confidence only optional PR comment
- `.intent-diff.yml`拡張
- risk path taxonomy
- mismatch taxonomy
- protected claims
- evaluation corpus
- feedback command
- model/cost controls

Milestone:

- 複数repoで毎週使われる。
- PRコメント過多にならない。
- authorがPR説明修正に使う。
- reviewerがChecks summaryをレビュー開始時に見る。

### Phase 2 / v1.0

期間目安: 3-6か月

Deliverables:

- GitHub App
- hosted service
- org-level policy
- run history
- dashboard
- enterprise privacy controls
- self-host/private provider path
- integrations for issue/task/agent logs
- stable JSON API

Milestone:

- チーム導入でレビュー初期の手戻り削減が見える。
- high-risk mismatchのprecisionが十分高い。
- OSS CLIからhosted productへの導線がある。

## 20. オープンクエスチョン

プロダクト:

- Alignment ScoreはA-Eがよいか、risk labelだけがよいか。
- Checks summaryだけで十分か、high-confidence PR commentをどの閾値で許可すべきか。
- PR author向けpreflightとreviewer向けorientationのどちらをPhase 1の主導線にするか。
- Attention mapの最小表現は「read first / verify / ask author」の3分類で足りるか。

技術:

- v0.1の実装言語はRust/Go/Pythonのどれがよいか。
- Difftastic/tree-sitterをいつ導入するか。
- JSON schemaは独自でよいか、SARIF/CodeClimate互換を検討すべきか。
- agent session summaryの標準形式をv1.0で定義するか、それとも各agent importerに閉じるか。

市場:

- AI code review botとの差別化を一言でどう伝えるか。
- OSSと商用の境界をどこに置くか。
- GitHub以外、特にGitLab対応はどの時点で必要か。

## 21. 参考情報

調査日: 2026-06-20

- Intent Diff repository: <https://github.com/yottayoshida/intent-diff>
- Stack Overflow Developer Survey 2025, AI section: <https://survey.stackoverflow.co/2025/ai>
- DORA 2024 Accelerate State of DevOps Report overview: <https://dora.dev/research/2024/dora-report/>
- DORA ROI of AI-assisted Software Development page: <https://cloud.google.com/resources/content/dora-roi-of-ai-assisted-software-development>
- GitHub Copilot code review docs: <https://docs.github.com/en/copilot/how-tos/use-copilot-agents/request-a-code-review/use-code-review>
- CodeRabbit homepage: <https://www.coderabbit.ai/>
- Qodo code review docs: <https://docs.qodo.ai/qodo-documentation/code-review>
- Graphite AI Reviews: <https://graphite.com/features/ai-reviews>
- Semgrep CI docs: <https://docs.semgrep.dev/semgrep-ci/sample-ci-configs>
- Difftastic: <https://difftastic.wilfred.me.uk/>
- Does AI Code Review Lead to Code Changes? A Case Study of GitHub Actions: <https://arxiv.org/abs/2508.18771>
- SWE-PRBench: Benchmarking AI Code Review Quality Against Pull Request Feedback: <https://arxiv.org/abs/2603.26130>
- How AI Coding Agents Communicate: A Study of Pull Request Description Characteristics and Human Review Responses: <https://arxiv.org/abs/2602.17084>
- Not One to Rule Them All: Mining Meaningful Code Review Orders From GitHub: <https://arxiv.org/abs/2506.10654>
- Measuring the Impact of Early-2025 AI on Experienced Open-Source Developer Productivity: <https://arxiv.org/abs/2507.09089>
