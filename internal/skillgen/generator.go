package skillgen

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"upgo/internal/github"

	"go.uber.org/zap"
)

// PRData represents PR data for skill generation
type PRData struct {
	ID        int
	Title     string
	Body      string
	State     string
	Author    string
	Summary   string
	Analysis  string
	Diff      string
	Comments  []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// LLMClient interface for generating skill content
type LLMClient interface {
	GenerateSkillContent(ctx context.Context, category string, prData []PRData) (string, error)
}

// Generator generates Claude Code skills from PR analysis
type Generator struct {
	db        *sql.DB
	prFetcher github.PRFetcherInterface
	llm       LLMClient
	logger    *zap.Logger
	owner     string
	repo      string
	outputDir string
}

// NewGenerator creates a new skill generator
func NewGenerator(
	db *sql.DB,
	prFetcher github.PRFetcherInterface,
	llm LLMClient,
	logger *zap.Logger,
	owner, repo, outputDir string,
) *Generator {
	return &Generator{
		db:        db,
		prFetcher: prFetcher,
		llm:       llm,
		logger:    logger,
		owner:     owner,
		repo:      repo,
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

// Sync fetches PR data from GitHub and stores it in the database
func (g *Generator) Sync(ctx context.Context) error {
	g.logger.Info("Starting PR sync")

	// Get or create repository
	var repoID int
	err := g.db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		g.owner, g.repo,
	).Scan(&repoID)

	if err == sql.ErrNoRows {
		result, err := g.db.Exec(
			"INSERT INTO repositories (owner, name, last_synced_at) VALUES (?, ?, ?)",
			g.owner, g.repo, nil,
		)
		if err != nil {
			return fmt.Errorf("failed to create repository: %w", err)
		}
		id, _ := result.LastInsertId()
		repoID = int(id)
	} else if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}

	// Get last sync time
	var lastSyncedAt sql.NullTime
	g.db.QueryRow(
		"SELECT last_synced_at FROM repositories WHERE id = ?",
		repoID,
	).Scan(&lastSyncedAt)

	sinceTime := time.Now().AddDate(0, 0, -30) // Default: 30 days
	if lastSyncedAt.Valid {
		sinceTime = lastSyncedAt.Time.Add(-5 * time.Minute)
	}

	// Fetch PRs from GitHub
	for _, state := range []string{"open", "closed"} {
		prs, err := g.prFetcher.FetchPRsUpdatedSince(ctx, g.owner, g.repo, state, sinceTime)
		if err != nil {
			g.logger.Warn("Failed to fetch PRs", zap.String("state", state), zap.Error(err))
			continue
		}

		for _, pr := range prs {
			prState := pr.GetState()
			if !pr.GetMergedAt().IsZero() {
				prState = "merged"
			}

			_, err := g.db.Exec(`
				INSERT OR REPLACE INTO pull_requests
				(repository_id, github_id, title, body, state, author, created_at, updated_at, last_synced_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				repoID, pr.GetNumber(), pr.GetTitle(), pr.GetBody(), prState,
				pr.GetUser().GetLogin(), pr.GetCreatedAt().Time, pr.GetUpdatedAt().Time, time.Now(),
			)
			if err != nil {
				g.logger.Warn("Failed to save PR", zap.Int("number", pr.GetNumber()), zap.Error(err))
			}
		}

		g.logger.Info("Synced PRs", zap.String("state", state), zap.Int("count", len(prs)))
	}

	// Update last sync time
	_, err = g.db.Exec(
		"UPDATE repositories SET last_synced_at = ? WHERE id = ?",
		time.Now(), repoID,
	)
	if err != nil {
		g.logger.Warn("Failed to update last_synced_at", zap.Error(err))
	}

	g.logger.Info("PR sync completed")
	return nil
}

// Generate creates skill files from PR analysis
func (g *Generator) Generate(ctx context.Context) error {
	g.logger.Info("Starting skill generation")

	// Fetch PR data with summaries and analyses
	prDataByCategory, err := g.fetchPRDataByCategory()
	if err != nil {
		return fmt.Errorf("failed to fetch PR data: %w", err)
	}

	if len(prDataByCategory) == 0 {
		g.logger.Info("No PR data found for skill generation")
		return nil
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate skills for each category
	for category, prData := range prDataByCategory {
		if err := g.generateSkill(ctx, category, prData); err != nil {
			g.logger.Warn("Failed to generate skill", zap.String("category", category), zap.Error(err))
			continue
		}
		g.logger.Info("Generated skill", zap.String("category", category))
	}

	g.logger.Info("Skill generation completed")
	return nil
}

// Update updates existing skills with new PR data
func (g *Generator) Update(ctx context.Context) error {
	g.logger.Info("Starting skill update")

	// First sync new PRs
	if err := g.Sync(ctx); err != nil {
		g.logger.Warn("Sync failed during update", zap.Error(err))
	}

	// Then regenerate skills
	return g.Generate(ctx)
}

// fetchPRDataByCategory retrieves PR data grouped by category
func (g *Generator) fetchPRDataByCategory() (map[string][]PRData, error) {
	rows, err := g.db.Query(`
		SELECT
			pr.id, pr.title, pr.body, pr.state, pr.author,
			pr.created_at, pr.updated_at,
			COALESCE(ps.summary, ''),
			COALESCE(pa.analysis, '')
		FROM pull_requests pr
		LEFT JOIN pr_summaries ps ON pr.id = ps.pr_id
		LEFT JOIN pr_analyses pa ON pr.id = pa.pr_id
		WHERE ps.summary IS NOT NULL OR pa.analysis IS NOT NULL
		ORDER BY pr.updated_at DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]PRData)

	for rows.Next() {
		var data PRData
		if err := rows.Scan(
			&data.ID, &data.Title, &data.Body, &data.State, &data.Author,
			&data.CreatedAt, &data.UpdatedAt, &data.Summary, &data.Analysis,
		); err != nil {
			continue
		}

		// Categorize based on title/analysis
		category := g.categorizeData(data)
		result[category] = append(result[category], data)
	}

	return result, nil
}

// categorizeData determines the category for a PR
func (g *Generator) categorizeData(data PRData) string {
	title := strings.ToLower(data.Title)
	analysis := strings.ToLower(data.Analysis)
	combined := title + " " + analysis

	categories := map[string][]string{
		"error-handling": {"error", "exception", "panic", "recover"},
		"testing":        {"test", "coverage", "mock", "fixture"},
		"performance":    {"perf", "benchmark", "optimize", "fast"},
		"concurrency":    {"goroutine", "channel", "mutex", "sync"},
		"api-design":     {"api", "endpoint", "rest", "handler"},
		"code-review":    {"review", "refactor", "cleanup", "improve"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(combined, keyword) {
				return category
			}
		}
	}

	return "go-patterns"
}

// generateSkill creates a skill file for a category
func (g *Generator) generateSkill(ctx context.Context, category string, prData []PRData) error {
	skillDir := filepath.Join(g.outputDir, category)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	var content string
	var err error

	if g.llm != nil {
		content, err = g.llm.GenerateSkillContent(ctx, category, prData)
		if err != nil {
			g.logger.Warn("LLM generation failed, using template", zap.Error(err))
			content = g.generateTemplateContent(category, prData)
		}
	} else {
		content = g.generateTemplateContent(category, prData)
	}

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	return nil
}

// generateTemplateContent creates skill content from a template
func (g *Generator) generateTemplateContent(category string, prData []PRData) string {
	description := g.getCategoryDescription(category)

	var patterns strings.Builder
	for i, pr := range prData {
		if i >= 5 { // Limit to 5 patterns
			break
		}
		patterns.WriteString(fmt.Sprintf("- %s\n", pr.Title))
		if pr.Summary != "" {
			patterns.WriteString(fmt.Sprintf("  - %s\n", truncate(pr.Summary, 200)))
		}
	}

	return fmt.Sprintf(`---
name: %s
description: %s
---

# %s

## Overview

This skill provides guidance on %s patterns learned from PR analysis.

## Patterns from PR Analysis

%s

## Best Practices

Based on the analyzed PRs, follow these best practices:

1. Write clear, descriptive commit messages
2. Include tests for new functionality
3. Follow existing code conventions
4. Document complex logic

## References

Generated from %d analyzed PRs in %s/%s.
`, category, description, formatTitle(category), category, patterns.String(), len(prData), g.owner, g.repo)
}

func (g *Generator) getCategoryDescription(category string) string {
	descriptions := map[string]string{
		"error-handling": "Go error handling patterns and best practices. Use when working with error handling, custom error types, or error wrapping.",
		"testing":        "Go testing patterns and strategies. Use when writing tests, mocks, or test fixtures.",
		"performance":    "Go performance optimization techniques. Use when optimizing code or writing benchmarks.",
		"concurrency":    "Go concurrency patterns. Use when working with goroutines, channels, or synchronization.",
		"api-design":     "API design patterns for Go. Use when designing REST APIs or HTTP handlers.",
		"code-review":    "Code review insights and refactoring patterns. Use when reviewing or improving code.",
		"go-patterns":    "General Go patterns and idioms. Use for general Go development guidance.",
	}
	if desc, ok := descriptions[category]; ok {
		return desc
	}
	return "Go development patterns and best practices."
}

func formatTitle(category string) string {
	parts := strings.Split(category, "-")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, " ")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
