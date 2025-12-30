package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Repository RepositoryConfig `mapstructure:"repository"`
	GitHub     GitHubConfig      `mapstructure:"github"`
	Scheduler  SchedulerConfig   `mapstructure:"scheduler"`
	LLM        LLMConfig         `mapstructure:"llm"`
	Database   DatabaseConfig    `mapstructure:"database"`
	Server     ServerConfig      `mapstructure:"server"`
	Logging    LoggingConfig     `mapstructure:"logging"`
	Backup     BackupConfig      `mapstructure:"backup"`
}

type RepositoryConfig struct {
	Owner string `mapstructure:"owner"`
	Name  string `mapstructure:"name"`
}

type GitHubConfig struct {
	Token   string `mapstructure:"token"`
	BaseURL string `mapstructure:"base_url"`
}

type SchedulerConfig struct {
	Interval string `mapstructure:"interval"`
	Enabled  bool   `mapstructure:"enabled"`
}

type LLMConfig struct {
	Provider string `mapstructure:"provider"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
	Timeout  int    `mapstructure:"timeout"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"-"` // 内部で使用（dev/prdから自動設定）
	Dev  string `mapstructure:"dev"`   // 開発環境用パス
	Prd  string `mapstructure:"prd"`  // 本番環境用パス
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
	Enabled   bool   `mapstructure:"enabled"`
	Interval  string `mapstructure:"interval"`
	MaxBackups int   `mapstructure:"max_backups"`
	Path      string `mapstructure:"path"`
}

var AppConfig *Config

func Load(configPath string) (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath(configPath)

	// Support environment variables
	viper.SetEnvPrefix("UPGO")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Load config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗しました: %w", err)
	}

	// Expand environment variables
	tokenValue := viper.GetString("github.token")
	if strings.HasPrefix(tokenValue, "${") && strings.HasSuffix(tokenValue, "}") {
		envVar := strings.TrimPrefix(strings.TrimSuffix(tokenValue, "}"), "${")
		if val := os.Getenv(envVar); val != "" {
			viper.Set("github.token", val)
		} else {
			// Also try GITHUB_TOKEN (for backward compatibility)
			if val := os.Getenv("GITHUB_TOKEN"); val != "" {
				viper.Set("github.token", val)
			}
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("設定の解析に失敗しました: %w", err)
	}

	// 環境に応じてデータベースパスを設定
	// 環境変数UPGO_ENVに基づいて自動選択
	env := os.Getenv("UPGO_ENV")
	isProduction := env == "production" || env == "prod"
	
	if isProduction {
		config.Database.Path = config.Database.Prd
	} else {
		config.Database.Path = config.Database.Dev
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
