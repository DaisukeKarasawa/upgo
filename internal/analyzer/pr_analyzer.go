package analyzer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"upgo/internal/llm"

	"go.uber.org/zap"
)

// PRAnalyzer analyzes PRs using LLM to extract Go insights
type PRAnalyzer struct {
	db     *sql.DB
	llm    *llm.Client
	logger *zap.Logger
}

// PRData represents PR data for analysis
type PRData struct {
	ID        int
	Number    int
	Title     string
	Body      string
	State     string
	Author    string
	Comments  string
	Diff      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PRAnalysis represents the analysis result
type PRAnalysis struct {
	Category    string   `json:"category"`
	Summary     string   `json:"summary"`
	Discussion  string   `json:"discussion"`
	Insights    []string `json:"insights"`
	GoPhilosophy string  `json:"go_philosophy"`
	KeyChanges  []string `json:"key_changes"`
}

// NewPRAnalyzer creates a new PR analyzer
func NewPRAnalyzer(db *sql.DB, llm *llm.Client, logger *zap.Logger) *PRAnalyzer {
	return &PRAnalyzer{
		db:     db,
		llm:    llm,
		logger: logger,
	}
}

// AnalyzeRecentPRs analyzes PRs that haven't been analyzed yet
func (a *PRAnalyzer) AnalyzeRecentPRs(ctx context.Context) (int, error) {
	// Get PRs that need analysis (no existing analysis or analysis is older than PR update)
	rows, err := a.db.Query(`
		SELECT
			pr.id, pr.github_id, pr.title, pr.body, pr.state, pr.author,
			pr.created_at, pr.updated_at,
			COALESCE(c.body, ''),
			COALESCE(d.diff_text, '')
		FROM pull_requests pr
		LEFT JOIN pull_request_comments c ON pr.id = c.pr_id
		LEFT JOIN pull_request_diffs d ON pr.id = d.pr_id
		LEFT JOIN pr_analyses pa ON pr.id = pa.pr_id
		WHERE pa.id IS NULL OR pa.updated_at < pr.updated_at
		ORDER BY pr.updated_at DESC
		LIMIT 50
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query PRs: %w", err)
	}
	defer rows.Close()

	var prs []PRData
	for rows.Next() {
		var pr PRData
		if err := rows.Scan(
			&pr.ID, &pr.Number, &pr.Title, &pr.Body, &pr.State, &pr.Author,
			&pr.CreatedAt, &pr.UpdatedAt, &pr.Comments, &pr.Diff,
		); err != nil {
			a.logger.Warn("Failed to scan PR", zap.Error(err))
			continue
		}
		prs = append(prs, pr)
	}

	if len(prs) == 0 {
		a.logger.Info("No PRs to analyze")
		return 0, nil
	}

	a.logger.Info("Analyzing PRs", zap.Int("count", len(prs)))

	analyzed := 0
	for _, pr := range prs {
		select {
		case <-ctx.Done():
			return analyzed, ctx.Err()
		default:
		}

		analysis, err := a.analyzePR(ctx, pr)
		if err != nil {
			a.logger.Warn("Failed to analyze PR", zap.Int("pr", pr.Number), zap.Error(err))
			continue
		}

		if err := a.saveAnalysis(pr.ID, analysis); err != nil {
			a.logger.Warn("Failed to save analysis", zap.Int("pr", pr.Number), zap.Error(err))
			continue
		}

		a.logger.Debug("Analyzed PR", zap.Int("pr", pr.Number), zap.String("category", analysis.Category))
		analyzed++
	}

	return analyzed, nil
}

func (a *PRAnalyzer) analyzePR(ctx context.Context, pr PRData) (*PRAnalysis, error) {
	prompt := buildAnalysisPrompt(pr)

	response, err := a.llm.Generate(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse JSON response
	var analysis PRAnalysis
	if err := json.Unmarshal([]byte(extractJSON(response)), &analysis); err != nil {
		// Fallback: create basic analysis from response
		analysis = PRAnalysis{
			Category:    categorizeFromTitle(pr.Title),
			Summary:     response,
			Discussion:  "",
			Insights:    []string{},
			GoPhilosophy: "",
			KeyChanges:  []string{},
		}
	}

	return &analysis, nil
}

func (a *PRAnalyzer) saveAnalysis(prID int, analysis *PRAnalysis) error {
	analysisJSON, err := json.Marshal(analysis)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(`
		INSERT INTO pr_analyses (pr_id, analysis, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(pr_id) DO UPDATE SET
		analysis = excluded.analysis,
		updated_at = excluded.updated_at`,
		prID, string(analysisJSON), time.Now(), time.Now(),
	)
	return err
}

func buildAnalysisPrompt(pr PRData) string {
	diffSummary := pr.Diff
	if len(diffSummary) > 5000 {
		diffSummary = diffSummary[:5000] + "\n... (truncated)"
	}

	return fmt.Sprintf(`あなたはGo言語のエキスパートです。以下のPRを分析し、Goの思想・哲学の観点から重要なポイントを抽出してください。

## PR情報

**タイトル**: %s
**番号**: #%d
**状態**: %s
**作成者**: %s

**説明**:
%s

**議論・レビューコメント**:
%s

**変更差分（一部）**:
%s

## 分析指示

以下のJSON形式で回答してください：

{
  "category": "カテゴリ（error-handling, testing, performance, concurrency, api-design, tooling, documentation, other のいずれか）",
  "summary": "このPRの概要（日本語で2-3文）",
  "discussion": "議論のポイント（日本語で、どのような議論が行われたか）",
  "insights": ["学び1", "学び2", "..."],
  "go_philosophy": "このPRから読み取れるGoの設計思想・哲学（日本語で）",
  "key_changes": ["主要な変更1", "主要な変更2", "..."]
}

JSONのみを出力してください。`, pr.Title, pr.Number, pr.State, pr.Author, pr.Body, pr.Comments, diffSummary)
}

func extractJSON(response string) string {
	// Find JSON in response
	start := -1
	end := -1
	depth := 0

	for i, c := range response {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 && start != -1 {
				end = i + 1
				break
			}
		}
	}

	if start != -1 && end != -1 {
		return response[start:end]
	}
	return "{}"
}

func categorizeFromTitle(title string) string {
	// Simple keyword-based categorization
	keywords := map[string][]string{
		"error-handling": {"error", "err", "panic", "recover"},
		"testing":        {"test", "bench", "fuzz"},
		"performance":    {"perf", "optimize", "fast", "slow", "memory"},
		"concurrency":    {"goroutine", "channel", "sync", "mutex", "race"},
		"api-design":     {"api", "http", "rpc", "handler"},
		"tooling":        {"cmd", "tool", "go build", "go mod"},
		"documentation":  {"doc", "comment", "readme"},
	}

	titleLower := title
	for category, words := range keywords {
		for _, word := range words {
			if containsIgnoreCase(titleLower, word) {
				return category
			}
		}
	}
	return "other"
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsIgnoreCase(s[1:], substr) || (len(s) >= len(substr) && equalFoldPrefix(s, substr)))
}

func equalFoldPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		c1 := s[i]
		c2 := prefix[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 'a' - 'A'
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 'a' - 'A'
		}
		if c1 != c2 {
			return false
		}
	}
	return true
}
