package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Gerrit    GerritConfig    `mapstructure:"gerrit"`
	Gitiles   GitilesConfig   `mapstructure:"gitiles"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Server    ServerConfig    `mapstructure:"server"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Backup    BackupConfig    `mapstructure:"backup"`
	Sync      SyncConfig      `mapstructure:"sync"`
	Diff      DiffConfig      `mapstructure:"diff"`
}

type GerritConfig struct {
	BaseURL  string   `mapstructure:"base_url"`
	Project  string   `mapstructure:"project"`
	Branches []string `mapstructure:"branches"`
	Status   []string `mapstructure:"status"`
}

type GitilesConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type SchedulerConfig struct {
	Interval string `mapstructure:"interval"`
	Enabled  bool   `mapstructure:"enabled"`
}

type SyncConfig struct {
	UpdatedDays int `mapstructure:"updated_days"`
	SafetyWindow int `mapstructure:"safety_window_minutes"`
}

type DiffConfig struct {
	MaxSizeBytes    int      `mapstructure:"max_size_bytes"`
	ExcludePaths    []string `mapstructure:"exclude_paths"`
	ExcludePatterns []string `mapstructure:"exclude_patterns"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"-"`
	Dev  string `mapstructure:"dev"`
	Prd  string `mapstructure:"prd"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type LoggingConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type BackupConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Interval   string `mapstructure:"interval"`
	MaxBackups int    `mapstructure:"max_backups"`
	Path       string `mapstructure:"path"`
}

var AppConfig *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configPath)

	viper.SetEnvPrefix("UPGO")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("設定の解析に失敗しました: %w", err)
	}

	env := os.Getenv("UPGO_ENV")
	isProduction := env == "production" || env == "prod"

	if isProduction {
		config.Database.Path = config.Database.Prd
	} else {
		config.Database.Path = config.Database.Dev
	}

	if config.Gerrit.BaseURL == "" {
		config.Gerrit.BaseURL = "https://go-review.googlesource.com"
	}
	if config.Gerrit.Project == "" {
		config.Gerrit.Project = "go"
	}
	if len(config.Gerrit.Branches) == 0 {
		config.Gerrit.Branches = []string{"master", "release-branch.go1.*"}
	}
	if len(config.Gerrit.Status) == 0 {
		config.Gerrit.Status = []string{"open", "merged"}
	}
	if config.Gitiles.BaseURL == "" {
		config.Gitiles.BaseURL = "https://go.googlesource.com"
	}
	if config.Sync.UpdatedDays == 0 {
		config.Sync.UpdatedDays = 30
	}
	if config.Sync.SafetyWindow == 0 {
		config.Sync.SafetyWindow = 10
	}
	if config.Diff.MaxSizeBytes == 0 {
		config.Diff.MaxSizeBytes = 1024 * 1024
	}

	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("設定の検証に失敗しました: %w", err)
	}

	AppConfig = &config
	return &config, nil
}

func Get() *Config {
	return AppConfig
}
