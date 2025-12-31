package gerrit

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/build/gerrit"
	"golang.org/x/time/rate"
)

type Client struct {
	client  *gerrit.Client
	logger  *zap.Logger
	limiter *rate.Limiter
	baseURL string

	rateLimitCache struct {
		sync.RWMutex
		lastCheck   time.Time
		cacheExpiry time.Duration
	}
}

func NewClient(baseURL string, logger *zap.Logger) *Client {
	client := gerrit.NewClient(baseURL, gerrit.NoAuth)
	limiter := rate.NewLimiter(rate.Limit(10), 20)

	c := &Client{
		client:  client,
		logger:  logger,
		limiter: limiter,
		baseURL: baseURL,
	}
	c.rateLimitCache.cacheExpiry = 60 * time.Second

	return c
}

func (c *Client) waitForRateLimit(ctx context.Context) error {
	return c.limiter.Wait(ctx)
}

type QueryOptions struct {
	Project  string
	Branches []string
	Status   []string
	After    time.Time
	Limit    int
	Start    int
}

func (c *Client) buildQuery(opts QueryOptions) string {
	var parts []string

	if opts.Project != "" {
		parts = append(parts, fmt.Sprintf("project:%s", opts.Project))
	}

	if len(opts.Branches) > 0 {
		var branchParts []string
		for _, branch := range opts.Branches {
			if strings.Contains(branch, "*") {
				branchParts = append(branchParts, fmt.Sprintf("branch:^%s", strings.ReplaceAll(branch, "*", ".*")))
			} else {
				branchParts = append(branchParts, fmt.Sprintf("branch:%s", branch))
			}
		}
		if len(branchParts) > 1 {
			parts = append(parts, "("+strings.Join(branchParts, " OR ")+")")
		} else if len(branchParts) == 1 {
			parts = append(parts, branchParts[0])
		}
	}

	if len(opts.Status) > 0 {
		var statusParts []string
		for _, status := range opts.Status {
			statusParts = append(statusParts, fmt.Sprintf("status:%s", status))
		}
		if len(statusParts) > 1 {
			parts = append(parts, "("+strings.Join(statusParts, " OR ")+")")
		} else if len(statusParts) == 1 {
			parts = append(parts, statusParts[0])
		}
	}

	if !opts.After.IsZero() {
		parts = append(parts, fmt.Sprintf("after:%s", opts.After.Format("2006-01-02")))
	}

	return strings.Join(parts, " ")
}

func (c *Client) QueryChanges(ctx context.Context, opts QueryOptions) ([]*gerrit.ChangeInfo, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	query := c.buildQuery(opts)
	c.logger.Debug("Gerrit変更を検索しています", zap.String("query", query))

	queryOpts := []gerrit.QueryChangesOpt{
		gerrit.QueryChangesOpt("o=DETAILED_ACCOUNTS"),
		gerrit.QueryChangesOpt("o=LABELS"),
		gerrit.QueryChangesOpt("o=CURRENT_REVISION"),
		gerrit.QueryChangesOpt("o=MESSAGES"),
	}

	if opts.Limit > 0 {
		queryOpts = append(queryOpts, gerrit.QueryChangesOpt(fmt.Sprintf("n=%d", opts.Limit)))
	}
	if opts.Start > 0 {
		queryOpts = append(queryOpts, gerrit.QueryChangesOpt(fmt.Sprintf("S=%d", opts.Start)))
	}

	changes, err := c.client.QueryChanges(ctx, query, queryOpts...)
	if err != nil {
		return nil, fmt.Errorf("変更の検索に失敗しました: %w", err)
	}

	c.logger.Info("Gerrit変更を取得しました", zap.Int("count", len(changes)))
	return changes, nil
}

func (c *Client) GetChangeDetail(ctx context.Context, changeID string) (*gerrit.ChangeInfo, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	c.logger.Debug("変更詳細を取得しています", zap.String("change_id", changeID))

	opts := []gerrit.QueryChangesOpt{
		gerrit.QueryChangesOpt("o=DETAILED_ACCOUNTS"),
		gerrit.QueryChangesOpt("o=ALL_REVISIONS"),
		gerrit.QueryChangesOpt("o=ALL_COMMITS"),
		gerrit.QueryChangesOpt("o=ALL_FILES"),
		gerrit.QueryChangesOpt("o=LABELS"),
		gerrit.QueryChangesOpt("o=DETAILED_LABELS"),
		gerrit.QueryChangesOpt("o=MESSAGES"),
		gerrit.QueryChangesOpt("o=CURRENT_ACTIONS"),
	}

	change, err := c.client.GetChangeDetail(ctx, changeID, opts...)
	if err != nil {
		return nil, fmt.Errorf("変更詳細の取得に失敗しました: %w", err)
	}

	return change, nil
}

func (c *Client) ListChangeComments(ctx context.Context, changeID string) (map[string][]gerrit.CommentInfo, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	c.logger.Debug("変更コメントを取得しています", zap.String("change_id", changeID))

	comments, err := c.client.ListChangeComments(ctx, changeID)
	if err != nil {
		return nil, fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	return comments, nil
}

func (c *Client) ListFiles(ctx context.Context, changeID, revision string) (map[string]*gerrit.FileInfo, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	c.logger.Debug("ファイル一覧を取得しています",
		zap.String("change_id", changeID),
		zap.String("revision", revision))

	files, err := c.client.ListFiles(ctx, changeID, revision)
	if err != nil {
		return nil, fmt.Errorf("ファイル一覧の取得に失敗しました: %w", err)
	}

	return files, nil
}

func (c *Client) GetFileDiff(ctx context.Context, changeID, revision, filePath string) (string, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return "", fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	c.logger.Debug("ファイル差分を取得しています",
		zap.String("change_id", changeID),
		zap.String("revision", revision),
		zap.String("file_path", filePath))

	diffURL := fmt.Sprintf("%s/changes/%s/revisions/%s/files/%s/diff",
		c.baseURL, changeID, revision, url.PathEscape(filePath))

	c.logger.Debug("差分URLを構築しました", zap.String("url", diffURL))

	return "", nil
}

func (c *Client) GetProjectInfo(ctx context.Context, project string) (*gerrit.ProjectInfo, error) {
	if err := c.waitForRateLimit(ctx); err != nil {
		return nil, fmt.Errorf("レート制限の待機に失敗しました: %w", err)
	}

	info, err := c.client.GetProjectInfo(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("プロジェクト情報の取得に失敗しました: %w", err)
	}

	return &info, nil
}

func (c *Client) GetClient() *gerrit.Client {
	return c.client
}
