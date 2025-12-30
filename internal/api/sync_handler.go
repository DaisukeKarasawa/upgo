package api

import (
	"context"
	"net/http"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SyncHandler struct {
	syncService *service.SyncService
	logger      *zap.Logger
}

func NewSyncHandler(syncService *service.SyncService, logger *zap.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
		logger:      logger,
	}
}

func (h *SyncHandler) Sync(c *gin.Context) {
	go func() {
		// goroutine内でコンテキストを作成し、完了時にキャンセルする
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

func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	// TODO: 同期ジョブの状態を取得
	c.JSON(http.StatusOK, gin.H{
		"status": "completed",
	})
}
