package api

import (
	"database/sql"

	"upgo/internal/config"
	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(router *gin.Engine, db *sql.DB, syncService *service.SyncService, cfg *config.Config, logger *zap.Logger) {
	handlers := NewHandlers(db, logger)
	syncHandler := NewSyncHandler(syncService, logger)
	backupHandler := NewBackupHandler(cfg, logger)

	api := router.Group("/api/v1")
	{
		api.GET("/prs", handlers.GetPRs)
		api.GET("/prs/:id", handlers.GetPR)

		api.GET("/issues", handlers.GetIssues)
		api.GET("/issues/:id", handlers.GetIssue)

		api.POST("/sync", syncHandler.Sync)
		api.GET("/sync/status", syncHandler.GetSyncStatus)

		api.POST("/backup", backupHandler.Backup)
	}
}
