// Package github provides a GitHub API client with rate limiting support.
// It wraps the go-github client to enforce rate limits and prevent API quota exhaustion.
package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Client wraps the GitHub API client with rate limiting capabilities.
// It ensures that API requests respect GitHub's rate limits to avoid
// hitting quota restrictions that would block subsequent requests.
type Client struct {
	client  *github.Client
	logger  *zap.Logger
	limiter *rate.Limiter
}

// NewClient creates a new GitHub client with rate limiting configured.
// The rate limiter is set to 4500 requests per hour (GitHub's authenticated user limit)
// with a burst capacity of 10 requests to allow short bursts while maintaining
// the overall rate limit.
func NewClient(token string, logger *zap.Logger) *Client {
	client := github.NewClient(nil).WithAuthToken(token)
	// 4500 requests/hour = 1.25 requests/second, with burst of 10
	limiter := rate.NewLimiter(rate.Limit(float64(4500)/3600), 10)

	return &Client{
		client:  client,
		logger:  logger,
		limiter: limiter,
	}
}

// waitForRateLimit blocks until the rate limiter allows the next request.
// This is used before making API calls to ensure we don't exceed the configured rate limit.
func (c *Client) waitForRateLimit(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

// checkRateLimit checks the current GitHub API rate limit status and proactively
// waits if we're approaching the limit (remaining < 100). This prevents hitting
// the hard limit which would block all requests until the reset time.
//
// The threshold of 100 remaining requests is chosen to provide a safety buffer
// while still allowing reasonable throughput. Waiting proactively avoids the
// more disruptive scenario of hitting the limit mid-operation.
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

// GetClient returns the underlying go-github client instance.
// This method is provided for cases where direct access to the raw client
// is needed, bypassing the rate limiting wrapper.
func (c *Client) GetClient() *github.Client {
	return c.client
}
