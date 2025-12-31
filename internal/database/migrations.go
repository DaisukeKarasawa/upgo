package database

import (
	"fmt"
	"strings"

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
		// Add head_sha column to pull_requests table (if not exists)
		addHeadShaToPullRequests,
		// Gerrit Change tables
		createChangesTable,
		createRevisionsTable,
		createChangeFilesTable,
		createChangeDiffsTable,
		createChangeCommentsTable,
		createChangeLabelsTable,
		createChangeMessagesTable,
	}

	for i, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			// SQLite returns "duplicate column name" when column already exists
			if strings.Contains(err.Error(), "duplicate column name") {
				logger.Debug("マイグレーションをスキップしました（カラムが既に存在します）", zap.Int("number", i+1))
			} else {
				return fmt.Errorf("マイグレーション %d の実行に失敗しました: %w", i+1, err)
			}
		} else {
			logger.Info("マイグレーションを実行しました", zap.Int("number", i+1))
		}
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
    head_sha TEXT,
    UNIQUE(repository_id, github_id),
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
`

const addHeadShaToPullRequests = `
ALTER TABLE pull_requests ADD COLUMN head_sha TEXT;
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

// Gerrit Change tables
const createChangesTable = `
CREATE TABLE IF NOT EXISTS changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    repository_id INTEGER NOT NULL,
    change_number INTEGER NOT NULL,
    change_id TEXT NOT NULL,
    project TEXT NOT NULL,
    branch TEXT NOT NULL,
    subject TEXT NOT NULL,
    message TEXT,
    status TEXT NOT NULL,
    previous_status TEXT,
    owner TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    submitted_at DATETIME,
    url TEXT NOT NULL,
    last_synced_at DATETIME,
    UNIQUE(repository_id, change_number),
    FOREIGN KEY (repository_id) REFERENCES repositories(id)
);
CREATE INDEX IF NOT EXISTS idx_changes_change_id ON changes(change_id);
CREATE INDEX IF NOT EXISTS idx_changes_status ON changes(status);
CREATE INDEX IF NOT EXISTS idx_changes_updated_at ON changes(updated_at);
`

const createRevisionsTable = `
CREATE TABLE IF NOT EXISTS revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_id INTEGER NOT NULL,
    patch_set_number INTEGER NOT NULL,
    revision_sha TEXT NOT NULL,
    uploader TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    kind TEXT,
    commit_message TEXT,
    author_name TEXT,
    author_email TEXT,
    committer_name TEXT,
    committer_email TEXT,
    UNIQUE(change_id, patch_set_number),
    FOREIGN KEY (change_id) REFERENCES changes(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_revisions_revision_sha ON revisions(revision_sha);
`

const createChangeFilesTable = `
CREATE TABLE IF NOT EXISTS change_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    revision_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT,
    old_path TEXT,
    lines_inserted INTEGER DEFAULT 0,
    lines_deleted INTEGER DEFAULT 0,
    size_delta INTEGER DEFAULT 0,
    size INTEGER DEFAULT 0,
    binary INTEGER DEFAULT 0,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_change_files_revision_id ON change_files(revision_id);
CREATE INDEX IF NOT EXISTS idx_change_files_file_path ON change_files(file_path);
`

const createChangeDiffsTable = `
CREATE TABLE IF NOT EXISTS change_diffs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    revision_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    diff_text TEXT,
    diff_size INTEGER DEFAULT 0,
    created_at DATETIME NOT NULL,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_change_diffs_revision_id ON change_diffs(revision_id);
CREATE INDEX IF NOT EXISTS idx_change_diffs_file_path ON change_diffs(file_path);
`

const createChangeCommentsTable = `
CREATE TABLE IF NOT EXISTS change_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_id INTEGER NOT NULL,
    revision_id INTEGER,
    comment_id TEXT NOT NULL,
    file_path TEXT,
    line INTEGER,
    patch_set_number INTEGER NOT NULL,
    message TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    in_reply_to TEXT,
    unresolved INTEGER DEFAULT 0,
    FOREIGN KEY (change_id) REFERENCES changes(id) ON DELETE CASCADE,
    FOREIGN KEY (revision_id) REFERENCES revisions(id) ON DELETE SET NULL,
    UNIQUE(change_id, comment_id)
);
CREATE INDEX IF NOT EXISTS idx_change_comments_change_id ON change_comments(change_id);
CREATE INDEX IF NOT EXISTS idx_change_comments_revision_id ON change_comments(revision_id);
CREATE INDEX IF NOT EXISTS idx_change_comments_file_path ON change_comments(file_path);
`

const createChangeLabelsTable = `
CREATE TABLE IF NOT EXISTS change_labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_id INTEGER NOT NULL,
    label_name TEXT NOT NULL,
    account TEXT NOT NULL,
    value INTEGER NOT NULL,
    date DATETIME NOT NULL,
    FOREIGN KEY (change_id) REFERENCES changes(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_change_labels_change_id ON change_labels(change_id);
CREATE INDEX IF NOT EXISTS idx_change_labels_label_name ON change_labels(label_name);
`

const createChangeMessagesTable = `
CREATE TABLE IF NOT EXISTS change_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_id INTEGER NOT NULL,
    message_id TEXT NOT NULL,
    author TEXT NOT NULL,
    message TEXT NOT NULL,
    date DATETIME NOT NULL,
    revision_number INTEGER,
    FOREIGN KEY (change_id) REFERENCES changes(id) ON DELETE CASCADE,
    UNIQUE(change_id, message_id)
);
CREATE INDEX IF NOT EXISTS idx_change_messages_change_id ON change_messages(change_id);
`

// ClearAllTables drops all tables and recreates them by running migrations
// This effectively clears all data from the database
func ClearAllTables(logger *zap.Logger) error {
	// Drop tables in reverse order of dependencies to avoid foreign key constraint errors
	// Order: dependent tables first, then parent tables
	// - pull_request_diffs, pull_request_comments, pull_request_summaries depend on pull_requests
	// - mental_model_analyses depends on repositories
	// - pull_requests depends on repositories
	// - change_diffs, change_files, change_comments, change_labels, change_messages depend on revisions/changes
	// - revisions depend on changes
	// - changes depend on repositories
	tables := []string{
		"pull_request_diffs",        // Dependent table
		"pull_request_comments",     // Dependent table
		"pull_request_summaries",    // Dependent table
		"mental_model_analyses",     // Dependent table (depends on repositories)
		"pull_requests",             // Parent table (depends on repositories)
		"change_diffs",              // Dependent table (depends on revisions)
		"change_files",              // Dependent table (depends on revisions)
		"change_comments",           // Dependent table (depends on changes/revisions)
		"change_labels",             // Dependent table (depends on changes)
		"change_messages",           // Dependent table (depends on changes)
		"revisions",                 // Parent table (depends on changes)
		"changes",                   // Parent table (depends on repositories)
		"repositories",              // Parent table
		"sync_jobs",                 // Independent table
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
