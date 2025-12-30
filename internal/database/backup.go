package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

func Backup(backupPath string, dbPath string, maxBackups int, logger *zap.Logger) error {
	// Create backup directory
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
	}

	// Backup file name with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("upgo_%s.db", timestamp)
	backupFilePath := filepath.Join(backupPath, backupFileName)

	// Use SQLite's VACUUM INTO for safe backup of active database
	// This ensures consistency even when the database is in use
	if _, err := DB.Exec(fmt.Sprintf("VACUUM INTO '%s'", backupFilePath)); err != nil {
		return fmt.Errorf("バックアップの作成に失敗しました: %w", err)
	}

	logger.Info("バックアップが完了しました", zap.String("path", backupFilePath))

	// Delete old backups
	if maxBackups > 0 {
		if err := cleanupOldBackups(backupPath, maxBackups, logger); err != nil {
			logger.Warn("古いバックアップの削除に失敗しました", zap.Error(err))
		}
	}

	return nil
}

func cleanupOldBackups(backupPath string, maxBackups int, logger *zap.Logger) error {
	files, err := filepath.Glob(filepath.Join(backupPath, "upgo_*.db"))
	if err != nil {
		return err
	}

	if len(files) <= maxBackups {
		return nil
	}

	// Sort files by modification time (oldest first)
	// Files with errors are placed at the end
	sort.Slice(files, func(i, j int) bool {
		infoI, errI := os.Stat(files[i])
		infoJ, errJ := os.Stat(files[j])
		
		// If an error occurs, place the file with error at the end
		if errI != nil {
			return false
		}
		if errJ != nil {
			return true
		}
		
		// Compare only when both file info are successfully retrieved
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Delete old files
	for i := 0; i < len(files)-maxBackups; i++ {
		if err := os.Remove(files[i]); err != nil {
			logger.Warn("バックアップファイルの削除に失敗しました", zap.String("file", files[i]), zap.Error(err))
		} else {
			logger.Info("古いバックアップを削除しました", zap.String("file", files[i]))
		}
	}

	return nil
}
