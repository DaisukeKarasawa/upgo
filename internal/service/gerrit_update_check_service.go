package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"upgo/internal/gerrit"

	"go.uber.org/zap"
)

// GerritUpdateCheckService provides lightweight update checking for Gerrit Changes.
// It checks if there are updates available without fetching full data.
type GerritUpdateCheckService struct {
	db            *sql.DB
	gerritClient  *gerrit.Client
	changeFetcher *gerrit.ChangeFetcher
	logger        *zap.Logger
	project       string

	// Cache for dashboard check results
	dashboardCache struct {
		mu              sync.RWMutex
		result          *GerritDashboardUpdateResult
		lastChecked     time.Time
		cacheTTL        time.Duration
	}
	// Cache for Change check results (keyed by Change ID)
	changeCache struct {
		mu              sync.RWMutex
		results         map[int]*GerritChangeUpdateResult
		cacheTTL        time.Duration
	}
}

// GerritDashboardUpdateResult represents the result of dashboard update check.
type GerritDashboardUpdateResult struct {
	HasMissingRecentChanges bool      `json:"has_missing_recent_changes"`
	MissingCount             int       `json:"missing_count"`
	LastCheckedAt            time.Time `json:"last_checked_at"`
}

// GerritChangeUpdateResult represents the result of Change update check.
type GerritChangeUpdateResult struct {
	UpdatedSinceLastSync bool       `json:"updated_since_last_sync"`
	LastSyncedAt         *time.Time `json:"last_synced_at"`
	GerritUpdatedAt      *time.Time `json:"gerrit_updated_at"`
	LastCheckedAt        time.Time  `json:"last_checked_at"`
}

// NewGerritUpdateCheckService creates a new GerritUpdateCheckService instance.
func NewGerritUpdateCheckService(
	db *sql.DB,
	gerritClient *gerrit.Client,
	changeFetcher *gerrit.ChangeFetcher,
	logger *zap.Logger,
	project string,
) *GerritUpdateCheckService {
	return &GerritUpdateCheckService{
		db:            db,
		gerritClient:  gerritClient,
		changeFetcher: changeFetcher,
		logger:        logger,
		project:       project,
		dashboardCache: struct {
			mu              sync.RWMutex
			result          *GerritDashboardUpdateResult
			lastChecked     time.Time
			cacheTTL        time.Duration
		}{
			cacheTTL: 60 * time.Second, // 60秒キャッシュ
		},
		changeCache: struct {
			mu              sync.RWMutex
			results         map[int]*GerritChangeUpdateResult
			cacheTTL        time.Duration
		}{
			results:  make(map[int]*GerritChangeUpdateResult),
			cacheTTL: 60 * time.Second, // 60秒キャッシュ
		},
	}
}

// CheckDashboardUpdates checks if there are Changes created in the last month that don't exist in DB.
// Uses caching to avoid excessive Gerrit API calls.
func (s *GerritUpdateCheckService) CheckDashboardUpdates(ctx context.Context) (*GerritDashboardUpdateResult, error) {
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

// GetCachedDashboardResult returns the cached dashboard result if available.
// This is used as a fallback when the actual check fails.
func (s *GerritUpdateCheckService) GetCachedDashboardResult() *GerritDashboardUpdateResult {
	s.dashboardCache.mu.RLock()
	defer s.dashboardCache.mu.RUnlock()
	if s.dashboardCache.result != nil {
		result := *s.dashboardCache.result
		return &result
	}
	return nil
}

// checkDashboardUpdates performs the actual Gerrit API call to check for missing Changes.
func (s *GerritUpdateCheckService) checkDashboardUpdates(ctx context.Context) (*GerritDashboardUpdateResult, error) {
	monthAgo := time.Now().AddDate(0, 0, -30)

	// Fetch Changes updated since a month ago (lightweight to reduce latency)
	changes, err := s.changeFetcher.FetchChangesUpdatedSinceLight(ctx, monthAgo)
	if err != nil {
		return nil, fmt.Errorf("Change一覧取得失敗: %w", err)
	}

	if len(changes) == 0 {
		return &GerritDashboardUpdateResult{
			HasMissingRecentChanges: false,
			MissingCount:             0,
			LastCheckedAt:            time.Now(),
		}, nil
	}

	// Get repository ID
	var repoID int
	err = s.db.QueryRow(
		"SELECT id FROM repositories WHERE owner = ? AND name = ?",
		"go", s.project,
	).Scan(&repoID)
	if err == sql.ErrNoRows {
		// Repository doesn't exist in DB, so all Changes are missing
		return &GerritDashboardUpdateResult{
			HasMissingRecentChanges: len(changes) > 0,
			MissingCount:             len(changes),
			LastCheckedAt:            time.Now(),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("リポジトリIDの取得に失敗しました: %w", err)
	}

	// Collect Change numbers
	changeNumbers := make([]int, 0, len(changes))
	for _, change := range changes {
		changeNumbers = append(changeNumbers, change.Number)
	}

	// Check which Changes exist in DB
	placeholders := ""
	args := make([]interface{}, 0, len(changeNumbers)+1)
	args = append(args, repoID)
	for i, num := range changeNumbers {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "?"
		args = append(args, num)
	}

	query := fmt.Sprintf(
		"SELECT change_number FROM changes WHERE repository_id = ? AND change_number IN (%s)",
		placeholders,
	)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("Change存在確認のクエリに失敗しました: %w", err)
	}
	defer rows.Close()

	existingNumbers := make(map[int]bool)
	for rows.Next() {
		var changeNumber int
		if err := rows.Scan(&changeNumber); err != nil {
			s.logger.Warn("行のスキャンに失敗しました", zap.Error(err))
			continue
		}
		existingNumbers[changeNumber] = true
	}

	// Count missing Changes
	missingCount := 0
	for _, change := range changes {
		if !existingNumbers[change.Number] {
			missingCount++
		}
	}

	return &GerritDashboardUpdateResult{
		HasMissingRecentChanges: missingCount > 0,
		MissingCount:             missingCount,
		LastCheckedAt:            time.Now(),
	}, nil
}

// CheckChangeUpdates checks if a specific Change has been updated on Gerrit since its last_synced_at.
// Uses caching to avoid excessive Gerrit API calls.
func (s *GerritUpdateCheckService) CheckChangeUpdates(ctx context.Context, changeID int) (*GerritChangeUpdateResult, error) {
	// Check cache first
	s.changeCache.mu.RLock()
	if cached, ok := s.changeCache.results[changeID]; ok {
		if time.Since(cached.LastCheckedAt) < s.changeCache.cacheTTL {
			result := *cached
			s.changeCache.mu.RUnlock()
			return &result, nil
		}
	}
	s.changeCache.mu.RUnlock()

	// Cache miss or expired, perform actual check
	result, err := s.checkChangeUpdates(ctx, changeID)
	if err != nil {
		return nil, err
	}

	// Update cache
	s.changeCache.mu.Lock()
	s.changeCache.results[changeID] = result
	s.changeCache.mu.Unlock()

	return result, nil
}

// checkChangeUpdates performs the actual Gerrit API call to check if Change has been updated.
func (s *GerritUpdateCheckService) checkChangeUpdates(ctx context.Context, changeID int) (*GerritChangeUpdateResult, error) {
	// Get Change info from database
	var changeNumber int
	var changeGerritID string
	var lastSyncedAt sql.NullTime
	err := s.db.QueryRow(
		"SELECT change_number, change_id, last_synced_at FROM changes WHERE id = ?",
		changeID,
	).Scan(&changeNumber, &changeGerritID, &lastSyncedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Changeが見つかりません: change_id=%d", changeID)
	}
	if err != nil {
		return nil, fmt.Errorf("Change情報の取得に失敗しました: %w", err)
	}

	// Fetch Change from Gerrit (lightweight - only basic info)
	changeInfo, err := s.changeFetcher.FetchChangeDetail(ctx, changeGerritID)
	if err != nil {
		return nil, fmt.Errorf("GerritからのChange取得に失敗しました: %w", err)
	}

	gerritUpdatedAt, err := parseGerritTime(changeInfo.Updated)
	if err != nil {
		return nil, fmt.Errorf("更新時刻の解析に失敗しました: %w", err)
	}

	// If last_synced_at is NULL, treat as not synced yet (show update indicator)
	var updatedSinceLastSync bool
	if !lastSyncedAt.Valid {
		updatedSinceLastSync = true
	} else {
		updatedSinceLastSync = gerritUpdatedAt.After(lastSyncedAt.Time)
	}

	var lastSyncedAtPtr *time.Time
	if lastSyncedAt.Valid {
		lastSyncedAtPtr = &lastSyncedAt.Time
	}

	return &GerritChangeUpdateResult{
		UpdatedSinceLastSync: updatedSinceLastSync,
		LastSyncedAt:         lastSyncedAtPtr,
		GerritUpdatedAt:      &gerritUpdatedAt,
		LastCheckedAt:        time.Now(),
	}, nil
}
