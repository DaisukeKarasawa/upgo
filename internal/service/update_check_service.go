package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"upgo/internal/github"

	ghub "github.com/google/go-github/v60/github"
	"go.uber.org/zap"
)

// UpdateCheckService provides lightweight update checking without full sync.
// It checks if there are updates available without fetching/analyzing data.
type UpdateCheckService struct {
	db           *sql.DB
	githubClient *github.Client
	prFetcher    github.PRFetcherInterface
	logger       *zap.Logger
	owner        string
	repo         string

	// Cache for dashboard check results
	dashboardCache struct {
		mu              sync.RWMutex
		result          *DashboardUpdateResult
		lastChecked     time.Time
		cacheTTL        time.Duration
	}
	// Cache for PR check results (keyed by PR ID)
	prCache struct {
		mu              sync.RWMutex
		results         map[int]*PRUpdateResult
		cacheTTL        time.Duration
	}
}

// DashboardUpdateResult represents the result of dashboard update check.
type DashboardUpdateResult struct {
	HasMissingRecentPRs bool      `json:"has_missing_recent_prs"`
	MissingCount         int       `json:"missing_count"`
	LastCheckedAt        time.Time `json:"last_checked_at"`
}

// PRUpdateResult represents the result of PR update check.
type PRUpdateResult struct {
	UpdatedSinceLastSync bool       `json:"updated_since_last_sync"`
	LastSyncedAt          *time.Time `json:"last_synced_at"`
	GitHubUpdatedAt       *time.Time `json:"github_updated_at"`
	LastCheckedAt         time.Time  `json:"last_checked_at"`
}

// NewUpdateCheckService creates a new UpdateCheckService instance.
func NewUpdateCheckService(
	db *sql.DB,
	githubClient *github.Client,
	prFetcher github.PRFetcherInterface,
	logger *zap.Logger,
	owner, repo string,
) *UpdateCheckService {
	return &UpdateCheckService{
		db:           db,
		githubClient: githubClient,
		prFetcher:    prFetcher,
		logger:       logger,
		owner:        owner,
		repo:         repo,
		dashboardCache: struct {
			mu              sync.RWMutex
			result          *DashboardUpdateResult
			lastChecked     time.Time
			cacheTTL        time.Duration
		}{
			cacheTTL: 60 * time.Second, // 60秒キャッシュ
		},
		prCache: struct {
			mu              sync.RWMutex
			results         map[int]*PRUpdateResult
			cacheTTL        time.Duration
		}{
			results:  make(map[int]*PRUpdateResult),
			cacheTTL: 60 * time.Second, // 60秒キャッシュ
		},
	}
}

// CheckDashboardUpdates checks if there are PRs created in the last month that don't exist in DB.
// Uses caching to avoid excessive GitHub API calls.
func (s *UpdateCheckService) CheckDashboardUpdates(ctx context.Context) (*DashboardUpdateResult, error) {
	// Check cache first
	s.dashboardCache.mu.RLock()
	if s.dashboardCache.result != nil && time.Since(s.dashboardCache.lastChecked) < s.dashboardCache.cacheTTL {
		result := *s.dashboardCache.result
		s.dashboardCache.mu.RUnlock()
		return &result, nil
	}
	s.dashboardCache.mu.RUnlock()

	// Cache miss or expired, perform actual check
	result, err := s.checkDashboardUpdates(ctx)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.dashboardCache.mu.Lock()
	s.dashboardCache.result = result
	s.dashboardCache.lastChecked = time.Now()
	s.dashboardCache.mu.Unlock()

	return result, nil
}

// checkDashboardUpdates performs the actual GitHub API call to check for missing PRs.
func (s *UpdateCheckService) checkDashboardUpdates(ctx context.Context) (*DashboardUpdateResult, error) {
	monthAgo := time.Now().AddDate(0, 0, -30)

	// Fetch PRs updated since a month ago (we'll filter by created_at later)
	// Use parallel fetching for open and closed states
	states := []string{"open", "closed"}
	var allRecentPRs []*ghub.PullRequest
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 2)

	for _, state := range states {
		wg.Add(1)
		go func(state string) {
			defer wg.Done()
			prs, err := s.prFetcher.FetchPRsUpdatedSince(ctx, s.owner, s.repo, state, monthAgo)
			if err != nil {
				errChan <- fmt.Errorf("PR一覧取得失敗 (%s): %w", state, err)
				return
			}
			mu.Lock()
			allRecentPRs = append(allRecentPRs, prs...)
			mu.Unlock()
		}(state)
	}

	wg.Wait()
	close(errChan)

	// Check for errors (drain buffered errors after close).
	// NOTE: Receiving from a closed channel yields a zero value immediately,
	// so we must not treat "channel closed" as an error.
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// Filter PRs created in the last month
	var recentCreatedPRs []*ghub.PullRequest
	for _, pr := range allRecentPRs {
		if pr.GetCreatedAt().Time.After(monthAgo) || pr.GetCreatedAt().Time.Equal(monthAgo) {
			recentCreatedPRs = append(recentCreatedPRs, pr)
		}
	}

	if len(recentCreatedPRs) == 0 {
		return &DashboardUpdateResult{
			HasMissingRecentPRs: false,
			MissingCount:         0,
			LastCheckedAt:        time.Now(),
		}, nil
	}

	// Get repository ID
	var repoID int
	err := s.db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		s.owner, s.repo,
	).Scan(&repoID)
	if err == sql.ErrNoRows {
		// Repository doesn't exist in DB, so all PRs are missing
		return &DashboardUpdateResult{
			HasMissingRecentPRs: len(recentCreatedPRs) > 0,
			MissingCount:         len(recentCreatedPRs),
			LastCheckedAt:        time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("リポジトリIDの取得に失敗しました: %w", err)
	}

	// Collect GitHub IDs (PR numbers)
	githubIDs := make([]int, 0, len(recentCreatedPRs))
	for _, pr := range recentCreatedPRs {
		githubIDs = append(githubIDs, pr.GetNumber())
	}

	// Check which PRs exist in DB
	// Build IN clause query
	placeholders := ""
	args := make([]interface{}, 0, len(githubIDs)+1)
	args = append(args, repoID)
	for i, id := range githubIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		"SELECT github_id FROM pull_requests WHERE repository_id = ? AND github_id IN (%s)",
		placeholders,
	)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("PR存在確認のクエリに失敗しました: %w", err)
	}
	defer rows.Close()

	existingIDs := make(map[int]bool)
	for rows.Next() {
		var githubID int
		if err := rows.Scan(&githubID); err != nil {
			s.logger.Warn("行のスキャンに失敗しました", zap.Error(err))
			continue
		}
		existingIDs[githubID] = true
	}

	// Count missing PRs
	missingCount := 0
	for _, pr := range recentCreatedPRs {
		if !existingIDs[pr.GetNumber()] {
			missingCount++
		}
	}

	return &DashboardUpdateResult{
		HasMissingRecentPRs: missingCount > 0,
		MissingCount:         missingCount,
		LastCheckedAt:        time.Now(),
	}, nil
}

// CheckPRUpdates checks if a specific PR has been updated on GitHub since its last_synced_at.
// Uses caching to avoid excessive GitHub API calls.
func (s *UpdateCheckService) CheckPRUpdates(ctx context.Context, prID int) (*PRUpdateResult, error) {
	// #region agent log
	if f, ferr := os.OpenFile("/Users/daisuke/dev/upgo/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
		entry := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "pre-fix",
			"hypothesisId": "H1",
			"location":     "service/update_check_service.go:252",
			"message":      "CheckPRUpdates entry",
			"data": map[string]interface{}{
				"prID": prID,
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if b, merr := json.Marshal(entry); merr == nil {
			_, _ = f.Write(append(b, '\n'))
		}
		_ = f.Close()
	}
	// #endregion agent log

	// Check cache first
	s.prCache.mu.RLock()
	if cached, ok := s.prCache.results[prID]; ok {
		if time.Since(cached.LastCheckedAt) < s.prCache.cacheTTL {
			result := *cached
			s.prCache.mu.RUnlock()

			// #region agent log
			if f, ferr := os.OpenFile("/Users/daisuke/dev/upgo/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
				entry := map[string]interface{}{
					"sessionId":    "debug-session",
					"runId":        "pre-fix",
					"hypothesisId": "H2",
					"location":     "service/update_check_service.go:270",
					"message":      "CheckPRUpdates cache hit",
					"data": map[string]interface{}{
						"prID":          prID,
						"lastCheckedAt": cached.LastCheckedAt.Format(time.RFC3339Nano),
					},
					"timestamp": time.Now().UnixMilli(),
				}
				if b, merr := json.Marshal(entry); merr == nil {
					_, _ = f.Write(append(b, '\n'))
				}
				_ = f.Close()
			}
			// #endregion agent log

			return &result, nil
		}
	}
	s.prCache.mu.RUnlock()

	// Cache miss or expired, perform actual check
	result, err := s.checkPRUpdates(ctx, prID)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.prCache.mu.Lock()
	s.prCache.results[prID] = result
	s.prCache.mu.Unlock()

	return result, nil
}

// checkPRUpdates performs the actual GitHub API call to check if PR has been updated.
func (s *UpdateCheckService) checkPRUpdates(ctx context.Context, prID int) (*PRUpdateResult, error) {
	// Get PR info from database
	var githubID int
	var lastSyncedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT github_id, last_synced_at FROM pull_requests WHERE id = ?",
		prID,
	).Scan(&githubID, &lastSyncedAt)

	// #region agent log
	if f, ferr := os.OpenFile("/Users/daisuke/dev/upgo/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
		entry := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "pre-fix",
			"hypothesisId": "H3",
			"location":     "service/update_check_service.go:294",
			"message":      "DB lookup result",
			"data": map[string]interface{}{
				"prID":        prID,
				"githubID":    githubID,
				"lastSynced":  lastSyncedAt.Time.Format(time.RFC3339Nano),
				"lastSyncedValid": lastSyncedAt.Valid,
				"err":         fmt.Sprint(err),
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if b, merr := json.Marshal(entry); merr == nil {
			_, _ = f.Write(append(b, '\n'))
		}
		_ = f.Close()
	}
	// #endregion agent log

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("PRが見つかりません: pr_id=%d", prID)
	}
	if err != nil {
		return nil, fmt.Errorf("PR情報の取得に失敗しました: %w", err)
	}

	// Fetch PR from GitHub
	pr, err := s.prFetcher.FetchPR(ctx, s.owner, s.repo, githubID)
	if err != nil {
		return nil, fmt.Errorf("GitHubからのPR取得に失敗しました: %w", err)
	}

	githubUpdatedAt := pr.GetUpdatedAt().Time

	// If last_synced_at is NULL, treat as not synced yet (show update indicator)
	var updatedSinceLastSync bool
	if !lastSyncedAt.Valid {
		updatedSinceLastSync = true
	} else {
		updatedSinceLastSync = githubUpdatedAt.After(lastSyncedAt.Time)
	}

	var lastSyncedAtPtr *time.Time
	if lastSyncedAt.Valid {
		lastSyncedAtPtr = &lastSyncedAt.Time
	}

	// #region agent log
	if f, ferr := os.OpenFile("/Users/daisuke/dev/upgo/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); ferr == nil {
		entry := map[string]interface{}{
			"sessionId":    "debug-session",
			"runId":        "pre-fix",
			"hypothesisId": "H4",
			"location":     "service/update_check_service.go:332",
			"message":      "GitHub comparison",
			"data": map[string]interface{}{
				"prID":                prID,
				"githubID":            githubID,
				"githubUpdatedAt":     githubUpdatedAt.Format(time.RFC3339Nano),
				"lastSyncedValid":     lastSyncedAt.Valid,
				"lastSyncedAt":        lastSyncedAt.Time.Format(time.RFC3339Nano),
				"updatedSinceLastSync": updatedSinceLastSync,
			},
			"timestamp": time.Now().UnixMilli(),
		}
		if b, merr := json.Marshal(entry); merr == nil {
			_, _ = f.Write(append(b, '\n'))
		}
		_ = f.Close()
	}
	// #endregion agent log

	return &PRUpdateResult{
		UpdatedSinceLastSync: updatedSinceLastSync,
		LastSyncedAt:         lastSyncedAtPtr,
		GitHubUpdatedAt:      &githubUpdatedAt,
		LastCheckedAt:        time.Now(),
	}, nil
}
