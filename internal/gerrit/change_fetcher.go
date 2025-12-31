package gerrit

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ChangeFetcher fetches changes from Gerrit
type ChangeFetcher struct {
	client    *Client
	logger    *zap.Logger
	project   string
	branches  []string
	statuses  []string
	days      int
}

// Days returns the configured days for fetching changes
func (f *ChangeFetcher) Days() int {
	return f.days
}

// NewChangeFetcher creates a new change fetcher
func NewChangeFetcher(client *Client, logger *zap.Logger, project string, branches []string, statuses []string, days int) *ChangeFetcher {
	return &ChangeFetcher{
		client:   client,
		logger:   logger,
		project:  project,
		branches: branches,
		statuses: statuses,
		days:     days,
	}
}

// FetchChangesUpdatedSince fetches changes updated since the given time
func (f *ChangeFetcher) FetchChangesUpdatedSince(ctx context.Context, sinceTime time.Time) ([]ChangeInfo, error) {
	return f.fetchChangesWithOptions(ctx, sinceTime, nil, 100)
}

// FetchChangesUpdatedSinceLight fetches changes with minimal options for lightweight update checks.
// It uses fewer Gerrit query options to reduce response size/latency.
func (f *ChangeFetcher) FetchChangesUpdatedSinceLight(ctx context.Context, sinceTime time.Time) ([]ChangeInfo, error) {
	// Use minimal options (no detailed fields) and smaller page size for faster response
	// Only fetch basic change info needed for update checking
	const lightLimit = 25 // Reduced from 50 for even faster response
	return f.fetchChangesWithOptions(ctx, sinceTime, nil, lightLimit)
}

// fetchChangesForStatus fetches changes for a specific status
func (f *ChangeFetcher) fetchChangesForStatus(ctx context.Context, status string, sinceTime time.Time) ([]ChangeInfo, error) {
	defaultOptions := []string{
		"CURRENT_REVISION",
		"CURRENT_FILES",
		"LABELS",
		"DETAILED_LABELS",
	}
	const defaultLimit = 100
	return f.fetchChangesForStatusWithOptions(ctx, status, sinceTime, defaultOptions, defaultLimit)
}

// fetchChangesWithOptions iterates over all statuses using the provided options/limit.
// Used by both full sync and lightweight update checks.
func (f *ChangeFetcher) fetchChangesWithOptions(ctx context.Context, sinceTime time.Time, options []string, limit int) ([]ChangeInfo, error) {
	var allChanges []ChangeInfo

	// If the caller context is already canceled, fail fast.
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var (
		anySuccess bool
		lastErr    error
	)

	for _, status := range f.statuses {
		changes, err := f.fetchChangesForStatusWithOptions(ctx, status, sinceTime, options, limit)
		if err != nil {
			// If context was canceled (e.g., client disconnected), do not keep trying.
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return nil, err
			}

			// Non-context error: keep going to allow partial results, but remember last error.
			lastErr = err
			f.logger.Warn("ステータス別の取得に失敗しました", zap.String("status", status), zap.Error(err))
			continue
		}
		anySuccess = true
		allChanges = append(allChanges, changes...)
	}

	// If nothing succeeded and we have an error, bubble it up (sync should fail visibly).
	if !anySuccess && lastErr != nil {
		return nil, lastErr
	}

	return allChanges, nil
}

// fetchChangesForStatusWithOptions fetches changes for a specific status with custom options/limit.
func (f *ChangeFetcher) fetchChangesForStatusWithOptions(ctx context.Context, status string, sinceTime time.Time, options []string, limit int) ([]ChangeInfo, error) {
	var allChanges []ChangeInfo
	start := 0
	if limit <= 0 {
		limit = 100
	}

	for {
		query := f.buildQuery(status, sinceTime)
		f.logger.Info("Gerritクエリを実行", zap.String("query", query), zap.Int("start", start), zap.Int("limit", limit))

		changes, err := f.client.QueryChanges(ctx, query, limit, start, options)
		if err != nil {
			return nil, fmt.Errorf("変更の取得に失敗しました: %w", err)
		}

		f.logger.Info("Gerritクエリ結果", zap.Int("fetched_count", len(changes)), zap.Int("start", start))

		if len(changes) == 0 {
			f.logger.Info("取得結果が0件のため終了", zap.String("status", status))
			break
		}

		// Apply branch filtering locally (Gerrit query doesn't support our wildcard/regex patterns reliably).
		matchedCount := 0
		for i := range changes {
			if f.MatchBranch(changes[i].Branch) {
				allChanges = append(allChanges, changes[i])
				matchedCount++
			}
		}
		f.logger.Info("ブランチフィルタリング結果", zap.Int("fetched", len(changes)), zap.Int("matched", matchedCount), zap.Int("total_accumulated", len(allChanges)))

		// If the page is full, fetch next page; otherwise stop.
		// Note: We check the fetched count, not the matched count, because Gerrit may have more pages
		// even if our branch filter excluded some results.
		if len(changes) < limit {
			f.logger.Info("ページが満杯でないため終了", zap.Int("fetched", len(changes)), zap.Int("limit", limit))
			break
		}
		start += limit
		f.logger.Info("次のページを取得", zap.Int("next_start", start))
	}

	return allChanges, nil
}

// buildQuery builds a Gerrit query string
func (f *ChangeFetcher) buildQuery(status string, sinceTime time.Time) string {
	var parts []string

	// Project/Repo: Use both for compatibility
	// repo: is more commonly used in Gerrit queries
	parts = append(parts, fmt.Sprintf("repo:%s", f.project))
	parts = append(parts, fmt.Sprintf("project:%s", f.project))

	// Status
	if status != "" {
		parts = append(parts, fmt.Sprintf("status:%s", status))
	}

	// Exclude WIP (Work In Progress) changes
	// This matches the expected query format: status:open -is:wip repo:go
	parts = append(parts, "-is:wip")

	// Updated since
	//
	// IMPORTANT:
	// - Gerrit query language supports time filtering via "after:" (alias: since:) / "before:".
	// - Gerrit's "after:" operator expects date format YYYY-MM-DD (time is optional but may cause issues).
	// - Use date-only format for better compatibility with Gerrit's query parser.
	parts = append(parts, fmt.Sprintf("after:%s", formatGerritDate(sinceTime)))

	// IMPORTANT: build a raw Gerrit query with spaces. URL encoding happens in QueryChanges.
	return strings.Join(parts, " ")
}

// formatGerritTime formats time for Gerrit query (legacy, kept for compatibility)
func formatGerritTime(t time.Time) string {
	// Gerrit expects format: YYYY-MM-DD[ HH:MM:SS[.SSS][ -TZ]]
	// If timezone is omitted, Gerrit defaults to UTC.
	return t.UTC().Format("2006-01-02 15:04:05")
}

// formatGerritDate formats date for Gerrit query (date-only format)
func formatGerritDate(t time.Time) string {
	// Gerrit's "after:" operator expects YYYY-MM-DD format.
	// Using date-only format avoids parsing issues with spaces/quotes.
	return t.UTC().Format("2006-01-02")
}

// hasMoreChanges checks if there are more changes to fetch
func hasMoreChanges(changes []ChangeInfo) bool {
	// Gerrit may set _more_changes field, but it's not in our ChangeInfo struct
	// For now, we'll rely on the count check
	return false // Will be determined by count < limit
}

// FetchChangeDetail fetches detailed information for a change
func (f *ChangeFetcher) FetchChangeDetail(ctx context.Context, changeID string) (*ChangeInfo, error) {
	options := []string{
		"CURRENT_REVISION",
		"ALL_REVISIONS",
		"CURRENT_FILES",
		"ALL_FILES",
		"DETAILED_LABELS",
		"DETAILED_ACCOUNTS",
		"MESSAGES",
		"CURRENT_COMMIT",
		"ALL_COMMITS",
	}

	change, err := f.client.GetChange(ctx, changeID, options)
	if err != nil {
		return nil, fmt.Errorf("変更詳細の取得に失敗しました: %w", err)
	}

	return change, nil
}

// FetchChangeComments fetches all comments for a change
func (f *ChangeFetcher) FetchChangeComments(ctx context.Context, changeID string) (map[string][]CommentInfo, error) {
	comments, err := f.client.GetChangeComments(ctx, changeID)
	if err != nil {
		return nil, fmt.Errorf("コメントの取得に失敗しました: %w", err)
	}

	return comments, nil
}

// FetchRevisionComments fetches comments for a specific revision
func (f *ChangeFetcher) FetchRevisionComments(ctx context.Context, changeID, revisionID string) (map[string][]CommentInfo, error) {
	comments, err := f.client.GetRevisionComments(ctx, changeID, revisionID)
	if err != nil {
		return nil, fmt.Errorf("リビジョンコメントの取得に失敗しました: %w", err)
	}

	return comments, nil
}

// FetchFileDiff fetches diff for a specific file
func (f *ChangeFetcher) FetchFileDiff(ctx context.Context, changeID, revisionID, filePath string) (*DiffInfo, error) {
	diff, err := f.client.GetFileDiff(ctx, changeID, revisionID, filePath)
	if err != nil {
		return nil, fmt.Errorf("ファイル差分の取得に失敗しました: %w", err)
	}

	return diff, nil
}

// MatchBranch checks if a branch matches any of the configured branch patterns
func (f *ChangeFetcher) MatchBranch(branch string) bool {
	for _, pattern := range f.branches {
		// Convert Gerrit branch pattern to regex
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		matched, err := regexp.MatchString("^"+regexPattern+"$", branch)
		if err == nil && matched {
			return true
		}
		// Also check exact match
		if pattern == branch {
			return true
		}
	}
	return false
}
