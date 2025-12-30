package api

import (
	"database/sql"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"upgo/internal/service"
)

func SetupRoutes(router *gin.Engine, db *sql.DB, syncService *service.SyncService, logger *zap.Logger) {
	handlers := NewHandlers(db, logger)
	syncHandler := NewSyncHandler(syncService, logger)

	// ヘルスチェックはmain.goで設定済み

	api := router.Group("/api/v1")
	{
		// PR関連
		api.GET("/prs", handlers.GetPRs)
		api.GET("/prs/:id", handlers.GetPR)

		// Issue関連
		api.GET("/issues", handlers.GetIssues)
		api.GET("/issues/:id", handlers.GetIssue)

		// 同期関連
		api.POST("/sync", syncHandler.Sync)
		api.GET("/sync/status", syncHandler.GetSyncStatus)
	}
}
