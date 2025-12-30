package database

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.uber.org/zap"
)

func Backup(backupPath string, dbPath string, maxBackups int, logger *zap.Logger) error {
	// バックアップディレクトリの作成
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("バックアップディレクトリの作成に失敗しました: %w", err)
	}

	// バックアップファイル名（タイムスタンプ付き）
	timestamp := time.Now().Format("20060102_150405")
	backupFileName := fmt.Sprintf("upgo_%s.db", timestamp)
	backupFilePath := filepath.Join(backupPath, backupFileName)

	// データベースファイルをコピー
	sourceFile, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("データベースファイルのオープンに失敗しました: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupFilePath)
	if err != nil {
		return fmt.Errorf("バックアップファイルの作成に失敗しました: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("バックアップファイルのコピーに失敗しました: %w", err)
	}

	logger.Info("バックアップが完了しました", zap.String("path", backupFilePath))

	// 古いバックアップの削除
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

	// ファイルを更新時刻でソート（古い順）
	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// 古いファイルを削除
	for i := 0; i < len(files)-maxBackups; i++ {
		if err := os.Remove(files[i]); err != nil {
			logger.Warn("バックアップファイルの削除に失敗しました", zap.String("file", files[i]), zap.Error(err))
		} else {
			logger.Info("古いバックアップを削除しました", zap.String("file", files[i]))
		}
	}

	return nil
}
