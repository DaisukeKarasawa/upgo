package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Analyzer struct {
	client *Client
	logger *zap.Logger
}

func NewAnalyzer(client *Client, logger *zap.Logger) *Analyzer {
	return &Analyzer{
		client: client,
		logger: logger,
	}
}

func (a *Analyzer) AnalyzeMergeReason(ctx context.Context, prInfo, comments, discussion string) (string, error) {
	// Sanitize all inputs to prevent prompt injection
	sanitizedPRInfo := sanitizeInput(prInfo)
	sanitizedComments := sanitizeInput(comments)
	sanitizedDiscussion := sanitizeInput(discussion)
	content := fmt.Sprintf("PR情報:\n%s\n\nコメント:\n%s\n\n議論:\n%s", sanitizedPRInfo, sanitizedComments, sanitizedDiscussion)
	prompt := fmt.Sprintf(PromptMergeReason, content)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.client.timeout)*time.Second)
	defer cancel()

	result, err := a.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", fmt.Errorf("Merge理由の分析に失敗しました: %w", err)
	}

	return result, nil
}

func (a *Analyzer) AnalyzeCloseReason(ctx context.Context, prInfo, comments, discussion string) (string, error) {
	// Sanitize all inputs to prevent prompt injection
	sanitizedPRInfo := sanitizeInput(prInfo)
	sanitizedComments := sanitizeInput(comments)
	sanitizedDiscussion := sanitizeInput(discussion)
	content := fmt.Sprintf("PR情報:\n%s\n\nコメント:\n%s\n\n議論:\n%s", sanitizedPRInfo, sanitizedComments, sanitizedDiscussion)
	prompt := fmt.Sprintf(PromptCloseReason, content)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.client.timeout)*time.Second)
	defer cancel()

	result, err := a.client.Generate(timeoutCtx, prompt)
	if err != nil {
		return "", fmt.Errorf("Close理由の分析に失敗しました: %w", err)
	}

	return result, nil
}

func (a *Analyzer) AnalyzeMentalModel(ctx context.Context, analysisType string, prsData, issuesData string) (string, error) {
	// Sanitize all inputs to prevent prompt injection
	sanitizedPRsData := sanitizeInput(prsData)
	sanitizedIssuesData := sanitizeInput(issuesData)
	sanitizedAnalysisType := sanitizeInput(analysisType)
	
	content := fmt.Sprintf("PRデータ:\n%s\n\nIssueデータ:\n%s", sanitizedPRsData, sanitizedIssuesData)
	
	// Customize prompt based on analysis type
	// First, format the base prompt with content
	fullPrompt := fmt.Sprintf(PromptMentalModel, content)
	if sanitizedAnalysisType != "" {
		// Then append the analysis type instruction
		fullPrompt = fmt.Sprintf("%s\n\n特に以下の観点に焦点を当ててください: %s", fullPrompt, sanitizedAnalysisType)
	}
	
	// Mental model analysis takes longer, so use double timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.client.timeout)*time.Second*2)
	defer cancel()

	result, err := a.client.Generate(timeoutCtx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("メンタルモデル分析に失敗しました: %w", err)
	}

	return result, nil
}

// RetryWithBackoff implements exponential backoff retry logic for LLM API calls.
// It uses a predefined backoff sequence (1s, 2s, 4s) for the first few retries,
// then switches to exponential backoff for additional retries beyond the sequence length.
//
// Connection errors are immediately returned without retry because they indicate
// a fundamental connectivity issue that won't be resolved by waiting. This prevents
// wasting time on retries that are guaranteed to fail.
func (a *Analyzer) RetryWithBackoff(ctx context.Context, fn func() (string, error), maxRetries int) (string, error) {
	var lastErr error
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for i := 0; i < maxRetries; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		
		// Fail immediately on connection errors - these won't be resolved by retrying
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "connect") {
			return "", err
		}

		if i < maxRetries-1 {
			// Use predefined backoff values for first few retries, then exponential backoff
			var waitTime time.Duration
			if i < len(backoff) {
				waitTime = backoff[i]
			} else {
				// Exponential backoff: 2^(i-len(backoff)+1) times the last backoff value
				waitTime = backoff[len(backoff)-1] * time.Duration(1<<uint(i-len(backoff)+1))
			}
			a.logger.Warn("リトライします", zap.Int("attempt", i+1), zap.Duration("wait", waitTime), zap.Error(err))
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("最大リトライ回数に達しました: %w", lastErr)
}
