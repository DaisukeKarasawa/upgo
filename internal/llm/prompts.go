package llm

// Prompt templates for LLM analysis operations.
const (
	// PromptPRDescriptionSummary is a template for summarizing PR descriptions in Japanese.
	PromptPRDescriptionSummary = `以下のPR説明を日本語で要約してください。
技術的な内容を正確に保ちながら、簡潔にまとめてください。

%s`

	// PromptDiffSummary is a template for analyzing and summarizing code diffs in Japanese.
	PromptDiffSummary = `以下のコード変更差分を分析し、日本語で以下の形式で出力してください：

【要約】
変更内容の要約を簡潔に記述してください。

【解説】
各変更の意図と理由、技術的な影響について詳しく解説してください。

%s`

	// PromptCommentsSummary is a template for analyzing PR comments and discussions in Japanese.
	PromptCommentsSummary = `以下のPRコメントと議論を分析し、日本語で以下の形式で出力してください：

【コメント要約】
主な議論のポイントと重要な指摘を要約してください。

【議論要約】
合意形成のプロセスと議論の流れを要約してください。

%s`

	// PromptMergeReason is a template for analyzing why a PR was merged.
	PromptMergeReason = `以下のPRがマージされた理由を分析してください。
コメント、議論、変更内容から以下を抽出してください：
1. マージされた理由・背景
2. 技術的な判断軸
3. 意思決定の意図

%s`

	// PromptCloseReason is a template for analyzing why a PR was closed.
	PromptCloseReason = `以下のPRがクローズされた理由を分析してください。
コメント、議論から以下を抽出してください：
1. クローズされた理由・背景
2. 判断の軸
3. 今後の方針（あれば）

%s`

	// PromptMentalModel is a template for analyzing committer/reviewer mental models from PR/Issue comments and discussions.
	PromptMentalModel = `以下のマージされたPR/Issueのコメントと議論を分析し、
コミッター（レビュアー）のメンタルモデルを抽出してください。

分析観点：
- コードスタイルやベストプラクティスの傾向
- レビューで重視するポイント（パフォーマンス、セキュリティ、可読性など）
- 意思決定パターン（どのような議論の流れで決定されるか）
- 技術的哲学や設計思想

%s`
)
