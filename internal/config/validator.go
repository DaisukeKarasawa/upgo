package config

import (
	"fmt"
	"os"
	"strings"
)

func Validate(cfg *Config) error {
	var errors []string

	// Repository設定の検証
	if cfg.Repository.Owner == "" {
		errors = append(errors, "repository.owner が設定されていません")
	}
	if cfg.Repository.Name == "" {
		errors = append(errors, "repository.name が設定されていません")
	}

	// GitHub設定の検証
	if cfg.GitHub.Token == "" {
		errors = append(errors, "github.token が設定されていません（環境変数 GITHUB_TOKEN を設定してください）")
	} else {
		// 環境変数から取得を試みる
		if strings.HasPrefix(cfg.GitHub.Token, "${") && strings.HasSuffix(cfg.GitHub.Token, "}") {
			envVar := strings.TrimPrefix(strings.TrimSuffix(cfg.GitHub.Token, "}"), "${")
			if val := os.Getenv(envVar); val == "" {
				errors = append(errors, fmt.Sprintf("環境変数 %s が設定されていません", envVar))
			}
		}
	}

	// LLM設定の検証
	if cfg.LLM.Provider == "" {
		errors = append(errors, "llm.provider が設定されていません")
	}
	if cfg.LLM.BaseURL == "" {
		errors = append(errors, "llm.base_url が設定されていません")
	}
	if cfg.LLM.Model == "" {
		errors = append(errors, "llm.model が設定されていません")
	}
	if cfg.LLM.Timeout <= 0 {
		errors = append(errors, "llm.timeout は1以上である必要があります")
	}

	// Database設定の検証
	if cfg.Database.Path == "" {
		errors = append(errors, "database.path が設定されていません")
	}

	// Server設定の検証
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errors = append(errors, "server.port は1-65535の範囲である必要があります")
	}
	if cfg.Server.Host == "" {
		errors = append(errors, "server.host が設定されていません")
	}

	// Logging設定の検証
	validLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLevels[cfg.Logging.Level] {
		errors = append(errors, "logging.level は debug/info/warn/error のいずれかである必要があります")
	}
	if cfg.Logging.Output != "stdout" && cfg.Logging.Output != "file" {
		errors = append(errors, "logging.output は stdout または file である必要があります")
	}
	if cfg.Logging.Output == "file" && cfg.Logging.FilePath == "" {
		errors = append(errors, "logging.file_path が設定されていません（logging.output=file の場合）")
	}

	// Backup設定の検証
	if cfg.Backup.Enabled && cfg.Backup.Path == "" {
		errors = append(errors, "backup.path が設定されていません（backup.enabled=true の場合）")
	}
	if cfg.Backup.MaxBackups < 0 {
		errors = append(errors, "backup.max_backups は0以上である必要があります")
	}

	if len(errors) > 0 {
		return fmt.Errorf("設定の検証に失敗しました:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
