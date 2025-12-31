package api

import (
	"net/http"
	"strconv"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SyncHandlers struct {
	syncService *service.GerritSyncService
	logger      *zap.Logger
}

func NewSyncHandlers(syncService *service.GerritSyncService, logger *zap.Logger) *SyncHandlers {
	return &SyncHandlers{
		syncService: syncService,
		logger:      logger,
	}
}

func (h *SyncHandlers) TriggerSync(c *gin.Context) {
	h.logger.Info("手動同期がトリガーされました")

	go func() {
		if err := h.syncService.Sync(c.Request.Context()); err != nil {
			h.logger.Error("同期に失敗しました", zap.Error(err))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "同期を開始しました",
		"status":  "accepted",
	})
}

func (h *SyncHandlers) SyncChange(c *gin.Context) {
	changeNumberStr := c.Param("change_number")
	changeNumber, err := strconv.Atoi(changeNumberStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効な変更番号です"})
		return
	}

	h.logger.Info("変更の同期がトリガーされました", zap.Int("change_number", changeNumber))

	go func() {
		if err := h.syncService.SyncChangeByNumber(c.Request.Context(), changeNumber); err != nil {
			h.logger.Error("変更の同期に失敗しました",
				zap.Int("change_number", changeNumber),
				zap.Error(err))
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message":       "変更の同期を開始しました",
		"change_number": changeNumber,
		"status":        "accepted",
	})
}

func (h *SyncHandlers) CheckUpdates(c *gin.Context) {
	hasUpdates, err := h.syncService.CheckUpdates(c.Request.Context())
	if err != nil {
		h.logger.Error("更新チェックに失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新チェックに失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_updates": hasUpdates,
	})
}
