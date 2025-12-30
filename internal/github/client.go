package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Client struct {
	client  *github.Client
	logger  *zap.Logger
	limiter *rate.Limiter
}

func NewClient(token string, logger *zap.Logger) *Client {
	client := github.NewClient(nil).WithAuthToken(token)

	// レート制限: 認証済みリクエストは5000リクエスト/時間
	// 安全のため、4500リクエスト/時間に制限
	limiter := rate.NewLimiter(rate.Limit(4500/3600), 10)

	return &Client{
		client:  client,
		logger:  logger,
		limiter: limiter,
	}
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

func (c *Client) checkRateLimit(ctx context.Context) error {
	rateLimit, _, err := c.client.RateLimits(ctx)
	if err != nil {
		return fmt.Errorf("レート制限の確認に失敗しました: %w", err)
	}

	core := rateLimit.Core
	if core.Remaining < 100 {
		resetTime := core.Reset.Time
		waitTime := time.Until(resetTime)
		if waitTime > 0 {
			c.logger.Warn("レート制限に近づいています。待機します",
				zap.Int("remaining", core.Remaining),
				zap.Duration("wait_time", waitTime),
			)
			time.Sleep(waitTime)
		}
	}

	return nil
}

func (c *Client) GetClient() *github.Client {
	return c.client
}
