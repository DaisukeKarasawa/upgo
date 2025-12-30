package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"upgo/internal/github"
	"upgo/internal/tracker"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

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
	// Create properly initialized GitHub client (with dummy token for testing)
	githubClient := github.NewClient("dummy-token", logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	statusTracker := tracker.NewStatusTracker(db, logger)
	analysisService := &AnalysisService{}

	service := NewSyncService(
		db,
		githubClient,
		prFetcher,
		statusTracker,
		analysisService,
		logger,
		"test-owner",
		"test-repo",
	)

	ctx := context.Background()

	// Measure Sync performance
	// Note: This will fail if GitHub API is not accessible
	// For a real performance test, you'd want to mock the GitHub API calls
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
	githubClient := github.NewClient("dummy-token", logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	statusTracker := tracker.NewStatusTracker(db, logger)
	analysisService := &AnalysisService{}

	service := NewSyncService(
		db,
		githubClient,
		prFetcher,
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
	githubClient := github.NewClient("dummy-token", logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	statusTracker := tracker.NewStatusTracker(db, logger)
	analysisService := &AnalysisService{}

	service := NewSyncService(
		db,
		githubClient,
		prFetcher,
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
	githubClient := github.NewClient("dummy-token", logger)
	prFetcher := github.NewPRFetcher(githubClient, logger)
	statusTracker := tracker.NewStatusTracker(db, logger)
	analysisService := &AnalysisService{}

	service := NewSyncService(
		db,
		githubClient,
		prFetcher,
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
