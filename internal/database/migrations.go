package database

import (
	"fmt"

	"go.uber.org/zap"
)

func RunMigrations(logger *zap.Logger) error {
	migrations := []string{
		createRepositoriesTable,
		createPullRequestsTable,
		createPullRequestSummariesTable,
		createPullRequestCommentsTable,
		createPullRequestDiffsTable,
		createMentalModelAnalysesTable,
		createSyncJobsTable,
	}

	for i, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("マイグレーション %d の実行に失敗しました: %w", i+1, err)
		}
		logger.Info("マイグレーションを実行しました", zap.Int("number", i+1))
	}

	logger.Info("すべてのマイグレーションが完了しました")
	return nil
}

const createRepositoriesTable = `
CREATE TABLE IF NOT EXISTS repositories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    name TEXT NOT NULL,
    last_synced_at DATETIME,
    UNIQUE(owner, name)
);
`

const createPullRequestsTable = `
CREATE TABLE IF NOT EXISTS pull_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    github_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    state TEXT NOT NULL,
    previous_state TEXT,
    author TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    merged_at DATETIME,
    closed_at DATETIME,
    url TEXT NOT NULL,
    last_synced_at DATETIME,
    UNIQUE(repository_id, github_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
`

const createPullRequestSummariesTable = `
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
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    UNIQUE(pr_id)
);
`

const createPullRequestCommentsTable = `
CREATE TABLE IF NOT EXISTS pull_request_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pr_id INTEGER NOT NULL,
    github_id INTEGER NOT NULL,
    body TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    UNIQUE(pr_id, github_id)
);
`

const createPullRequestDiffsTable = `
CREATE TABLE IF NOT EXISTS pull_request_diffs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pr_id INTEGER NOT NULL,
    diff_text TEXT NOT NULL,
    file_path TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (pr_id) REFERENCES pull_requests(id) ON DELETE CASCADE
);
`

const createMentalModelAnalysesTable = `
CREATE TABLE IF NOT EXISTS mental_model_analyses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    analysis_type TEXT NOT NULL,
    analysis_content TEXT NOT NULL,
    analyzed_pr_ids TEXT,
    analyzed_issue_ids TEXT,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
`

const createSyncJobsTable = `
CREATE TABLE IF NOT EXISTS sync_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_type TEXT NOT NULL,
    status TEXT NOT NULL,
    started_at DATETIME NOT NULL,
    completed_at DATETIME,
    error_message TEXT
);
`

// ClearAllTables drops all tables and recreates them by running migrations
// This effectively clears all data from the database
func ClearAllTables(logger *zap.Logger) error {
	// Drop tables in reverse order of dependencies to avoid foreign key constraint errors
	tables := []string{
		"pull_request_diffs",
		"pull_request_comments",
		"pull_request_summaries",
		"pull_requests",
		"mental_model_analyses",
		"repositories",
		"sync_jobs",
	}

	logger.Info("データベースの全テーブルを削除しています...")
	for _, table := range tables {
		if _, err := DB.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("テーブル %s の削除に失敗しました: %w", table, err)
		}
		logger.Info("テーブルを削除しました", zap.String("table", table))
	}

	logger.Info("マイグレーションを再実行してテーブルを再作成しています...")
	// Recreate all tables by running migrations
	return RunMigrations(logger)
}
