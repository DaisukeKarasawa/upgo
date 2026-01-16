package service

import (
	"context"
	"database/sql"
	"testing"

	"upgo/internal/llm"
	"upgo/legacy/internal/tracker"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// BenchmarkSyncService_Sync benchmarks the Sync operation
// This measures the time taken to synchronize PRs from GitHub
func BenchmarkSyncService_Sync(b *testing.B) {
	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize database schema (simplified for benchmark)
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
		b.Fatalf("Failed to create schema: %v", err)
	}

	logger := zap.NewNop()
	// Use mock PRFetcher to avoid actual GitHub API calls
	mockPRFetcher := &mockPRFetcher{}
	statusTracker := tracker.NewStatusTracker(db, logger)
	// Initialize AnalysisService with test doubles to avoid nil pointer dereference
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

	// Benchmark Sync with mock PRFetcher (no network calls)
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = service.Sync(ctx)
	}
}

// BenchmarkSyncService_savePR benchmarks saving a single PR
func BenchmarkSyncService_savePR(b *testing.B) {
	// Setup similar to above
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
		b.Fatalf("Failed to create schema: %v", err)
	}

	// Insert test repository
	_, err = db.Exec("INSERT INTO repositories (owner, name) VALUES (?, ?)", "test-owner", "test-repo")
	if err != nil {
		b.Fatalf("Failed to insert repository: %v", err)
	}

	logger := zap.NewNop()
	// Use mock PRFetcher to avoid actual GitHub API calls
	mockPRFetcher := &mockPRFetcher{}
	statusTracker := tracker.NewStatusTracker(db, logger)
	// Initialize AnalysisService with test doubles to avoid nil pointer dereference
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

	// Note: This benchmark requires a mock PR object
	// For a real benchmark, you'd need to create a mock github.PullRequest
	b.Skip("Skipping savePR benchmark - requires mock PR object")
	
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = service
		_ = ctx
	}
}
