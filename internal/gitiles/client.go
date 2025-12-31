package gitiles

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Client struct {
	httpClient *http.Client
	logger     *zap.Logger
	limiter    *rate.Limiter
	baseURL    string

	rateLimitCache struct {
		sync.RWMutex
		lastCheck   time.Time
		cacheExpiry time.Duration
	}
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	limiter := rate.NewLimiter(rate.Limit(10), 20)

	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		logger:     logger,
		limiter:    limiter,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}
	c.rateLimitCache.cacheExpiry = 60 * time.Second

	return c
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

type RefInfo struct {
	Value  string `json:"value"`
	Peeled string `json:"peeled,omitempty"`
	Target string `json:"target,omitempty"`
}

type RefsResponse map[string]RefInfo

type CommitInfo struct {
	Commit    string         `json:"commit"`
	Tree      string         `json:"tree"`
	Parents   []ParentInfo   `json:"parents"`
	Author    GitPersonInfo  `json:"author"`
	Committer GitPersonInfo  `json:"committer"`
	Message   string         `json:"message"`
	TreeDiff  []TreeDiffInfo `json:"tree_diff,omitempty"`
}

type ParentInfo struct {
	Commit string `json:"commit"`
}

type GitPersonInfo struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Time  time.Time `json:"time"`
}

type TreeDiffInfo struct {
	Type    string `json:"type"`
	OldID   string `json:"old_id,omitempty"`
	OldMode int    `json:"old_mode,omitempty"`
	OldPath string `json:"old_path,omitempty"`
	NewID   string `json:"new_id,omitempty"`
	NewMode int    `json:"new_mode,omitempty"`
	NewPath string `json:"new_path,omitempty"`
}

type LogResponse struct {
	Log  []CommitInfo `json:"log"`
	Next string       `json:"next,omitempty"`
}

func (c *Client) doRequest(ctx context.Context, path string) ([]byte, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	url := fmt.Sprintf("%s/%s?format=JSON", c.baseURL, strings.TrimPrefix(path, "/"))
	c.logger.Debug("Gitilesリクエストを送信しています", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("リクエストの作成に失敗しました: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("リクエストの送信に失敗しました: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("予期しないステータスコード: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("レスポンスの読み取りに失敗しました: %w", err)
	}

	if len(body) > 4 && string(body[:4]) == ")]}'"{
		body = body[4:]
	}
	body = []byte(strings.TrimLeft(string(body), "\n"))

	return body, nil
}

func (c *Client) GetRefs(ctx context.Context, project string) (RefsResponse, error) {
	path := fmt.Sprintf("%s/+refs", project)

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("refs の取得に失敗しました: %w", err)
	}

	var refs RefsResponse
	if err := json.Unmarshal(body, &refs); err != nil {
		return nil, fmt.Errorf("refs のパースに失敗しました: %w", err)
	}

	c.logger.Info("refs を取得しました", zap.Int("count", len(refs)))
	return refs, nil
}

func (c *Client) GetLog(ctx context.Context, project, ref string, limit int) (*LogResponse, error) {
	path := fmt.Sprintf("%s/+log/%s", project, ref)
	if limit > 0 {
		path = fmt.Sprintf("%s?n=%d", path, limit)
	}

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("ログの取得に失敗しました: %w", err)
	}

	var logResp LogResponse
	if err := json.Unmarshal(body, &logResp); err != nil {
		return nil, fmt.Errorf("ログのパースに失敗しました: %w", err)
	}

	c.logger.Info("コミットログを取得しました",
		zap.String("ref", ref),
		zap.Int("count", len(logResp.Log)))
	return &logResp, nil
}

func (c *Client) GetCommit(ctx context.Context, project, commit string) (*CommitInfo, error) {
	path := fmt.Sprintf("%s/+/%s", project, commit)

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("コミットの取得に失敗しました: %w", err)
	}

	var commitInfo CommitInfo
	if err := json.Unmarshal(body, &commitInfo); err != nil {
		return nil, fmt.Errorf("コミットのパースに失敗しました: %w", err)
	}

	return &commitInfo, nil
}

func (c *Client) GetDiff(ctx context.Context, project, commit string) (string, error) {
	path := fmt.Sprintf("%s/+/%s^!", project, commit)

	body, err := c.doRequest(ctx, path)
	if err != nil {
		return "", fmt.Errorf("差分の取得に失敗しました: %w", err)
	}

	return string(body), nil
}

func ParseChangeIDFromMessage(message string) string {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Change-Id:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Change-Id:"))
		}
	}
	return ""
}

func ParseReviewedOnFromMessage(message string) string {
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Reviewed-on:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "Reviewed-on:"))
		}
	}
	return ""
}
