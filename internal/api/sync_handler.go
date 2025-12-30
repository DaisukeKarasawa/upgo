package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SyncHandler handles HTTP requests for GitHub synchronization operations.
type SyncHandler struct {
	syncService *service.SyncService
	logger      *zap.Logger
	// semaphore limits the number of concurrent sync operations
	semaphore chan struct{}
	// maxConcurrency is the maximum number of concurrent sync operations
	maxConcurrency int
	// wg tracks running goroutines for graceful shutdown
	wg sync.WaitGroup
}

// NewSyncHandler creates a new sync handler with the provided dependencies.
// maxConcurrency limits the number of concurrent sync operations (default: 3).
func NewSyncHandler(syncService *service.SyncService, logger *zap.Logger, maxConcurrency int) *SyncHandler {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // default to 3 concurrent syncs
	}
	return &SyncHandler{
		syncService:    syncService,
		logger:         logger,
		semaphore:      make(chan struct{}, maxConcurrency),
		maxConcurrency: maxConcurrency,
	}
}

// Sync triggers an asynchronous synchronization of GitHub data.
// The operation runs in a background goroutine to avoid blocking the HTTP response,
// allowing the client to receive an immediate acknowledgment while the potentially
// long-running sync operation proceeds. A 10-minute timeout is enforced to prevent
// runaway operations. Concurrent sync operations are limited by maxConcurrency.
func (h *SyncHandler) Sync(c *gin.Context) {
	// Check if we can start a new sync operation
	select {
	case h.semaphore <- struct{}{}:
		// Acquired semaphore, proceed with sync
	default:
		// Too many concurrent syncs, reject request
		c.JSON(http.StatusTooManyRequests, gin.H{
			"message": "同時実行数の上限に達しています。しばらく待ってから再試行してください。",
		})
		return
	}

	h.wg.Add(1)
	go func() {
		// Create context inside goroutine so it lives for the operation duration
		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Minute)
		defer func() {
			cancel() // Cancel context when goroutine exits
			<-h.semaphore // Release semaphore
			h.wg.Done()
		}()

		if err := h.syncService.Sync(ctx); err != nil {
			if ctx.Err() == context.Canceled {
				h.logger.Info("同期がキャンセルされました")
			} else {
				h.logger.Error("同期に失敗しました", zap.Error(err))
			}
		} else {
			h.logger.Info("同期が完了しました")
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "同期を開始しました",
	})
}

// Wait blocks until all in-flight sync operations complete.
// This should be called during graceful shutdown.
func (h *SyncHandler) Wait() {
	h.wg.Wait()
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
