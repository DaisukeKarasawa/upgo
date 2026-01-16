package api

import (
	"database/sql"

	"upgo/internal/config"
	"upgo/legacy/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRoutes(router *gin.Engine, db *sql.DB, syncService *service.SyncService, updateCheckService *service.UpdateCheckService, cfg *config.Config, logger *zap.Logger) *SyncHandler {
	handlers := NewHandlers(db, logger)
	// Default to 3 concurrent sync operations
	syncHandler := NewSyncHandler(syncService, logger, 3)
	backupHandler := NewBackupHandler(cfg, logger)
	clearHandler := NewClearHandler(logger, syncHandler)
	updateHandler := NewUpdateHandler(updateCheckService, logger)

	api := router.Group("/api/v1")
	{
		api.GET("/prs", handlers.GetPRs)
		api.GET("/prs/:id", handlers.GetPR)
		api.POST("/prs/:id/sync", syncHandler.SyncPR)

		api.POST("/sync", syncHandler.Sync)
		api.GET("/sync/status", syncHandler.GetSyncStatus)

		api.GET("/updates/dashboard", updateHandler.CheckDashboardUpdates)
		api.GET("/updates/pr/:id", updateHandler.CheckPRUpdates)

		api.POST("/backup", backupHandler.Backup)
		api.POST("/clear", clearHandler.Clear)
	}

	return syncHandler
}
