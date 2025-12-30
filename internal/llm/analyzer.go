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
	content := fmt.Sprintf("PR情報:\n%s\n\nコメント:\n%s\n\n議論:\n%s", prInfo, comments, discussion)
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
	content := fmt.Sprintf("PR情報:\n%s\n\nコメント:\n%s\n\n議論:\n%s", prInfo, comments, discussion)
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
	content := fmt.Sprintf("PRデータ:\n%s\n\nIssueデータ:\n%s", prsData, issuesData)
	
	// 分析タイプに応じたプロンプトのカスタマイズ
	prompt := PromptMentalModel
	if analysisType != "" {
		prompt = fmt.Sprintf("%s\n\n特に以下の観点に焦点を当ててください: %s", PromptMentalModel, analysisType)
	}
	
	fullPrompt := fmt.Sprintf(prompt, content)
	
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(a.client.timeout)*time.Second*2) // メンタルモデル分析は時間がかかるため
	defer cancel()

	result, err := a.client.Generate(timeoutCtx, fullPrompt)
	if err != nil {
		return "", fmt.Errorf("メンタルモデル分析に失敗しました: %w", err)
	}

	return result, nil
}

func (a *Analyzer) RetryWithBackoff(ctx context.Context, fn func() (string, error), maxRetries int) (string, error) {
	var lastErr error
	backoff := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}

	for i := 0; i < maxRetries; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err
		
		// 接続エラーの場合は即座に失敗
		if strings.Contains(err.Error(), "connection") || strings.Contains(err.Error(), "connect") {
			return "", err
		}

		if i < maxRetries-1 {
			waitTime := backoff[i]
			a.logger.Warn("リトライします", zap.Int("attempt", i+1), zap.Duration("wait", waitTime), zap.Error(err))
			time.Sleep(waitTime)
		}
	}

	return "", fmt.Errorf("最大リトライ回数に達しました: %w", lastErr)
}
