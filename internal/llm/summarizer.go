package llm

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Summarizer struct {
	client *Client
	logger *zap.Logger
}

func NewSummarizer(client *Client, logger *zap.Logger) *Summarizer {
	return &Summarizer{
		client: client,
		logger: logger,
	}
}

func (s *Summarizer) SummarizeDescription(ctx context.Context, description string) (string, error) {
	if description == "" {
		return "", nil
	}

	prompt := fmt.Sprintf(PromptPRDescriptionSummary, description)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.client.timeout)*time.Second)
	defer cancel()

	result, err := s.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", fmt.Errorf("説明の要約に失敗しました: %w", err)
	}

	return result, nil
}

func (s *Summarizer) SummarizeDiff(ctx context.Context, diff string) (string, string, error) {
	if diff == "" {
		return "", "", nil
	}

	prompt := fmt.Sprintf(PromptDiffSummary, diff)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.client.timeout)*time.Second)
	defer cancel()

	result, err := s.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", "", fmt.Errorf("差分の要約に失敗しました: %w", err)
	}

	// 結果を要約と解説に分割（簡易実装）
	// 実際の実装では、より高度な解析が必要かもしれません
	summary := result
	explanation := result

	return summary, explanation, nil
}

func (s *Summarizer) SummarizeComments(ctx context.Context, comments []string) (string, string, error) {
	if len(comments) == 0 {
		return "", "", nil
	}

	commentsText := ""
	for i, comment := range comments {
		commentsText += fmt.Sprintf("コメント%d: %s\n\n", i+1, comment)
	}

	prompt := fmt.Sprintf(PromptCommentsSummary, commentsText)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.client.timeout)*time.Second)
	defer cancel()

	result, err := s.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", "", fmt.Errorf("コメントの要約に失敗しました: %w", err)
	}

	// 結果をコメント要約と議論要約に分割（簡易実装）
	commentsSummary := result
	discussionSummary := result

	return commentsSummary, discussionSummary, nil
}
