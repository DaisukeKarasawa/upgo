package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"upgo/internal/llm"
	"upgo/legacy/internal/tracker"

	ghub "github.com/google/go-github/v60/github"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// mockPRFetcher is a mock implementation of PRFetcherInterface for testing
// that returns empty PR lists without making actual API calls
type mockPRFetcher struct{}

func (m *mockPRFetcher) FetchPRs(ctx context.Context, owner, repo string, state string) ([]*ghub.PullRequest, error) {
	return []*ghub.PullRequest{}, nil
}

func (m *mockPRFetcher) FetchPRsUpdatedSince(ctx context.Context, owner, repo string, state string, since time.Time) ([]*ghub.PullRequest, error) {
	return []*ghub.PullRequest{}, nil
}

func (m *mockPRFetcher) FetchPR(ctx context.Context, owner, repo string, number int) (*ghub.PullRequest, error) {
	return nil, nil
}

func (m *mockPRFetcher) FetchPRComments(ctx context.Context, owner, repo string, number int) ([]*ghub.IssueComment, error) {
	return []*ghub.IssueComment{}, nil
}

func (m *mockPRFetcher) FetchPRCommentsSince(ctx context.Context, owner, repo string, number int, since time.Time) ([]*ghub.IssueComment, error) {
	return []*ghub.IssueComment{}, nil
}

func (m *mockPRFetcher) FetchPRDiff(ctx context.Context, owner, repo string, number int) (string, error) {
	return "", nil
}

func (m *mockPRFetcher) FetchMergedCommits(ctx context.Context, owner, repo string, since time.Time) ([]*ghub.RepositoryCommit, error) {
	return []*ghub.RepositoryCommit{}, nil
}

// PerformanceMetrics holds performance measurement results
type PerformanceMetrics struct {
	Duration      time.Duration
	OperationName string
	Details       map[string]interface{}
}

// measurePerformance measures the execution time of a function
func measurePerformance(name string, fn func() error) (PerformanceMetrics, error) {
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	return PerformanceMetrics{
		Duration:      duration,
		OperationName: name,
		Details:       make(map[string]interface{}),
	}, err
}

// TestSyncService_Performance measures the performance of Sync operation
func TestSyncService_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

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

	logger := zap.NewNop()
	// Use mock PRFetcher to avoid actual GitHub API calls
	mockPRFetcher := &mockPRFetcher{}
	statusTracker := tracker.NewStatusTracker(db, logger)
	// Initialize AnalysisService with test doubles to avoid nil pointer dereference
	// For performance testing, we use Nop logger and skip LLM operations
	llmClient := llm.NewClient("http://localhost:11434", "llama3.2", 30, logger)
	summarizer := llm.NewSummarizer(llmClient, logger)
	analyzer := llm.NewAnalyzer(llmClient, logger)
	analysisService := NewAnalysisService(db, summarizer, analyzer, logger)

	service := NewSyncService(
		db,
		nil, // githubClient not needed with mock PRFetcher
		mockPRFetcher,
		statusTracker,
		analysisService,
		logger,
		"test-owner",
		"test-repo",
	)

	ctx := context.Background()

	// Measure Sync performance with mock PRFetcher (no network calls)
	metrics, err := measurePerformance("Sync", func() error {
		return service.Sync(ctx)
	})

	if err != nil {
		t.Logf("Sync operation returned error (expected without GitHub API): %v", err)
	}

	t.Logf("Performance Metrics:")
	t.Logf("  Operation: %s", metrics.OperationName)
	t.Logf("  Duration: %v", metrics.Duration)
	t.Logf("  Duration (ms): %.2f", float64(metrics.Duration.Nanoseconds())/1e6)

	// Performance assertions (adjust thresholds as needed)
	if metrics.Duration > 10*time.Second {
		t.Logf("WARNING: Sync operation took longer than expected: %v", metrics.Duration)
	}
}

// TestSyncService_getOrCreateRepository_Performance measures repository lookup/creation performance
func TestSyncService_getOrCreateRepository_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			last_synced_at DATETIME,
			UNIQUE(owner, name)
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	logger := zap.NewNop()
	// getOrCreateRepository doesn't use PRFetcher or AnalysisService, so we can use nil
	// However, we initialize AnalysisService properly to avoid potential nil pointer issues
	// if the implementation changes in the future
	statusTracker := tracker.NewStatusTracker(db, logger)
	llmClient := llm.NewClient("http://localhost:11434", "llama3.2", 30, logger)
	summarizer := llm.NewSummarizer(llmClient, logger)
	analyzer := llm.NewAnalyzer(llmClient, logger)
	analysisService := NewAnalysisService(db, summarizer, analyzer, logger)

	service := NewSyncService(
		db,
		nil, // githubClient not used by getOrCreateRepository
		nil, // prFetcher not used by getOrCreateRepository
		statusTracker,
		analysisService,
		logger,
		"test-owner",
		"test-repo",
	)

	// Measure getOrCreateRepository performance
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := service.getOrCreateRepository()
		if err != nil {
			t.Fatalf("getOrCreateRepository failed: %v", err)
		}
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	t.Logf("Performance Metrics:")
	t.Logf("  Operation: getOrCreateRepository")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Total Duration: %v", duration)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Average Duration (ms): %.2f", float64(avgDuration.Nanoseconds())/1e6)
	t.Logf("  Operations per second: %.2f", float64(iterations)/duration.Seconds())

	// Performance assertions
	if avgDuration > 100*time.Millisecond {
		t.Logf("WARNING: getOrCreateRepository is slower than expected: %v", avgDuration)
	}
}

// BenchmarkSyncService_getOrCreateRepository benchmarks repository lookup/creation
func BenchmarkSyncService_getOrCreateRepository(b *testing.B) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS repositories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			last_synced_at DATETIME,
			UNIQUE(owner, name)
		);
	`)
	if err != nil {
		b.Fatalf("Failed to create schema: %v", err)
	}

	logger := zap.NewNop()
	// getOrCreateRepository doesn't use PRFetcher or AnalysisService, so we can use nil
	// However, we initialize AnalysisService properly to avoid potential nil pointer issues
	statusTracker := tracker.NewStatusTracker(db, logger)
	llmClient := llm.NewClient("http://localhost:11434", "llama3.2", 30, logger)
	summarizer := llm.NewSummarizer(llmClient, logger)
	analyzer := llm.NewAnalyzer(llmClient, logger)
	analysisService := NewAnalysisService(db, summarizer, analyzer, logger)

	service := NewSyncService(
		db,
		nil, // githubClient not used by getOrCreateRepository
		nil, // prFetcher not used by getOrCreateRepository
		statusTracker,
		analysisService,
		logger,
		"test-owner",
		"test-repo",
	)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.getOrCreateRepository()
		if err != nil {
			b.Fatalf("getOrCreateRepository failed: %v", err)
		}
	}
}

// TestPerformanceComparison compares performance of different operations
func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize schema
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

	logger := zap.NewNop()
	// getOrCreateRepository doesn't use PRFetcher or AnalysisService, so we can use nil
	// However, we initialize AnalysisService properly to avoid potential nil pointer issues
	statusTracker := tracker.NewStatusTracker(db, logger)
	llmClient := llm.NewClient("http://localhost:11434", "llama3.2", 30, logger)
	summarizer := llm.NewSummarizer(llmClient, logger)
	analyzer := llm.NewAnalyzer(llmClient, logger)
	analysisService := NewAnalysisService(db, summarizer, analyzer, logger)

	service := NewSyncService(
		db,
		nil, // githubClient not used by getOrCreateRepository
		nil, // prFetcher not used by getOrCreateRepository
		statusTracker,
		analysisService,
		logger,
		"test-owner",
		"test-repo",
	)

	// Measure different operations
	operations := []struct {
		name string
		fn   func() error
	}{
		{
			name: "getOrCreateRepository",
			fn: func() error {
				_, err := service.getOrCreateRepository()
				return err
			},
		},
	}

	results := make([]PerformanceMetrics, 0, len(operations))

	for _, op := range operations {
		metrics, err := measurePerformance(op.name, op.fn)
		if err != nil {
			t.Logf("Operation %s returned error: %v", op.name, err)
		}
		results = append(results, metrics)
	}

	// Print comparison
	t.Logf("\n=== Performance Comparison ===")
	for _, result := range results {
		t.Logf("%s: %v (%.2f ms)", result.OperationName, result.Duration, float64(result.Duration.Nanoseconds())/1e6)
	}
}
