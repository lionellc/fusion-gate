package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"github.com/lionellc/fusion-gate/internal/infra/redis"
)

func NewRouter(_ *db.DB, _ *redis.RedisClient) *gin.Engine {
	engine := gin.New()

	engine.Use(gin.Logger(), gin.Recovery())

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	return engine
}
