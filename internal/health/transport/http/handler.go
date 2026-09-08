package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/railzwaylabs/github-actions-samples/internal/platform/config"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Module = fx.Module("health.http", fx.Invoke(Register))

type Handler struct {
	db      *gorm.DB
	log     *zap.Logger
	version string
}

func Register(router *gin.Engine, db *gorm.DB, log *zap.Logger, cfg config.Config) {
	handler := &Handler{db: db, log: log, version: cfg.Version}
	router.GET("/health", handler.Health)
	router.GET("/ready", handler.Ready)
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok", "version": h.version})
}

func (h *Handler) Ready(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		h.log.Error("readiness database handle failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		h.log.Warn("readiness database ping failed", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready", "database": "ok", "version": h.version})
}
