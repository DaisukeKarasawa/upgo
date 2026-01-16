package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	ghub "github.com/google/go-github/v60/github"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// testPRFetcher is a mock implementation of PRFetcherInterface for testing
// that allows customizing return values
type testPRFetcher struct {
	fetchPRsUpdatedSinceFunc func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error)
	fetchPRFunc              func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error)
}

func (m *testPRFetcher) FetchPRs(ctx context.Context, owner, repo string, state string) ([]*ghub.PullRequest, error) {
	return []*ghub.PullRequest{}, nil
}

func (m *testPRFetcher) FetchPRsUpdatedSince(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
	if m.fetchPRsUpdatedSinceFunc != nil {
		return m.fetchPRsUpdatedSinceFunc(ctx, owner, repo, state, since)
	}
	return []*ghub.PullRequest{}, nil
}

func (m *testPRFetcher) FetchPR(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
	if m.fetchPRFunc != nil {
		return m.fetchPRFunc(ctx, owner, repo, number)
	}
	return nil, nil
}

func (m *testPRFetcher) FetchPRComments(ctx context.Context, owner, repo string, number int) ([]*ghub.IssueComment, error) {
	return []*ghub.IssueComment{}, nil
}

func (m *testPRFetcher) FetchPRCommentsSince(ctx context.Context, owner, repo string, number int, since time.Time) ([]*ghub.IssueComment, error) {
	return []*ghub.IssueComment{}, nil
}

func (m *testPRFetcher) FetchPRDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	return "", nil
}

func (m *testPRFetcher) FetchMergedCommits(ctx context.Context, owner, repo string, since time.Time) ([]*ghub.RepositoryCommit, error) {
	return []*ghub.RepositoryCommit{}, nil
}

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Initialize database schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			last_synced_at DATETIME,
			UNIQUE(owner, name)
		);
		CREATE TABLE IF NOT EXISTS pull_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			repository_id INTEGER NOT NULL,
			github_id INTEGER NOT NULL,
			title TEXT,
			body TEXT,
			state TEXT,
			author TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			merged_at DATETIME,
			closed_at DATETIME,
			url TEXT,
			last_synced_at DATETIME,
			head_sha TEXT,
			UNIQUE(repository_id, github_id)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return db
}

func TestUpdateCheckService_CheckDashboardUpdates(t *testing.T) {
	tests := []struct {
		name           string
		setupDB        func(*sql.DB) error
		mockPRFetcher  *testPRFetcher
		wantHasMissing bool
		wantCount      int
		wantError      bool
	}{
		{
			name: "no recent PRs",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo')
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
					return []*ghub.PullRequest{}, nil
				},
			},
			wantHasMissing: false,
			wantCount:      0,
			wantError:      false,
		},
		{
			name: "recent PRs exist but all in DB",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (repository_id, github_id, title) VALUES (1, 100, 'PR 100');
					INSERT INTO pull_requests (repository_id, github_id, title) VALUES (1, 101, 'PR 101');
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
					now := time.Now()
					return []*ghub.PullRequest{
						{Number: intPtr(100), CreatedAt: &ghub.Timestamp{Time: now.Add(-10 * 24 * time.Hour)}},
						{Number: intPtr(101), CreatedAt: &ghub.Timestamp{Time: now.Add(-5 * 24 * time.Hour)}},
					}, nil
				},
			},
			wantHasMissing: false,
			wantCount:      0,
			wantError:      false,
		},
		{
			name: "recent PRs exist but some missing in DB",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (repository_id, github_id, title) VALUES (1, 100, 'PR 100');
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
					now := time.Now()
					// Return different PRs for open and closed states to avoid duplicates
					if state == "open" {
						return []*ghub.PullRequest{
							{Number: intPtr(100), CreatedAt: &ghub.Timestamp{Time: now.Add(-10 * 24 * time.Hour)}},
							{Number: intPtr(101), CreatedAt: &ghub.Timestamp{Time: now.Add(-5 * 24 * time.Hour)}},
						}, nil
					}
					// closed state
					return []*ghub.PullRequest{
						{Number: intPtr(102), CreatedAt: &ghub.Timestamp{Time: now.Add(-3 * 24 * time.Hour)}},
					}, nil
				},
			},
			wantHasMissing: true,
			wantCount:      2, // PRs 101 and 102 are missing
			wantError:      false,
		},
		{
			name: "repository not in DB",
			setupDB: func(db *sql.DB) error {
				// Don't insert repository
				return nil
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
					now := time.Now()
					// Return different PRs for open and closed states
					if state == "open" {
						return []*ghub.PullRequest{
							{Number: intPtr(100), CreatedAt: &ghub.Timestamp{Time: now.Add(-10 * 24 * time.Hour)}},
						}, nil
					}
					// closed state - return empty to avoid duplicates
					return []*ghub.PullRequest{}, nil
				},
			},
			wantHasMissing: true,
			wantCount:      1, // Only PR 100 from open state
			wantError:      false,
		},
		{
			name: "GitHub API error",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo')
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
					return nil, fmt.Errorf("API error")
				},
			},
			wantHasMissing: false,
			wantCount:      0,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			if tt.setupDB != nil {
				if err := tt.setupDB(db); err != nil {
					t.Fatalf("Setup DB failed: %v", err)
				}
			}

			logger := zap.NewNop()
			service := NewUpdateCheckService(
				db,
				nil, // githubClient not needed with mock PRFetcher
				tt.mockPRFetcher,
				logger,
				"test-owner",
				"test-repo",
			)

			ctx := context.Background()
			result, err := service.CheckDashboardUpdates(ctx)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CheckDashboardUpdates() error = %v, want no error", err)
			}

			if result.HasMissingRecentPRs != tt.wantHasMissing {
				t.Errorf("HasMissingRecentPRs = %v, want %v", result.HasMissingRecentPRs, tt.wantHasMissing)
			}

			if result.MissingCount != tt.wantCount {
				t.Errorf("MissingCount = %d, want %d", result.MissingCount, tt.wantCount)
			}
		})
	}
}

func TestUpdateCheckService_CheckDashboardUpdates_Cache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo')
	`)
	if err != nil {
		t.Fatalf("Setup DB failed: %v", err)
	}

	callCount := 0
	mockPRFetcher := &testPRFetcher{
		fetchPRsUpdatedSinceFunc: func(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
			callCount++
			return []*ghub.PullRequest{}, nil
		},
	}

	logger := zap.NewNop()
	service := NewUpdateCheckService(
		db,
		nil,
		mockPRFetcher,
		logger,
		"test-owner",
		"test-repo",
	)

	ctx := context.Background()

	// First call should hit GitHub API
	_, err = service.CheckDashboardUpdates(ctx)
	if err != nil {
		t.Fatalf("CheckDashboardUpdates() error = %v", err)
	}
	if callCount == 0 {
		t.Error("Expected GitHub API to be called on first request")
	}

	// Second call within cache TTL should use cache
	firstCallCount := callCount
	_, err = service.CheckDashboardUpdates(ctx)
	if err != nil {
		t.Fatalf("CheckDashboardUpdates() error = %v", err)
	}
	if callCount != firstCallCount {
		t.Errorf("Expected cache hit but GitHub API was called. Call count: %d -> %d", firstCallCount, callCount)
	}
}

func TestUpdateCheckService_CheckPRUpdates(t *testing.T) {
	tests := []struct {
		name                string
		setupDB             func(*sql.DB) error
		mockPRFetcher       *testPRFetcher
		prID                int
		wantUpdated         bool
		wantLastSyncedAtNil bool
		wantError           bool
	}{
		{
			name: "PR not found in DB",
			setupDB: func(db *sql.DB) error {
				return nil
			},
			mockPRFetcher: &testPRFetcher{},
			prID:          1,
			wantError:      true,
		},
		{
			name: "PR exists, last_synced_at is NULL",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (id, repository_id, github_id, title, last_synced_at) VALUES (1, 1, 100, 'PR 100', NULL);
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRFunc: func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
					now := time.Now()
					return &ghub.PullRequest{
						Number:    intPtr(100),
						UpdatedAt: &ghub.Timestamp{Time: now},
					}, nil
				},
			},
			prID:                1,
			wantUpdated:         true,
			wantLastSyncedAtNil: true,
			wantError:           false,
		},
		{
			name: "PR exists, updated since last sync",
			setupDB: func(db *sql.DB) error {
				now := time.Now()
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (id, repository_id, github_id, title, last_synced_at) VALUES (1, 1, 100, 'PR 100', ?);
				`, now.Add(-24*time.Hour))
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRFunc: func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
					now := time.Now()
					return &ghub.PullRequest{
						Number:    intPtr(100),
						UpdatedAt: &ghub.Timestamp{Time: now},
					}, nil
				},
			},
			prID:                1,
			wantUpdated:         true,
			wantLastSyncedAtNil: false,
			wantError:           false,
		},
		{
			name: "PR exists, not updated since last sync",
			setupDB: func(db *sql.DB) error {
				now := time.Now()
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (id, repository_id, github_id, title, last_synced_at) VALUES (1, 1, 100, 'PR 100', ?);
				`, now.Add(-1*time.Hour))
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRFunc: func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
					now := time.Now().Add(-2 * time.Hour)
					return &ghub.PullRequest{
						Number:    intPtr(100),
						UpdatedAt: &ghub.Timestamp{Time: now},
					}, nil
				},
			},
			prID:                1,
			wantUpdated:         false,
			wantLastSyncedAtNil: false,
			wantError:           false,
		},
		{
			name: "GitHub API error",
			setupDB: func(db *sql.DB) error {
				_, err := db.Exec(`
					INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
					INSERT INTO pull_requests (id, repository_id, github_id, title) VALUES (1, 1, 100, 'PR 100');
				`)
				return err
			},
			mockPRFetcher: &testPRFetcher{
				fetchPRFunc: func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
					return nil, fmt.Errorf("API error")
				},
			},
			prID:      1,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			if tt.setupDB != nil {
				if err := tt.setupDB(db); err != nil {
					t.Fatalf("Setup DB failed: %v", err)
				}
			}

			logger := zap.NewNop()
			service := NewUpdateCheckService(
				db,
				nil,
				tt.mockPRFetcher,
				logger,
				"test-owner",
				"test-repo",
			)

			ctx := context.Background()
			result, err := service.CheckPRUpdates(ctx, tt.prID)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("CheckPRUpdates() error = %v, want no error", err)
			}

			if result.UpdatedSinceLastSync != tt.wantUpdated {
				t.Errorf("UpdatedSinceLastSync = %v, want %v", result.UpdatedSinceLastSync, tt.wantUpdated)
			}

			if tt.wantLastSyncedAtNil {
				if result.LastSyncedAt != nil {
					t.Errorf("LastSyncedAt = %v, want nil", result.LastSyncedAt)
				}
			} else {
				if result.LastSyncedAt == nil {
					t.Error("LastSyncedAt = nil, want non-nil")
				}
			}

			if result.GitHubUpdatedAt == nil {
				t.Error("GitHubUpdatedAt = nil, want non-nil")
			}
		})
	}
}

func TestUpdateCheckService_CheckPRUpdates_Cache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	now := time.Now()
	_, err := db.Exec(`
		INSERT INTO repositories (id, owner, name) VALUES (1, 'test-owner', 'test-repo');
		INSERT INTO pull_requests (id, repository_id, github_id, title, last_synced_at) VALUES (1, 1, 100, 'PR 100', ?);
	`, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("Setup DB failed: %v", err)
	}

	callCount := 0
	mockPRFetcher := &testPRFetcher{
		fetchPRFunc: func(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
			callCount++
			return &ghub.PullRequest{
				Number:    intPtr(100),
				UpdatedAt: &ghub.Timestamp{Time: time.Now()},
			}, nil
		},
	}

	logger := zap.NewNop()
	service := NewUpdateCheckService(
		db,
		nil,
		mockPRFetcher,
		logger,
		"test-owner",
		"test-repo",
	)

	ctx := context.Background()

	// First call should hit GitHub API
	_, err = service.CheckPRUpdates(ctx, 1)
	if err != nil {
		t.Fatalf("CheckPRUpdates() error = %v", err)
	}
	if callCount == 0 {
		t.Error("Expected GitHub API to be called on first request")
	}

	// Second call within cache TTL should use cache
	firstCallCount := callCount
	_, err = service.CheckPRUpdates(ctx, 1)
	if err != nil {
		t.Fatalf("CheckPRUpdates() error = %v", err)
	}
	if callCount != firstCallCount {
		t.Errorf("Expected cache hit but GitHub API was called. Call count: %d -> %d", firstCallCount, callCount)
	}
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
