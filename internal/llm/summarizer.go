package llm

import (
	"context"
	"fmt"
	"strings"
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

	// Sanitize input to prevent prompt injection
	sanitizedDescription := sanitizeInput(description)
	prompt := fmt.Sprintf(PromptPRDescriptionSummary, sanitizedDescription)
	
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

	// Sanitize input to prevent prompt injection
	sanitizedDiff := sanitizeInput(diff)
	prompt := fmt.Sprintf(PromptDiffSummary, sanitizedDiff)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.client.timeout)*time.Second)
	defer cancel()

	result, err := s.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", "", fmt.Errorf("差分の要約に失敗しました: %w", err)
	}

	// Split the result into summary and explanation
	summary, explanation := s.splitDiffResult(result)

	return summary, explanation, nil
}

func (s *Summarizer) SummarizeComments(ctx context.Context, comments []string) (string, string, error) {
	if len(comments) == 0 {
		return "", "", nil
	}

	commentsText := ""
	for i, comment := range comments {
		// Sanitize each comment to prevent prompt injection
		sanitizedComment := sanitizeInput(comment)
		commentsText += fmt.Sprintf("コメント%d: %s\n\n", i+1, sanitizedComment)
	}

	prompt := fmt.Sprintf(PromptCommentsSummary, commentsText)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(s.client.timeout)*time.Second)
	defer cancel()

	result, err := s.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", "", fmt.Errorf("コメントの要約に失敗しました: %w", err)
	}

	// Split the result into comments summary and discussion summary
	commentsSummary, discussionSummary := s.splitCommentsResult(result)

	return commentsSummary, discussionSummary, nil
}

// splitDiffResult splits LLM result into summary and explanation
func (s *Summarizer) splitDiffResult(result string) (summary string, explanation string) {
	// Split by markers 【要約】 and 【解説】
	summaryMarker := "【要約】"
	explanationMarker := "【解説】"

	summaryStart := -1
	explanationStart := -1

	// Find marker positions (case-insensitive)
	resultLower := strings.ToLower(result)
	if idx := strings.Index(resultLower, strings.ToLower(summaryMarker)); idx != -1 {
		summaryStart = idx + len(summaryMarker)
	}
	if idx := strings.Index(resultLower, strings.ToLower(explanationMarker)); idx != -1 {
		explanationStart = idx + len(explanationMarker)
	}

	// Extract summary
	if summaryStart != -1 && explanationStart != -1 && explanationStart > summaryStart {
		summary = strings.TrimSpace(result[summaryStart:explanationStart])
		explanation = strings.TrimSpace(result[explanationStart:])
	} else if summaryStart != -1 {
		summary = strings.TrimSpace(result[summaryStart:])
		if explanationStart != -1 {
			explanation = strings.TrimSpace(result[explanationStart:])
		}
	} else if explanationStart != -1 {
		explanation = strings.TrimSpace(result[explanationStart:])
		summary = ""
	} else {
		// If markers are not found, split the result in half
		mid := len(result) / 2
		summary = strings.TrimSpace(result[:mid])
		explanation = strings.TrimSpace(result[mid:])
	}

	return summary, explanation
}

// splitCommentsResult splits LLM result into comments summary and discussion summary
func (s *Summarizer) splitCommentsResult(result string) (commentsSummary string, discussionSummary string) {
	// Split by markers 【コメント要約】 and 【議論要約】
	commentsMarker := "【コメント要約】"
	discussionMarker := "【議論要約】"

	commentsStart := -1
	discussionStart := -1

	// Find marker positions (case-insensitive)
	resultLower := strings.ToLower(result)
	if idx := strings.Index(resultLower, strings.ToLower(commentsMarker)); idx != -1 {
		commentsStart = idx + len(commentsMarker)
	}
	if idx := strings.Index(resultLower, strings.ToLower(discussionMarker)); idx != -1 {
		discussionStart = idx + len(discussionMarker)
	}

	// Extract comments summary
	if commentsStart != -1 && discussionStart != -1 && discussionStart > commentsStart {
		commentsSummary = strings.TrimSpace(result[commentsStart:discussionStart])
		discussionSummary = strings.TrimSpace(result[discussionStart:])
	} else if commentsStart != -1 {
		commentsSummary = strings.TrimSpace(result[commentsStart:])
		if discussionStart != -1 {
			discussionSummary = strings.TrimSpace(result[discussionStart:])
		}
	} else if discussionStart != -1 {
		discussionSummary = strings.TrimSpace(result[discussionStart:])
		commentsSummary = ""
	} else {
		// If markers are not found, split the result in half
		mid := len(result) / 2
		commentsSummary = strings.TrimSpace(result[:mid])
		discussionSummary = strings.TrimSpace(result[mid:])
	}

	return commentsSummary, discussionSummary
}
