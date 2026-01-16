package service

import (
	"context"
	"database/sql"
	"testing"

	"upgo/internal/llm"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

// BenchmarkAnalysisService_AnalyzePR benchmarks the PR analysis operation
func BenchmarkAnalysisService_AnalyzePR(b *testing.B) {
	// Setup: Create in-memory database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		b.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize database schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS pull_requests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT,
			body TEXT,
			state TEXT,
			author TEXT
		);
		CREATE TABLE IF NOT EXISTS pull_request_comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pr_id INTEGER NOT NULL,
			body TEXT,
			created_at DATETIME
		);
		CREATE TABLE IF NOT EXISTS pull_request_diffs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pr_id INTEGER NOT NULL,
			diff_text TEXT
		);
		CREATE TABLE IF NOT EXISTS pull_request_summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pr_id INTEGER NOT NULL,
			description_summary TEXT,
			diff_summary TEXT,
			diff_explanation TEXT,
			comments_summary TEXT,
			discussion_summary TEXT,
			merge_reason TEXT,
			close_reason TEXT,
			updated_at DATETIME,
			UNIQUE(pr_id)
		);
	`)
	if err != nil {
		b.Fatalf("Failed to create schema: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO pull_requests (id, title, body, state, author) 
		VALUES (1, 'Test PR', 'Test body', 'open', 'test-author')
	`)
	if err != nil {
		b.Fatalf("Failed to insert test PR: %v", err)
	}

	logger := zap.NewNop()
	// Initialize LLM client properly (but skip if Ollama is not available)
	// For benchmark, we'll use a dummy URL - the test will fail gracefully if Ollama is not running
	llmClient := llm.NewClient("http://localhost:11434", "llama3.2", 30, logger)
	summarizer := llm.NewSummarizer(llmClient, logger)
	analyzer := llm.NewAnalyzer(llmClient, logger)

	service := NewAnalysisService(db, summarizer, analyzer, logger)

	ctx := context.Background()

	// Check if Ollama is available, skip if not
	if err := llmClient.CheckConnection(ctx); err != nil {
		b.Skipf("Ollama is not available: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := service.AnalyzePR(ctx, 1); err != nil {
			b.Fatalf("AnalyzePR failed: %v", err)
		}
	}
}
