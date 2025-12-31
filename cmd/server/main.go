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
	"upgo/internal/logger"
	"upgo/internal/scheduler"
	"upgo/internal/service"

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

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	gerritClient := gerrit.NewClient(cfg.Gerrit.BaseURL, log)

	router.GET("/health", func(c *gin.Context) {
		if err := database.Get().Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	gerritSyncService := service.NewGerritSyncService(
		database.Get(),
		gerritClient,
		log,
		cfg,
	)

	changeHandlers := api.NewChangeHandlers(database.Get(), log)
	syncHandlers := api.NewSyncHandlers(gerritSyncService, log)

	apiGroup := router.Group("/api")
	{
		apiGroup.GET("/changes", changeHandlers.GetChanges)
		apiGroup.GET("/changes/:id", changeHandlers.GetChange)
		apiGroup.GET("/branches", changeHandlers.GetBranches)
		apiGroup.GET("/statuses", changeHandlers.GetStatuses)

		apiGroup.POST("/sync", syncHandlers.TriggerSync)
		apiGroup.POST("/sync/change/:change_number", syncHandlers.SyncChange)
		apiGroup.GET("/sync/check", syncHandlers.CheckUpdates)
	}

	router.Static("/assets", "./web/dist/assets")
	router.StaticFile("/favicon.ico", "./web/dist/favicon.ico")

	router.GET("/", func(c *gin.Context) {
		c.File("./web/dist/index.html")
	})

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./web/dist/index.html")
	})

	var schedulerCancel context.CancelFunc = nil
	if cfg.Scheduler.Enabled {
		sched, err := scheduler.NewScheduler(
			cfg.Scheduler.Interval,
			cfg.Scheduler.Enabled,
			func(ctx context.Context) error {
				_, err := gerritSyncService.CheckUpdates(ctx)
				return err
			},
			log,
		)
		if err != nil {
			log.Fatal("スケジューラーの初期化に失敗しました", zap.Error(err))
		}

		var schedulerCtx context.Context
		schedulerCtx, schedulerCancel = context.WithCancel(context.Background())

		go sched.Start(schedulerCtx)
		log.Info("スケジューラーを起動しました（更新チェック）", zap.String("interval", cfg.Scheduler.Interval))
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

	log.Info("サーバーが正常にシャットダウンしました")
}

func initializeChecks(cfg *config.Config, log *zap.Logger) error {
	if cfg.Gerrit.BaseURL == "" {
		return fmt.Errorf("Gerrit URLが設定されていません")
	}
	log.Info("Gerrit設定の確認が完了しました", zap.String("base_url", cfg.Gerrit.BaseURL))

	dbDir := filepath.Dir(cfg.Database.Path)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return fmt.Errorf("データベースディレクトリの作成に失敗しました: %w", err)
	}
	log.Info("データベースディレクトリの確認が完了しました", zap.String("path", dbDir))

	return nil
}
