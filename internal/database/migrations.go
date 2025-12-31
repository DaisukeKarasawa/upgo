package database

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func RunMigrations(logger *zap.Logger) error {
	migrations := []string{
		createChangesTable,
		createRevisionsTable,
		createFilesTable,
		createDiffsTable,
		createCommentsTable,
		createLabelsTable,
		createMessagesTable,
		createCommitsTable,
		createSyncJobsTable,
		createChangesIndexes,
	}

	for i, migration := range migrations {
		_, err := DB.Exec(migration)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate column name") ||
				strings.Contains(err.Error(), "already exists") {
				logger.Debug("マイグレーションをスキップしました（既に存在します）", zap.Int("number", i+1))
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

const createChangesTable = `
CREATE TABLE IF NOT EXISTS changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_id TEXT NOT NULL UNIQUE,
    change_number INTEGER NOT NULL,
    project TEXT NOT NULL,
    branch TEXT NOT NULL,
    status TEXT NOT NULL,
    subject TEXT NOT NULL,
    message TEXT,
    owner_name TEXT NOT NULL,
    owner_email TEXT,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    submitted DATETIME,
    last_synced_at DATETIME
);
`

const createRevisionsTable = `
CREATE TABLE IF NOT EXISTS revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_db_id INTEGER NOT NULL,
    revision_id TEXT NOT NULL,
    patchset_num INTEGER NOT NULL,
    uploader_name TEXT,
    uploader_email TEXT,
    created DATETIME NOT NULL,
    commit_message TEXT,
    FOREIGN KEY (change_db_id) REFERENCES changes(id) ON DELETE CASCADE,
    UNIQUE(change_db_id, patchset_num)
);
`

const createFilesTable = `
CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    revision_db_id INTEGER NOT NULL,
    file_path TEXT NOT NULL,
    status TEXT,
    lines_inserted INTEGER DEFAULT 0,
    lines_deleted INTEGER DEFAULT 0,
    size_delta INTEGER DEFAULT 0,
    FOREIGN KEY (revision_db_id) REFERENCES revisions(id) ON DELETE CASCADE,
    UNIQUE(revision_db_id, file_path)
);
`

const createDiffsTable = `
CREATE TABLE IF NOT EXISTS diffs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_db_id INTEGER NOT NULL,
    diff_content TEXT,
    is_binary INTEGER DEFAULT 0,
    size_exceeded INTEGER DEFAULT 0,
    FOREIGN KEY (file_db_id) REFERENCES files(id) ON DELETE CASCADE,
    UNIQUE(file_db_id)
);
`

const createCommentsTable = `
CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_db_id INTEGER NOT NULL,
    comment_id TEXT NOT NULL,
    revision_db_id INTEGER,
    file_path TEXT,
    line INTEGER,
    author_name TEXT NOT NULL,
    author_email TEXT,
    message TEXT NOT NULL,
    created DATETIME NOT NULL,
    updated DATETIME NOT NULL,
    in_reply_to TEXT,
    unresolved INTEGER DEFAULT 0,
    FOREIGN KEY (change_db_id) REFERENCES changes(id) ON DELETE CASCADE,
    FOREIGN KEY (revision_db_id) REFERENCES revisions(id) ON DELETE SET NULL,
    UNIQUE(change_db_id, comment_id)
);
`

const createLabelsTable = `
CREATE TABLE IF NOT EXISTS labels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_db_id INTEGER NOT NULL,
    label_name TEXT NOT NULL,
    value INTEGER NOT NULL,
    account_name TEXT,
    account_email TEXT,
    granted_on DATETIME NOT NULL,
    FOREIGN KEY (change_db_id) REFERENCES changes(id) ON DELETE CASCADE
);
`

const createMessagesTable = `
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    change_db_id INTEGER NOT NULL,
    message_id TEXT NOT NULL,
    author_name TEXT,
    author_email TEXT,
    message TEXT NOT NULL,
    date DATETIME NOT NULL,
    revision_number INTEGER,
    FOREIGN KEY (change_db_id) REFERENCES changes(id) ON DELETE CASCADE,
    UNIQUE(change_db_id, message_id)
);
`

const createCommitsTable = `
CREATE TABLE IF NOT EXISTS commits (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    hash TEXT NOT NULL UNIQUE,
    author_name TEXT NOT NULL,
    author_email TEXT,
    commit_date DATETIME NOT NULL,
    message TEXT,
    change_id TEXT,
    reviewed_on TEXT,
    change_db_id INTEGER,
    FOREIGN KEY (change_db_id) REFERENCES changes(id) ON DELETE SET NULL
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

const createChangesIndexes = `
CREATE INDEX IF NOT EXISTS idx_changes_status ON changes(status);
CREATE INDEX IF NOT EXISTS idx_changes_branch ON changes(branch);
CREATE INDEX IF NOT EXISTS idx_changes_updated ON changes(updated);
CREATE INDEX IF NOT EXISTS idx_changes_project ON changes(project);
CREATE INDEX IF NOT EXISTS idx_revisions_change ON revisions(change_db_id);
CREATE INDEX IF NOT EXISTS idx_files_revision ON files(revision_db_id);
CREATE INDEX IF NOT EXISTS idx_comments_change ON comments(change_db_id);
CREATE INDEX IF NOT EXISTS idx_labels_change ON labels(change_db_id);
CREATE INDEX IF NOT EXISTS idx_messages_change ON messages(change_db_id);
CREATE INDEX IF NOT EXISTS idx_commits_change_id ON commits(change_id);
`

func ClearAllTables(logger *zap.Logger) error {
	tables := []string{
		"diffs",
		"files",
		"comments",
		"labels",
		"messages",
		"revisions",
		"commits",
		"changes",
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
	return RunMigrations(logger)
}
