package skillgen

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"upgo/internal/llm"

	"go.uber.org/zap"
)

// Generator generates Claude Code skills from PR analysis
type Generator struct {
	db        *sql.DB
	llm       *llm.Client
	logger    *zap.Logger
	outputDir string
}

// PRAnalysisData represents PR analysis data
type PRAnalysisData struct {
	PRID      int
	PRNumber  int
	Title     string
	State     string
	Author    string
	Analysis  string
	UpdatedAt time.Time
}

// ParsedAnalysis represents parsed analysis JSON
type ParsedAnalysis struct {
	Category     string   `json:"category"`
	Summary      string   `json:"summary"`
	Discussion   string   `json:"discussion"`
	Insights     []string `json:"insights"`
	GoPhilosophy string   `json:"go_philosophy"`
	KeyChanges   []string `json:"key_changes"`
}

// NewGenerator creates a new skill generator
func NewGenerator(db *sql.DB, llm *llm.Client, logger *zap.Logger, outputDir string) *Generator {
	if outputDir == "" {
		outputDir = "skills"
	}
	return &Generator{
		db:        db,
		llm:       llm,
		logger:    logger,
		outputDir: outputDir,
	}
}

// List returns all generated skills
func (g *Generator) List() ([]string, error) {
	if _, err := os.Stat(g.outputDir); os.IsNotExist(err) {
		return []string{}, nil
	}

	entries, err := os.ReadDir(g.outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var skills []string
	for _, entry := range entries {
		if entry.IsDir() {
			skillFile := filepath.Join(g.outputDir, entry.Name(), "SKILL.md")
			if _, err := os.Stat(skillFile); err == nil {
				skills = append(skills, entry.Name())
			}
		}
	}

	return skills, nil
}

// Generate creates skill files from PR analysis
func (g *Generator) Generate(ctx context.Context) ([]string, error) {
	// Fetch all analyses grouped by category
	analysisMap, err := g.fetchAnalysesByCategory()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analyses: %w", err)
	}

	if len(analysisMap) == 0 {
		g.logger.Info("No analyses found for skill generation")
		return []string{}, nil
	}

	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate skill for each category
	var generatedSkills []string
	for category, analyses := range analysisMap {
		select {
		case <-ctx.Done():
			return generatedSkills, ctx.Err()
		default:
		}

		if err := g.generateSkill(ctx, category, analyses); err != nil {
			g.logger.Warn("Failed to generate skill", zap.String("category", category), zap.Error(err))
			continue
		}

		generatedSkills = append(generatedSkills, category)
		g.logger.Info("Generated skill", zap.String("category", category), zap.Int("pr_count", len(analyses)))
	}

	// Also generate a summary skill
	if err := g.generateSummarySkill(ctx, analysisMap); err != nil {
		g.logger.Warn("Failed to generate summary skill", zap.Error(err))
	} else {
		generatedSkills = append(generatedSkills, "go-weekly-digest")
	}

	sort.Strings(generatedSkills)
	return generatedSkills, nil
}

func (g *Generator) fetchAnalysesByCategory() (map[string][]PRAnalysisData, error) {
	rows, err := g.db.Query(`
		SELECT
			pr.id, pr.github_id, pr.title, pr.state, pr.author,
			pa.analysis, pr.updated_at
		FROM pull_requests pr
		JOIN pr_analyses pa ON pr.id = pa.pr_id
		WHERE pr.updated_at > datetime('now', '-30 days')
		ORDER BY pr.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]PRAnalysisData)
	for rows.Next() {
		var data PRAnalysisData
		if err := rows.Scan(
			&data.PRID, &data.PRNumber, &data.Title, &data.State, &data.Author,
			&data.Analysis, &data.UpdatedAt,
		); err != nil {
			g.logger.Warn("Failed to scan analysis", zap.Error(err))
			continue
		}

		// Parse analysis to get category
		var parsed ParsedAnalysis
		if err := json.Unmarshal([]byte(data.Analysis), &parsed); err != nil {
			parsed.Category = "other"
		}

		category := parsed.Category
		if category == "" {
			category = "other"
		}

		result[category] = append(result[category], data)
	}

	return result, nil
}

func (g *Generator) generateSkill(ctx context.Context, category string, analyses []PRAnalysisData) error {
	// Create skill directory
	skillDir := filepath.Join(g.outputDir, "go-"+category)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// Generate content using LLM or template
	content, err := g.generateSkillContent(ctx, category, analyses)
	if err != nil {
		g.logger.Warn("LLM generation failed, using template", zap.Error(err))
		content = g.generateTemplateContent(category, analyses)
	}

	// Write skill file
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	return nil
}

func (g *Generator) generateSkillContent(ctx context.Context, category string, analyses []PRAnalysisData) (string, error) {
	// Build prompt with PR analyses
	var prSummaries strings.Builder
	for i, data := range analyses {
		if i >= 10 { // Limit to 10 PRs
			break
		}
		var parsed ParsedAnalysis
		json.Unmarshal([]byte(data.Analysis), &parsed)

		prSummaries.WriteString(fmt.Sprintf("### PR #%d: %s\n", data.PRNumber, data.Title))
		prSummaries.WriteString(fmt.Sprintf("- 状態: %s\n", data.State))
		prSummaries.WriteString(fmt.Sprintf("- 概要: %s\n", parsed.Summary))
		if parsed.Discussion != "" {
			prSummaries.WriteString(fmt.Sprintf("- 議論: %s\n", parsed.Discussion))
		}
		if parsed.GoPhilosophy != "" {
			prSummaries.WriteString(fmt.Sprintf("- Go思想: %s\n", parsed.GoPhilosophy))
		}
		prSummaries.WriteString("\n")
	}

	prompt := fmt.Sprintf(`以下のGo言語公式リポジトリのPR分析結果を基に、Claude Code用のスキルファイル（SKILL.md）を生成してください。

## カテゴリ: %s

## 分析済みPR

%s

## 出力形式

以下の形式でSKILL.mdを生成してください：

---
name: go-%s
description: golang/goリポジトリの%s関連PRから抽出したGoの知見とベストプラクティス。
---

# Go %s の最新動向

## 概要
[このカテゴリの全体的な傾向や重要なポイント]

## 注目のPR

[各PRの内容と学びをまとめる]

## Goの思想・哲学

[PRから読み取れるGoの設計思想]

## 実践的な学び

[開発者が活かせる具体的な知見]

---

日本語で出力してください。`, category, prSummaries.String(), category, category, formatCategoryTitle(category))

	return g.llm.Generate(ctx, prompt)
}

func (g *Generator) generateTemplateContent(category string, analyses []PRAnalysisData) string {
	var prList strings.Builder
	for i, data := range analyses {
		if i >= 10 {
			break
		}
		var parsed ParsedAnalysis
		json.Unmarshal([]byte(data.Analysis), &parsed)

		prList.WriteString(fmt.Sprintf("### PR #%d: %s\n\n", data.PRNumber, data.Title))
		prList.WriteString(fmt.Sprintf("**状態**: %s | **作成者**: %s\n\n", data.State, data.Author))
		if parsed.Summary != "" {
			prList.WriteString(fmt.Sprintf("**概要**: %s\n\n", parsed.Summary))
		}
		if parsed.Discussion != "" {
			prList.WriteString(fmt.Sprintf("**議論のポイント**: %s\n\n", parsed.Discussion))
		}
		if len(parsed.Insights) > 0 {
			prList.WriteString("**学び**:\n")
			for _, insight := range parsed.Insights {
				prList.WriteString(fmt.Sprintf("- %s\n", insight))
			}
			prList.WriteString("\n")
		}
		if parsed.GoPhilosophy != "" {
			prList.WriteString(fmt.Sprintf("**Go思想**: %s\n\n", parsed.GoPhilosophy))
		}
		prList.WriteString("---\n\n")
	}

	return fmt.Sprintf(`---
name: go-%s
description: golang/goリポジトリの%s関連PRから抽出したGoの知見とベストプラクティス。このカテゴリに関する質問やコードレビュー時に使用。
---

# Go %s の最新動向

> 📅 最終更新: %s
> 📊 分析PR数: %d件

## 概要

このスキルは golang/go リポジトリの直近1ヶ月の%s関連PRを分析し、Goの最新の設計思想や議論をまとめたものです。

## 注目のPR

%s

## このスキルの活用方法

- %s関連のコードを書く際の参考に
- コードレビュー時のチェックポイントとして
- Goの最新動向のキャッチアップに

## 参考リンク

- [golang/go リポジトリ](https://github.com/golang/go)
- [Go 公式ドキュメント](https://go.dev/doc/)
`, category, category, formatCategoryTitle(category),
		time.Now().Format("2006-01-02"),
		len(analyses), category, prList.String(), category)
}

func (g *Generator) generateSummarySkill(ctx context.Context, analysisMap map[string][]PRAnalysisData) error {
	skillDir := filepath.Join(g.outputDir, "go-weekly-digest")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return err
	}

	var categories []string
	var totalPRs int
	for cat, analyses := range analysisMap {
		categories = append(categories, cat)
		totalPRs += len(analyses)
	}
	sort.Strings(categories)

	var content strings.Builder
	content.WriteString(fmt.Sprintf(`---
name: go-weekly-digest
description: golang/goリポジトリの直近1ヶ月のPR動向サマリー。Goの最新動向を把握したいときに使用。
---

# Go Weekly Digest

> 📅 最終更新: %s
> 📊 分析PR総数: %d件

## カテゴリ別の動向

`, time.Now().Format("2006-01-02"), totalPRs))

	for _, cat := range categories {
		analyses := analysisMap[cat]
		content.WriteString(fmt.Sprintf("### %s (%d件)\n\n", formatCategoryTitle(cat), len(analyses)))

		// List top 3 PRs
		for i, data := range analyses {
			if i >= 3 {
				break
			}
			var parsed ParsedAnalysis
			json.Unmarshal([]byte(data.Analysis), &parsed)
			content.WriteString(fmt.Sprintf("- **#%d**: %s\n", data.PRNumber, data.Title))
			if parsed.Summary != "" {
				content.WriteString(fmt.Sprintf("  - %s\n", truncate(parsed.Summary, 100)))
			}
		}
		content.WriteString("\n")
	}

	content.WriteString(`## 詳細スキル

各カテゴリの詳細は以下のスキルを参照してください：

`)
	for _, cat := range categories {
		content.WriteString(fmt.Sprintf("- `go-%s`: %s関連の詳細\n", cat, formatCategoryTitle(cat)))
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	return os.WriteFile(skillFile, []byte(content.String()), 0644)
}

func formatCategoryTitle(category string) string {
	titles := map[string]string{
		"error-handling": "エラーハンドリング",
		"testing":        "テスト",
		"performance":    "パフォーマンス",
		"concurrency":    "並行処理",
		"api-design":     "API設計",
		"tooling":        "ツール",
		"documentation":  "ドキュメント",
		"other":          "その他",
	}
	if title, ok := titles[category]; ok {
		return title
	}
	return category
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
