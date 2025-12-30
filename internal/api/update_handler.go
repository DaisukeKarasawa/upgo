package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UpdateHandler handles HTTP requests for update check operations.
type UpdateHandler struct {
	updateCheckService *service.UpdateCheckService
	logger             *zap.Logger
}

// NewUpdateHandler creates a new update handler.
func NewUpdateHandler(updateCheckService *service.UpdateCheckService, logger *zap.Logger) *UpdateHandler {
	return &UpdateHandler{
		updateCheckService: updateCheckService,
		logger:             logger,
	}
}

// CheckDashboardUpdates checks for missing recent PRs (dashboard update check).
func (h *UpdateHandler) CheckDashboardUpdates(c *gin.Context) {
	// Use a detached context so a client disconnect doesn't cancel the shared update check.
	// (We still apply a timeout to avoid runaway requests.)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.updateCheckService.CheckDashboardUpdates(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Warn("ダッシュボード更新チェックがタイムアウトしました", zap.Error(err))
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "更新チェックがタイムアウトしました",
			})
			return
		}
		if errors.Is(err, context.Canceled) {
			h.logger.Info("ダッシュボード更新チェックがキャンセルされました", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "更新チェックがキャンセルされました",
			})
			return
		}
		h.logger.Error("ダッシュボード更新チェックに失敗しました", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新チェックに失敗しました",
		})
		return
	}
	if result == nil {
		// Guard: avoid panic if service unexpectedly returns (nil, nil)
		h.logger.Error("ダッシュボード更新チェック結果がnilです")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新チェックに失敗しました",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_missing_recent_prs": result.HasMissingRecentPRs,
		"missing_count":           result.MissingCount,
		"last_checked_at":         result.LastCheckedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// CheckPRUpdates checks if a specific PR has been updated since last sync.
func (h *UpdateHandler) CheckPRUpdates(c *gin.Context) {
	prIDStr := c.Param("id")
	prID, err := strconv.Atoi(prIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "無効なPR IDです",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := h.updateCheckService.CheckPRUpdates(ctx, prID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			h.logger.Warn("PR更新チェックがタイムアウトしました", zap.Int("pr_id", prID), zap.Error(err))
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error": "更新チェックがタイムアウトしました",
			})
			return
		}
		if errors.Is(err, context.Canceled) {
			h.logger.Info("PR更新チェックがキャンセルされました", zap.Int("pr_id", prID), zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "更新チェックがキャンセルされました",
			})
			return
		}
		h.logger.Error("PR更新チェックに失敗しました", zap.Int("pr_id", prID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新チェックに失敗しました",
		})
		return
	}

	response := gin.H{
		"updated_since_last_sync": result.UpdatedSinceLastSync,
		"last_checked_at":          result.LastCheckedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if result.LastSyncedAt != nil {
		response["last_synced_at"] = result.LastSyncedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		response["last_synced_at"] = nil
	}

	if result.GitHubUpdatedAt != nil {
		response["github_updated_at"] = result.GitHubUpdatedAt.Format("2006-01-02T15:04:05Z07:00")
	} else {
		response["github_updated_at"] = nil
	}

	c.JSON(http.StatusOK, response)
}
