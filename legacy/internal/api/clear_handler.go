package api

import (
	"net/http"

	"upgo/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ClearHandler struct {
	logger      *zap.Logger
	syncHandler *SyncHandler
}

func NewClearHandler(logger *zap.Logger, syncHandler *SyncHandler) *ClearHandler {
	return &ClearHandler{
		logger:      logger,
		syncHandler: syncHandler,
	}
}

func (h *ClearHandler) Clear(c *gin.Context) {
	// Require explicit confirmation to prevent accidental clears
	if c.GetHeader("X-Confirm-Clear") != "yes" {
		h.logger.Warn("データベースクリアの確認ヘッダーが不足しています")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "確認ヘッダーが必要です。X-Confirm-Clear: yes を設定してください",
		})
		return
	}

	// Wait for all in-flight sync operations to complete before clearing
	// This prevents race conditions and data corruption
	h.logger.Info("実行中の同期操作の完了を待機しています...")
	h.syncHandler.Wait()
	h.logger.Info("すべての同期操作が完了しました。データベースのクリアを開始します。")

	if err := database.ClearAllTables(h.logger); err != nil {
		h.logger.Error("データベースのクリアに失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "データベースのクリアに失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "データベースが正常にクリアされました",
	})
}
