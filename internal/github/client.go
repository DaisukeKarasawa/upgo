// Package github provides a GitHub API client with rate limiting support.
// It wraps the go-github client to enforce rate limits and prevent API quota exhaustion.
package github

import (
	"context"
	"fmt"
	"sync"
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

	// Rate limit cache to reduce API calls
	rateLimitCache struct {
		sync.RWMutex
		lastCheck    time.Time
		remaining    int
		resetTime    time.Time
		cacheExpiry  time.Duration // How long to cache rate limit info
	}
}

// NewClient creates a new GitHub client with rate limiting configured.
// The rate limiter is set to 4500 requests per hour (GitHub's authenticated user limit)
// with a burst capacity of 10 requests to allow short bursts while maintaining
// the overall rate limit.
func NewClient(token string, logger *zap.Logger) *Client {
	client := github.NewClient(nil).WithAuthToken(token)
	// 4500 requests/hour = 1.25 requests/second, with burst of 10
	limiter := rate.NewLimiter(rate.Limit(float64(4500)/3600), 10)

	c := &Client{
		client:  client,
		logger:  logger,
		limiter: limiter,
	}
	// Cache rate limit info for 60 seconds to reduce API calls
	c.rateLimitCache.cacheExpiry = 60 * time.Second

	return c
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
//
// Uses caching to reduce API calls: rate limit info is cached for 60 seconds.
func (c *Client) checkRateLimit(ctx context.Context) error {
	c.rateLimitCache.RLock()
	now := time.Now()
	// Check if cached data is still valid
	if !c.rateLimitCache.lastCheck.IsZero() &&
		now.Sub(c.rateLimitCache.lastCheck) < c.rateLimitCache.cacheExpiry {
		// Use cached data
		remaining := c.rateLimitCache.remaining
		resetTime := c.rateLimitCache.resetTime
		c.rateLimitCache.RUnlock()

		if remaining < 100 {
			waitTime := time.Until(resetTime)
			if waitTime > 0 {
				c.logger.Warn("レート制限に近づいています（キャッシュ情報）。待機します",
					zap.Int("remaining", remaining),
					zap.Duration("wait_time", waitTime),
				)
				time.Sleep(waitTime)
			}
		}
		return nil
	}
	c.rateLimitCache.RUnlock()

	// Cache expired or not set, fetch fresh data
	c.rateLimitCache.Lock()
	defer c.rateLimitCache.Unlock()

	// Double-check after acquiring write lock (another goroutine might have updated it)
	if !c.rateLimitCache.lastCheck.IsZero() &&
		now.Sub(c.rateLimitCache.lastCheck) < c.rateLimitCache.cacheExpiry {
		remaining := c.rateLimitCache.remaining
		resetTime := c.rateLimitCache.resetTime
		if remaining < 100 {
			waitTime := time.Until(resetTime)
			if waitTime > 0 {
				c.logger.Warn("レート制限に近づいています（キャッシュ情報）。待機します",
					zap.Int("remaining", remaining),
					zap.Duration("wait_time", waitTime),
				)
				time.Sleep(waitTime)
			}
		}
		return nil
	}

	// Fetch fresh rate limit info
	rateLimit, _, err := c.client.RateLimits(ctx)
	if err != nil {
		return fmt.Errorf("レート制限の確認に失敗しました: %w", err)
	}

	core := rateLimit.Core
	// Update cache
	c.rateLimitCache.lastCheck = now
	c.rateLimitCache.remaining = core.Remaining
	c.rateLimitCache.resetTime = core.Reset.Time

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
