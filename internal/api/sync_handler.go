package api

import (
	"context"
	"net/http"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SyncHandler handles HTTP requests for GitHub synchronization operations.
type SyncHandler struct {
	syncService *service.SyncService
	logger      *zap.Logger
}

// NewSyncHandler creates a new sync handler with the provided dependencies.
func NewSyncHandler(syncService *service.SyncService, logger *zap.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger,
	}
}

// Sync triggers an asynchronous synchronization of GitHub data.
// The operation runs in a background goroutine to avoid blocking the HTTP response,
// allowing the client to receive an immediate acknowledgment while the potentially
// long-running sync operation proceeds. A 10-minute timeout is enforced to prevent
// runaway operations.
func (h *SyncHandler) Sync(c *gin.Context) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		if err := h.syncService.Sync(ctx); err != nil {
			h.logger.Error("同期に失敗しました", zap.Error(err))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "同期を開始しました",
	})
}

// GetSyncStatus returns the current status of synchronization operations.
// TODO: Implement actual job status tracking. Currently returns a placeholder
// response. This requires adding a job queue or status store to track
// ongoing sync operations and their progress.
func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	// TODO: Get sync job status
	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
	})
}
