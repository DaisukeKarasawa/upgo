package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"upgo/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GerritUpdateHandler handles update check requests for Gerrit Changes.
type GerritUpdateHandler struct {
	updateCheckService *service.GerritUpdateCheckService
	logger             *zap.Logger
}

// NewGerritUpdateHandler creates a new Gerrit update handler.
func NewGerritUpdateHandler(updateCheckService *service.GerritUpdateCheckService, logger *zap.Logger) *GerritUpdateHandler {
	return &GerritUpdateHandler{
		updateCheckService: updateCheckService,
		logger:             logger,
	}
}

// CheckDashboardUpdates checks if there are Changes created in the last month that don't exist in DB.
// Uses an independent context with timeout to avoid being canceled by client disconnection.
func (h *GerritUpdateHandler) CheckDashboardUpdates(c *gin.Context) {
	// Use independent context with timeout (45 seconds) instead of request context.
	// Increased timeout to handle slower Gerrit responses, especially with repo: filter.
	// This prevents the check from being canceled when the browser cancels the request.
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	result, err := h.updateCheckService.CheckDashboardUpdates(ctx)
	if err != nil {
		// On error, return cached result if available, otherwise return default (no updates).
		// This ensures the UI can still function even if Gerrit API is temporarily unavailable.
		h.logger.Warn("ダッシュボード更新チェックに失敗しました（キャッシュまたはデフォルト値を返します）", zap.Error(err))
		
		// Try to get cached result as fallback
		cachedResult := h.updateCheckService.GetCachedDashboardResult()
		if cachedResult != nil {
			h.logger.Info("キャッシュされた結果を返します", 
				zap.Bool("has_missing", cachedResult.HasMissingRecentChanges),
				zap.Int("missing_count", cachedResult.MissingCount))
			c.JSON(http.StatusOK, cachedResult)
			return
		}
		
		// No cache available, return default (no updates) to clear the badge
		h.logger.Info("キャッシュなし、デフォルト値（更新なし）を返します")
		c.JSON(http.StatusOK, &service.GerritDashboardUpdateResult{
			HasMissingRecentChanges: false,
			MissingCount:             0,
			LastCheckedAt:            time.Now(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckChangeUpdates checks if a specific Change has been updated on Gerrit since its last_synced_at.
func (h *GerritUpdateHandler) CheckChangeUpdates(c *gin.Context) {
	changeIDStr := c.Param("id")
	changeID, err := strconv.Atoi(changeIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なChange IDです"})
		return
	}

	result, err := h.updateCheckService.CheckChangeUpdates(c.Request.Context(), changeID)
	if err != nil {
		h.logger.Error("Change更新チェックに失敗しました", zap.Int("change_id", changeID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新チェックに失敗しました"})
		return
	}

	c.JSON(http.StatusOK, result)
}
