package api

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GerritSyncHandler handles HTTP requests for Gerrit synchronization operations.
type GerritSyncHandler struct {
	syncService *service.GerritSyncService
	logger      *zap.Logger
	// semaphore limits the number of concurrent sync operations
	semaphore chan struct{}
	// maxConcurrency is the maximum number of concurrent sync operations
	maxConcurrency int
	// wg tracks running goroutines for graceful shutdown
	wg sync.WaitGroup
}

// NewGerritSyncHandler creates a new Gerrit sync handler with the provided dependencies.
// maxConcurrency limits the number of concurrent sync operations (default: 3).
func NewGerritSyncHandler(syncService *service.GerritSyncService, logger *zap.Logger, maxConcurrency int) *GerritSyncHandler {
	if maxConcurrency <= 0 {
		maxConcurrency = 3 // default to 3 concurrent syncs
	}
	return &GerritSyncHandler{
		syncService:    syncService,
		logger:         logger,
		semaphore:      make(chan struct{}, maxConcurrency),
		maxConcurrency: maxConcurrency,
	}
}

// Sync triggers an asynchronous synchronization of Gerrit data.
func (h *GerritSyncHandler) Sync(c *gin.Context) {
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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer func() {
			cancel()
			<-h.semaphore
			h.wg.Done()
		}()

		if err := h.syncService.Sync(ctx); err != nil {
			switch ctx.Err() {
			case context.Canceled:
				h.logger.Info("同期がキャンセルされました")
			case context.DeadlineExceeded:
				h.logger.Warn("同期がタイムアウトしました", zap.Error(err))
			default:
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

// SyncFull triggers a full synchronization (30 days) regardless of last_synced_at.
func (h *GerritSyncHandler) SyncFull(c *gin.Context) {
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
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute) // Longer timeout for full sync
		defer func() {
			cancel()
			<-h.semaphore
			h.wg.Done()
		}()

		// Use SyncWithOptions with forceFullSync=true to force 30-day sync
		if err := h.syncService.SyncWithOptions(ctx, true); err != nil {
			switch ctx.Err() {
			case context.Canceled:
				h.logger.Info("フル同期がキャンセルされました")
			case context.DeadlineExceeded:
				h.logger.Warn("フル同期がタイムアウトしました", zap.Error(err))
			default:
				h.logger.Error("フル同期に失敗しました", zap.Error(err))
			}
		} else {
			h.logger.Info("フル同期が完了しました")
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "フル同期を開始しました（30日分）",
	})
}

// SyncChange triggers an asynchronous synchronization of a single Change by its database ID.
func (h *GerritSyncHandler) SyncChange(c *gin.Context) {
	changeIDStr := c.Param("id")
	changeID, err := strconv.Atoi(changeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "無効なChange IDです",
		})
		return
	}

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
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer func() {
			cancel()
			<-h.semaphore
			h.wg.Done()
		}()

		if err := h.syncService.SyncChangeByID(ctx, changeID); err != nil {
			switch ctx.Err() {
			case context.Canceled:
				h.logger.Info("Change同期がキャンセルされました", zap.Int("change_id", changeID))
			case context.DeadlineExceeded:
				h.logger.Warn("Change同期がタイムアウトしました", zap.Int("change_id", changeID), zap.Error(err))
			default:
				h.logger.Error("Change同期に失敗しました", zap.Int("change_id", changeID), zap.Error(err))
			}
		} else {
			h.logger.Info("Change同期が完了しました", zap.Int("change_id", changeID))
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Change同期を開始しました",
	})
}

// Wait blocks until all in-flight sync operations complete.
func (h *GerritSyncHandler) Wait() {
	h.wg.Wait()
}

// GetSyncStatus returns the current status of synchronization operations.
func (h *GerritSyncHandler) GetSyncStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
	})
}
