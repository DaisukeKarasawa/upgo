package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"upgo/internal/api"
	"upgo/internal/config"
	"upgo/internal/database"
	"upgo/internal/github"
	"upgo/internal/llm"
	"upgo/internal/logger"
	"upgo/internal/scheduler"
	"upgo/internal/service"
	"upgo/internal/tracker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 設定の読み込み
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	// 設定の検証
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "設定の検証に失敗しました: %v\n", err)
		os.Exit(1)
	}

	// ロガーの初期化
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Output, cfg.Logging.FilePath); err != nil {
		fmt.Fprintf(os.Stderr, "ロガーの初期化に失敗しました: %v\n", err)
		os.Exit(1)
	}
	log := logger.Get()
	defer log.Sync()

	// 起動時の初期化チェック
	if err := initializeChecks(cfg, log); err != nil {
		log.Fatal("初期化チェックに失敗しました", zap.Error(err))
	}

	// データベース接続
	if err := database.Connect(cfg.Database.Path, log); err != nil {
		log.Fatal("データベース接続に失敗しました", zap.Error(err))
	}
	defer database.Close()

	// マイグレーション実行
	if err := database.RunMigrations(log); err != nil {
		log.Fatal("マイグレーションの実行に失敗しました", zap.Error(err))
	}

	// Ginルーターの設定
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS設定
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// GitHubクライアントの初期化
	githubClient := github.NewClient(cfg.GitHub.Token, log)
	prFetcher := github.NewPRFetcher(githubClient, log)
	issueFetcher := github.NewIssueFetcher(githubClient, log)
	statusTracker := tracker.NewStatusTracker(database.Get(), log)

	// LLMクライアントの初期化
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, log)

	// ヘルスチェックエンドポイント（初期化済みのクライアントを再利用）
	router.GET("/health", func(c *gin.Context) {
		// DB接続確認
		if err := database.Get().Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		// Ollama接続確認（初期化済みのクライアントを再利用）
		if err := llmClient.CheckConnection(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "ollama connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	summarizer := llm.NewSummarizer(llmClient, log)
	analyzer := llm.NewAnalyzer(llmClient, log)
	analysisService := service.NewAnalysisService(database.Get(), summarizer, analyzer, log)

	// 同期サービスの初期化
	syncService := service.NewSyncService(
		database.Get(),
		githubClient,
		prFetcher,
		issueFetcher,
		statusTracker,
		analysisService,
		log,
		cfg.Repository.Owner,
		cfg.Repository.Name,
	)

	// APIルーティングの設定
	api.SetupRoutes(router, database.Get(), syncService, log)

	// 静的ファイルの配信（フロントエンド）
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
	router.NoRoute(func(c *gin.Context) {
		// APIエンドポイント以外はindex.htmlを返す（SPA用）
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.File("./web/dist/index.html")
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		}
	})

	// スケジューラーの初期化と起動
	var schedulerCancel context.CancelFunc
	if cfg.Scheduler.Enabled {
		sched, err := scheduler.NewScheduler(
			cfg.Scheduler.Interval,
			cfg.Scheduler.Enabled,
			func(ctx context.Context) error {
				return syncService.Sync(ctx)
			},
			log,
		)
		if err != nil {
			log.Fatal("スケジューラーの初期化に失敗しました", zap.Error(err))
		}

		var schedulerCtx context.Context
		schedulerCtx, schedulerCancel = context.WithCancel(context.Background())

		go sched.Start(schedulerCtx)
		log.Info("スケジューラーを起動しました", zap.String("interval", cfg.Scheduler.Interval))
	}

	// サーバー起動
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// グレースフルシャットダウン
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("サーバーの起動に失敗しました", zap.Error(err))
		}
	}()

	log.Info("サーバーが起動しました", zap.String("address", addr))

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("サーバーをシャットダウンしています...")

	// スケジューラーの停止
	if schedulerCancel != nil {
		schedulerCancel()
		log.Info("スケジューラーの停止を要求しました")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("サーバーのシャットダウンに失敗しました", zap.Error(err))
	}

	log.Info("サーバーが正常にシャットダウンしました")
}

func initializeChecks(cfg *config.Config, log *zap.Logger) error {
	// GitHubトークンの確認
	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHubトークンが設定されていません（環境変数 GITHUB_TOKEN を設定してください）")
	}
	log.Info("GitHubトークンの確認が完了しました")

	// データベースディレクトリの作成
	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("データベースディレクトリの作成に失敗しました: %w", err)
	}
	log.Info("データベースディレクトリの確認が完了しました", zap.String("path", dbDir))

	// Ollama接続確認
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, log)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := llmClient.CheckConnection(ctx); err != nil {
		return fmt.Errorf("Ollama接続確認に失敗しました: %w", err)
	}
	log.Info("Ollama接続確認が完了しました", zap.String("model", cfg.LLM.Model))

	return nil
}
