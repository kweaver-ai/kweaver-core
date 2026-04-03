package handler

import (
	"context"
	"fmt"
	"net/http"
	"oss-gateway/internal/cache"
	"oss-gateway/pkg/errors"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *cache.RedisClient
}

func NewHealthHandler(db *gorm.DB, redis *cache.RedisClient) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Alive(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"res":    0,
		"status": "ok",
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := make(map[string]string)
	allHealthy := true

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "failed"
		allHealthy = false
	} else {
		if err := sqlDB.Ping(); err != nil {
			checks["database"] = "failed"
			allHealthy = false
		} else {
			checks["database"] = "ok"
		}
	}

	// Check Redis
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.redis.Client().Ping(ctx).Err(); err != nil {
		checks["redis"] = "failed"
		allHealthy = false
	} else {
		checks["redis"] = fmt.Sprintf("ok (%s)", h.redis.GetClusterMode())
	}

	if allHealthy {
		c.JSON(http.StatusOK, gin.H{
			"res":    0,
			"status": "ok",
			"checks": checks,
		})
	} else {
		code := errors.ServiceNotReady
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":        code.Code,
			"message":     code.Message,
			"description": code.Description,
			"checks":      checks,
		})
	}
}
