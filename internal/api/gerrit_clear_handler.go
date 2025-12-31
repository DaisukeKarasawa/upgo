package api

import (
	"net/http"

	"upgo/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GerritClearHandler handles database clearing requests.
type GerritClearHandler struct {
	logger          *zap.Logger
	gerritSyncHandler *GerritSyncHandler
}

// NewGerritClearHandler creates a new clear handler.
func NewGerritClearHandler(logger *zap.Logger, gerritSyncHandler *GerritSyncHandler) *GerritClearHandler {
	return &GerritClearHandler{
		logger:            logger,
		gerritSyncHandler: gerritSyncHandler,
	}
}

// Clear clears all tables in the database.
func (h *GerritClearHandler) Clear(c *gin.Context) {
	// Require confirmation header
	confirm := c.GetHeader("X-Confirm-Clear")
	if confirm != "yes" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "確認ヘッダーが必要です。X-Confirm-Clear: yes を設定してください。",
		})
		return
	}

	// Wait for all sync operations to complete
	h.gerritSyncHandler.Wait()

	// Clear all tables
	if err := database.ClearAllTables(h.logger); err != nil {
		h.logger.Error("データベースのクリアに失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "データベースのクリアに失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "データベースがクリアされました",
	})
}
