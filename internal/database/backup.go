package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

func Backup(backupPath string, maxBackups int, logger *zap.Logger) error {
	// Normalize and resolve the backup root directory to prevent path traversal attacks
	// The backupPath from config is treated as the root directory for backups
	backupRootAbs, err := filepath.Abs(filepath.Clean(backupPath))
	if err != nil {
		return fmt.Errorf("バックアップルートディレクトリの解決に失敗しました: %w", err)
	}

	// Ensure the root path doesn't contain any suspicious characters or patterns
	if err := validateBackupPath(backupRootAbs); err != nil {
		return fmt.Errorf("バックアップルートパスの検証に失敗しました: %w", err)
	}

	// Use the resolved root as the backup directory
	absBackupPath := backupRootAbs

	// Create backup directory
	if err := os.MkdirAll(absBackupPath, 0755); err != nil {
		return fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
	}

	// Backup file name with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("upgo_%s.db", timestamp)
	backupFilePath := filepath.Join(absBackupPath, backupFileName)

	// Validate the final backup file path
	if err := validateBackupPath(backupFilePath); err != nil {
		return fmt.Errorf("バックアップファイルパスの検証に失敗しました: %w", err)
	}

	// Use SQLite's VACUUM INTO for safe backup of active database
	// This ensures consistency even when the database is in use
	// Escape single quotes in the path to prevent SQL injection
	escapedPath := escapeSQLString(backupFilePath)
	if _, err := DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", escapedPath)); err != nil {
		return fmt.Errorf("バックアップの作成に失敗しました: %w", err)
	}

	logger.Info("バックアップが完了しました", zap.String("path", backupFilePath))

	// Delete old backups
	if maxBackups > 0 {
		if err := cleanupOldBackups(absBackupPath, maxBackups, logger); err != nil {
			logger.Warn("古いバックアップの削除に失敗しました", zap.Error(err))
		}
	}

	return nil
}

// validateBackupPath validates the backup path to prevent path traversal attacks.
// It checks that the path doesn't contain suspicious patterns like ".." or control characters.
func validateBackupPath(path string) error {
	// Check for path traversal attempts
	if filepath.Clean(path) != path {
		return fmt.Errorf("パストラバーサルが検出されました: %s", path)
	}
	// Check for control characters and other suspicious patterns
	for _, char := range path {
		if char < 32 {
			return fmt.Errorf("不正な文字が検出されました: %s", path)
		}
	}
	return nil
}

// escapeSQLString escapes single quotes in a string for use in SQL.
// This prevents SQL injection when using the path in VACUUM INTO.
func escapeSQLString(s string) string {
	// Replace single quotes with two single quotes (SQL escape)
	result := ""
	for _, char := range s {
		if char == '\'' {
			result += "''"
		} else {
			result += string(char)
		}
	}
	return result
}

func cleanupOldBackups(backupPath string, maxBackups int, logger *zap.Logger) error {
	files, err := filepath.Glob(filepath.Join(backupPath, "upgo_*.db"))
	if err != nil {
		return err
	}

	// Filter out files that fail stat, cache FileInfo, and log warnings
	type fileWithInfo struct {
		path string
		info os.FileInfo
	}
	var statableFiles []fileWithInfo
	for _, file := range files {
		if info, err := os.Stat(file); err != nil {
			logger.Warn("バックアップファイルの状態確認に失敗しました", zap.String("file", file), zap.Error(err))
		} else {
			statableFiles = append(statableFiles, fileWithInfo{path: file, info: info})
		}
	}

	if len(statableFiles) <= maxBackups {
		return nil
	}

	// Sort files by modification time (oldest first)
	sort.Slice(statableFiles, func(i, j int) bool {
		return statableFiles[i].info.ModTime().Before(statableFiles[j].info.ModTime())
	})

	// Delete old files
	for i := 0; i < len(statableFiles)-maxBackups; i++ {
		if err := os.Remove(statableFiles[i].path); err != nil {
			logger.Warn("バックアップファイルの削除に失敗しました", zap.String("file", statableFiles[i].path), zap.Error(err))
		} else {
			logger.Info("古いバックアップを削除しました", zap.String("file", statableFiles[i].path))
		}
	}

	return nil
}
