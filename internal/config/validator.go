package config

import (
	"fmt"
	"os"
	"strings"
)

func Validate(cfg *Config) error {
	var errors []string

	// Validate Repository settings
	if cfg.Repository.Owner == "" {
		errors = append(errors, "repository.owner が設定されていません")
	}
	if cfg.Repository.Name == "" {
		errors = append(errors, "repository.name が設定されていません")
	}

	// Validate Gerrit settings
	if cfg.Gerrit.BaseURL == "" {
		errors = append(errors, "gerrit.base_url が設定されていません")
	}

	// Validate Gitiles settings
	if cfg.Gitiles.BaseURL == "" {
		errors = append(errors, "gitiles.base_url が設定されていません")
	}

	// Validate GerritFetch settings
	if cfg.GerritFetch.Project == "" {
		errors = append(errors, "gerrit_fetch.project が設定されていません")
	}
	if len(cfg.GerritFetch.Branches) == 0 {
		errors = append(errors, "gerrit_fetch.branches が設定されていません")
	}
	if len(cfg.GerritFetch.Status) == 0 {
		errors = append(errors, "gerrit_fetch.status が設定されていません")
	}
	if cfg.GerritFetch.Days <= 0 {
		errors = append(errors, "gerrit_fetch.days は1以上である必要があります")
	}
	if cfg.GerritFetch.DiffSizeLimit <= 0 {
		errors = append(errors, "gerrit_fetch.diff_size_limit は1以上である必要があります")
	}

	// GitHub settings are optional (for backward compatibility during migration)
	// Validate only if token is provided
	if cfg.GitHub.Token != "" {
		// Try to get from environment variable
		if strings.HasPrefix(cfg.GitHub.Token, "${") && strings.HasSuffix(cfg.GitHub.Token, "}") {
			envVar := strings.TrimPrefix(strings.TrimSuffix(cfg.GitHub.Token, "}"), "${")
			if val := os.Getenv(envVar); val == "" {
				errors = append(errors, fmt.Sprintf("環境変数 %s が設定されていません", envVar))
			}
		}
	}

	// LLM settings are optional (analysis feature is disabled)
	// Validate only if provider is set (for backward compatibility)
	if cfg.LLM.Provider != "" {
		if cfg.LLM.BaseURL == "" {
			errors = append(errors, "llm.base_url が設定されていません（llm.provider が設定されている場合）")
		}
		if cfg.LLM.Model == "" {
			errors = append(errors, "llm.model が設定されていません（llm.provider が設定されている場合）")
		}
		if cfg.LLM.Timeout <= 0 {
			errors = append(errors, "llm.timeout は1以上である必要があります（llm.provider が設定されている場合）")
		}
	}

	// Validate Database settings
	// devとprdの両方が設定されている必要がある
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

	if len(errors) > 0 {
		return fmt.Errorf("設定の検証に失敗しました:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
