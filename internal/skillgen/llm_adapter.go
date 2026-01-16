package skillgen

import (
	"context"
	"fmt"
	"strings"

	"upgo/internal/llm"
)

// LLMAdapter adapts the llm.Client to the LLMClient interface
type LLMAdapter struct {
	client *llm.Client
}

// NewLLMAdapter creates a new LLM adapter
func NewLLMAdapter(client *llm.Client) *LLMAdapter {
	return &LLMAdapter{client: client}
}

// GenerateSkillContent generates skill content using the LLM
func (a *LLMAdapter) GenerateSkillContent(ctx context.Context, category string, prData []PRData) (string, error) {
	prompt := buildSkillPrompt(category, prData)

	response, err := a.client.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	return response, nil
}

func buildSkillPrompt(category string, prData []PRData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`あなたはClaude Code用のスキルファイルを生成するアシスタントです。
以下のPR分析データを基に、カテゴリ「%s」のSKILL.mdファイルを生成してください。

## フォーマット

YAML フロントマターとMarkdownで出力してください：

---
name: %s
description: [スキルの説明（日本語）。このスキルがいつ使用されるべきかを含めてください]
---

# [タイトル]

## 概要
[このスキルの概要]

## パターン
[PR分析から抽出されたパターンとベストプラクティス]

## 例
[具体的なコード例やパターンの使用例]

## 注意点
[よくある間違いや注意すべき点]

---

## PR分析データ

`, category, category))

	for i, pr := range prData {
		if i >= 10 { // Limit to 10 PRs
			break
		}
		sb.WriteString(fmt.Sprintf("### PR #%d: %s\n", pr.ID, pr.Title))
		if pr.Summary != "" {
			sb.WriteString(fmt.Sprintf("要約: %s\n", pr.Summary))
		}
		if pr.Analysis != "" {
			sb.WriteString(fmt.Sprintf("分析: %s\n", pr.Analysis))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n上記のデータを基に、実用的で詳細なSKILL.mdファイルを生成してください。")

	return sb.String()
}
