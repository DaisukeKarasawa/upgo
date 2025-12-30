package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var DB *sql.DB

func Connect(dbPath string, logger *zap.Logger) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=1")
	if err != nil {
		return fmt.Errorf("データベース接続に失敗しました: %w", err)
	}

	// 接続確認
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("データベース接続確認に失敗しました: %w", err)
	}

	logger.Info("データベース接続に成功しました", zap.String("path", dbPath))
	return nil
}

func Get() *sql.DB {
	return DB
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
