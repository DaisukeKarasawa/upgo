package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"upgo/internal/llm"

	"go.uber.org/zap"
)

type AnalysisService struct {
	db         *sql.DB
	summarizer *llm.Summarizer
	analyzer   *llm.Analyzer
	logger     *zap.Logger
}

func NewAnalysisService(
	db *sql.DB,
	summarizer *llm.Summarizer,
	analyzer *llm.Analyzer,
	logger *zap.Logger,
) *AnalysisService {
	return &AnalysisService{
		db:         db,
		summarizer: summarizer,
		analyzer:   analyzer,
		logger:     logger,
	}
}

func (s *AnalysisService) AnalyzePR(ctx context.Context, prID int) error {
	var pr struct {
		Body   string
		State  string
		Title  string
		Author string
	}

	err := s.db.QueryRow(
		"SELECT title, body, state, author FROM pull_requests WHERE id = ?",
		prID,
	).Scan(&pr.Title, &pr.Body, &pr.State, &pr.Author)

	if err != nil {
		return fmt.Errorf("PR情報の取得に失敗しました: %w", err)
	}

	rows, err := s.db.Query(
		"SELECT body FROM pull_request_comments WHERE pr_id = ? ORDER BY created_at",
		prID,
	)
	if err != nil {
		return fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var comments []string
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			continue
		}
		comments = append(comments, body)
	}

	if err := rows.Err(); err != nil {
		s.logger.Warn("PRコメントの取得中にエラーが発生しました", zap.Error(err))
	}

	var diff string
	err = s.db.QueryRow(
		"SELECT diff_text FROM pull_request_diffs WHERE pr_id = ? LIMIT 1",
		prID,
	).Scan(&diff)
	if err != nil && err != sql.ErrNoRows {
		s.logger.Warn("差分の取得に失敗しました", zap.Error(err))
	}

	descriptionSummary, err := s.summarizer.SummarizeDescription(ctx, pr.Body)
	if err != nil {
		s.logger.Warn("説明の要約に失敗しました", zap.Error(err))
	}

	diffSummary, diffExplanation, err := s.summarizer.SummarizeDiff(ctx, diff)
	if err != nil {
		s.logger.Warn("差分の要約に失敗しました", zap.Error(err))
	}

	commentsSummary, discussionSummary, err := s.summarizer.SummarizeComments(ctx, comments)
	if err != nil {
		s.logger.Warn("コメントの要約に失敗しました", zap.Error(err))
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO pull_request_summaries 
		(pr_id, description_summary, diff_summary, diff_explanation, comments_summary, discussion_summary, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		prID, descriptionSummary, diffSummary, diffExplanation, commentsSummary, discussionSummary, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("要約の保存に失敗しました: %w", err)
	}

	if pr.State == "merged" {
		prInfo := fmt.Sprintf("タイトル: %s\n説明: %s\n作成者: %s", pr.Title, pr.Body, pr.Author)
		// Format comments with indexed structure for better LLM understanding
		commentsText := ""
		for i, comment := range comments {
			commentsText += fmt.Sprintf("コメント%d: %s\n\n", i+1, comment)
		}
		mergeReason, err := s.analyzer.AnalyzeMergeReason(ctx, prInfo, commentsText, discussionSummary)
		if err != nil {
			s.logger.Warn("Merge理由の分析に失敗しました", zap.Error(err))
		} else {
			_, err = s.db.Exec(
				"UPDATE pull_request_summaries SET merge_reason = ? WHERE pr_id = ?",
				mergeReason, prID,
			)
			if err != nil {
				s.logger.Warn("Merge理由の保存に失敗しました", zap.Error(err))
			}
		}
	} else if pr.State == "closed" {
		prInfo := fmt.Sprintf("タイトル: %s\n説明: %s\n作成者: %s", pr.Title, pr.Body, pr.Author)
		// Format comments with indexed structure for better LLM understanding
		commentsText := ""
		for i, comment := range comments {
			commentsText += fmt.Sprintf("コメント%d: %s\n\n", i+1, comment)
		}
		closeReason, err := s.analyzer.AnalyzeCloseReason(ctx, prInfo, commentsText, discussionSummary)
		if err != nil {
			s.logger.Warn("Close理由の分析に失敗しました", zap.Error(err))
		} else {
			_, err = s.db.Exec(
				"UPDATE pull_request_summaries SET close_reason = ? WHERE pr_id = ?",
				closeReason, prID,
			)
			if err != nil {
				s.logger.Warn("Close理由の保存に失敗しました", zap.Error(err))
			}
		}
	}

	return nil
}

func (s *AnalysisService) AnalyzeIssue(ctx context.Context, issueID int) error {
	var issue struct {
		Body   string
		State  string
		Title  string
		Author string
	}

	err := s.db.QueryRow(
		"SELECT title, body, state, author FROM issues WHERE id = ?",
		issueID,
	).Scan(&issue.Title, &issue.Body, &issue.State, &issue.Author)

	if err != nil {
		return fmt.Errorf("Issue情報の取得に失敗しました: %w", err)
	}

	rows, err := s.db.Query(
		"SELECT body FROM issue_comments WHERE issue_id = ? ORDER BY created_at",
		issueID,
	)
	if err != nil {
		return fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}
	defer rows.Close()

	var comments []string
	for rows.Next() {
		var body string
		if err := rows.Scan(&body); err != nil {
			continue
		}
		comments = append(comments, body)
	}

	if err := rows.Err(); err != nil {
		s.logger.Warn("Issueコメントの取得中にエラーが発生しました", zap.Error(err))
	}

	descriptionSummary, err := s.summarizer.SummarizeDescription(ctx, issue.Body)
	if err != nil {
		s.logger.Warn("説明の要約に失敗しました", zap.Error(err))
	}

	commentsSummary, discussionSummary, err := s.summarizer.SummarizeComments(ctx, comments)
	if err != nil {
		s.logger.Warn("コメントの要約に失敗しました", zap.Error(err))
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO issue_summaries 
		(issue_id, description_summary, comments_summary, discussion_summary, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		issueID, descriptionSummary, commentsSummary, discussionSummary, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("要約の保存に失敗しました: %w", err)
	}

	return nil
}
