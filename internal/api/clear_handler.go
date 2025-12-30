package api

import (
	"net/http"

	"upgo/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ClearHandler struct {
	logger *zap.Logger
}

func NewClearHandler(logger *zap.Logger) *ClearHandler {
	return &ClearHandler{
		logger: logger,
	}
}

func (h *ClearHandler) Clear(c *gin.Context) {
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
