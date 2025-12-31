package config

import (
	"fmt"
	"strings"
)

func Validate(cfg *Config) error {
	var errors []string

	// Validate Gerrit settings
	if cfg.Gerrit.BaseURL == "" {
		errors = append(errors, "gerrit.base_url が設定されていません")
	}
	if cfg.Gerrit.Project == "" {
		errors = append(errors, "gerrit.project が設定されていません")
	}

	// Validate Database settings
	if cfg.Database.Dev == "" {
		errors = append(errors, "database.dev が設定されていません")
	}
	if cfg.Database.Prd == "" {
		errors = append(errors, "database.prd が設定されていません")
	}

	// Validate Server settings
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		errors = append(errors, "server.port は1-65535の範囲である必要があります")
	}
	if cfg.Server.Host == "" {
		errors = append(errors, "server.host が設定されていません")
	}

	// Validate Logging settings
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

	// Validate Backup settings
	if cfg.Backup.Enabled && cfg.Backup.Path == "" {
		errors = append(errors, "backup.path が設定されていません（backup.enabled=true の場合）")
	}
	if cfg.Backup.MaxBackups < 0 {
		errors = append(errors, "backup.max_backups は0以上である必要があります")
	}

	// Validate Sync settings
	if cfg.Sync.UpdatedDays <= 0 {
		errors = append(errors, "sync.updated_days は1以上である必要があります")
	}

	// Validate Diff settings
	if cfg.Diff.MaxSizeBytes <= 0 {
		errors = append(errors, "diff.max_size_bytes は1以上である必要があります")
	}

	if len(errors) > 0 {
		return fmt.Errorf("設定の検証に失敗しました:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
