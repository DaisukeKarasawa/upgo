package api

import (
	"net/http"

	"upgo/internal/config"
	"upgo/internal/database"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type BackupHandler struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewBackupHandler(cfg *config.Config, logger *zap.Logger) *BackupHandler {
	return &BackupHandler{
		cfg:    cfg,
		logger: logger,
	}
}

func (h *BackupHandler) Backup(c *gin.Context) {
	if !h.cfg.Backup.Enabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "バックアップ機能が無効になっています",
		})
		return
	}

	if h.cfg.Backup.Path == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "バックアップパスが設定されていません",
		})
		return
	}

	if err := database.Backup(
		h.cfg.Backup.Path,
		h.cfg.Database.Path,
		h.cfg.Backup.MaxBackups,
		h.logger,
	); err != nil {
		h.logger.Error("バックアップに失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "バックアップに失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "バックアップが完了しました",
	})
}
