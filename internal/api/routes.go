package api

import (
	"database/sql"

	"upgo/internal/config"
	"upgo/internal/service"

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
		// GitHub PR endpoints (deprecated but kept for backward compatibility)
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

// SetupGerritRoutes sets up Gerrit-specific routes
func SetupGerritRoutes(router *gin.Engine, db *sql.DB, gerritSyncService *service.GerritSyncService, gerritUpdateCheckService *service.GerritUpdateCheckService, cfg *config.Config, logger *zap.Logger) *GerritSyncHandler {
	handlers := NewHandlers(db, logger)
	// Default to 3 concurrent sync operations
	gerritSyncHandler := NewGerritSyncHandler(gerritSyncService, logger, 3)
	backupHandler := NewBackupHandler(cfg, logger)
	clearHandler := NewGerritClearHandler(logger, gerritSyncHandler)
	gerritUpdateHandler := NewGerritUpdateHandler(gerritUpdateCheckService, logger)

	api := router.Group("/api/v1")
	{
		// Gerrit Change endpoints
		api.GET("/changes", handlers.GetChanges)
		api.GET("/changes/:id", handlers.GetChange)
		api.POST("/changes/:id/sync", gerritSyncHandler.SyncChange)

		api.POST("/sync", gerritSyncHandler.Sync)
		api.POST("/sync/full", gerritSyncHandler.SyncFull) // Force full sync (30 days)
		api.GET("/sync/status", gerritSyncHandler.GetSyncStatus)

		api.GET("/updates/dashboard", gerritUpdateHandler.CheckDashboardUpdates)
		api.GET("/updates/change/:id", gerritUpdateHandler.CheckChangeUpdates)

		api.POST("/backup", backupHandler.Backup)
		api.POST("/clear", clearHandler.Clear)
	}

	return gerritSyncHandler
}
