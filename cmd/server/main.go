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
	"upgo/internal/gerrit"
	"upgo/internal/gitiles"
	"upgo/internal/llm"
	"upgo/internal/logger"
	"upgo/internal/scheduler"
	"upgo/internal/service"
	"upgo/internal/tracker"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定の読み込みに失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "設定の検証に失敗しました: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Output, cfg.Logging.FilePath); err != nil {
		fmt.Fprintf(os.Stderr, "ロガーの初期化に失敗しました: %v\n", err)
		os.Exit(1)
	}
	log := logger.Get()
	defer log.Sync()

	if err := initializeChecks(cfg, log); err != nil {
		log.Fatal("初期化チェックに失敗しました", zap.Error(err))
	}

	if err := database.Connect(cfg.Database.Path, log); err != nil {
		log.Fatal("データベース接続に失敗しました", zap.Error(err))
	}
	defer database.Close()

	if err := database.RunMigrations(log); err != nil {
		log.Fatal("マイグレーションの実行に失敗しました", zap.Error(err))
	}

	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware: Allow all origins for development convenience.
	// In production, consider restricting to specific origins for security.
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		// Handle preflight OPTIONS requests immediately
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Gerrit clients
	gerritClient := gerrit.NewClient(cfg.Gerrit.BaseURL, cfg.Gerrit.Username, cfg.Gerrit.Password, log)
	_ = gitiles.NewClient(cfg.Gitiles.BaseURL, log) // Gitiles client (for future use)
	
	// Gerrit change fetcher
	changeFetcher := gerrit.NewChangeFetcher(
		gerritClient,
		log,
		cfg.GerritFetch.Project,
		cfg.GerritFetch.Branches,
		cfg.GerritFetch.Status,
		cfg.GerritFetch.Days,
	)
	
	// Diff policy
	diffPolicy := gerrit.NewDiffPolicy(cfg.GerritFetch.DiffSizeLimit)
	
	statusTracker := tracker.NewStatusTracker(database.Get(), log)

	// LLM client (for backward compatibility, but analysis is disabled)
	_ = llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, log) // LLM client (analysis disabled)

	router.GET("/health", func(c *gin.Context) {
		if err := database.Get().Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		// Ollama check is optional (analysis is disabled)
		// if err := llmClient.CheckConnection(context.Background()); err != nil {
		// 	c.JSON(http.StatusServiceUnavailable, gin.H{
		// 		"status": "unhealthy",
		// 		"error":  "ollama connection failed",
		// 	})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})
	
	// Analysis service (disabled - code kept but not used)
	// summarizer := llm.NewSummarizer(llmClient, log)
	// analyzer := llm.NewAnalyzer(llmClient, log)
	// analysisService := service.NewAnalysisService(database.Get(), summarizer, analyzer, log)

	// Gerrit sync service
	gerritSyncService := service.NewGerritSyncService(
		database.Get(),
		gerritClient,
		changeFetcher,
		diffPolicy,
		statusTracker,
		log,
		cfg.GerritFetch.Project,
	)

	// Gerrit update check service
	gerritUpdateCheckService := service.NewGerritUpdateCheckService(
		database.Get(),
		gerritClient,
		changeFetcher,
		log,
		cfg.GerritFetch.Project,
	)

	// Setup Gerrit routes (primary)
	gerritSyncHandler := api.SetupGerritRoutes(router, database.Get(), gerritSyncService, gerritUpdateCheckService, cfg, log)

	// GitHub services are disabled (migrated to Gerrit)
	// Keeping code for reference but not registering routes to avoid conflicts
	// githubClient := github.NewClient(cfg.GitHub.Token, log)
	// prFetcher := github.NewPRFetcher(githubClient, log)
	// syncService := service.NewSyncService(...)
	// updateCheckService := service.NewUpdateCheckService(...)
	// _ = api.SetupRoutes(router, database.Get(), syncService, updateCheckService, cfg, log)

	// Serve static assets
	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")
	
	// SPA routing: Serve index.html for root and all non-API routes
	// This allows React Router to handle all frontend routes.
	router.GET("/", func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})
	
	router.NoRoute(func(c *gin.Context) {
		// API routes that don't exist return 404
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		// All other routes serve index.html for SPA routing
		c.File("./web/dist/index.html")
	})

	// Conditionally start the scheduler if enabled in config.
	// The scheduler runs periodic update checks (lightweight) in the background.
	// For Gerrit: runs daily at 00:00 to check for updates (badge only).
	// Full sync is manual-only.
	var schedulerCancel context.CancelFunc = nil
	if cfg.Scheduler.Enabled {
		// Check if interval is cron format
		var sched interface {
			Start(ctx context.Context)
			Stop()
		}

		if scheduler.IsCronSpec(cfg.Scheduler.Interval) {
			// Use cron scheduler for cron format (e.g., "0 0 * * *")
			cronSched, err := scheduler.NewCronScheduler(
				cfg.Scheduler.Interval,
				cfg.Scheduler.Enabled,
				func(ctx context.Context) error {
					// Use Gerrit update check service for badge polling
					_, err := gerritUpdateCheckService.CheckDashboardUpdates(ctx)
					return err
				},
				log,
			)
			if err != nil {
				log.Fatal("Cronスケジューラーの初期化に失敗しました", zap.Error(err))
			}
			sched = cronSched
		} else {
			// Use duration-based scheduler for duration format (e.g., "24h")
			durationSched, err := scheduler.NewScheduler(
				cfg.Scheduler.Interval,
				cfg.Scheduler.Enabled,
				func(ctx context.Context) error {
					// Use Gerrit update check service for badge polling
					_, err := gerritUpdateCheckService.CheckDashboardUpdates(ctx)
					return err
				},
				log,
			)
			if err != nil {
				log.Fatal("スケジューラーの初期化に失敗しました", zap.Error(err))
			}
			sched = durationSched
		}

		var schedulerCtx context.Context
		schedulerCtx, schedulerCancel = context.WithCancel(context.Background())

		go sched.Start(schedulerCtx)
		log.Info("スケジューラーを起動しました（Gerrit更新チェック）", zap.String("interval", cfg.Scheduler.Interval))
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("サーバーの起動に失敗しました", zap.Error(err))
		}
	}()

	log.Info("サーバーが起動しました", zap.String("address", addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("サーバーをシャットダウンしています...")

	if schedulerCancel != nil {
		schedulerCancel()
		log.Info("スケジューラーの停止を要求しました")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("サーバーのシャットダウンに失敗しました", zap.Error(err))
	}

	// Wait for all in-flight sync operations to complete
	log.Info("実行中の同期操作の完了を待機しています...")
	gerritSyncHandler.Wait()
	log.Info("すべての同期操作が完了しました")

	log.Info("サーバーが正常にシャットダウンしました")
}

func initializeChecks(cfg *config.Config, log *zap.Logger) error {
	// Gerrit configuration check
	if cfg.Gerrit.BaseURL == "" {
		return fmt.Errorf("Gerrit base_urlが設定されていません")
	}
	log.Info("Gerrit設定の確認が完了しました", zap.String("base_url", cfg.Gerrit.BaseURL))

	// Gitiles configuration check
	if cfg.Gitiles.BaseURL == "" {
		return fmt.Errorf("Gitiles base_urlが設定されていません")
	}
	log.Info("Gitiles設定の確認が完了しました", zap.String("base_url", cfg.Gitiles.BaseURL))

	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("データベースディレクトリの作成に失敗しました: %w", err)
	}
	log.Info("データベースディレクトリの確認が完了しました", zap.String("path", dbDir))

	// LLM/Ollama check is optional (analysis is disabled)
	// llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.Model, cfg.LLM.Timeout, log)
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err := llmClient.CheckConnection(ctx); err != nil {
	// 	return fmt.Errorf("Ollama接続確認に失敗しました: %w", err)
	// }
	// log.Info("Ollama接続確認が完了しました", zap.String("model", cfg.LLM.Model))

	return nil
}
