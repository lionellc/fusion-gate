package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/infra/redis"
	"github.com/lionellc/fusion-gate/internal/middleware"
)

type Handler struct {
	userHandler *UserHandler
}

func NewHandler(userHandler *UserHandler) *Handler {
	return &Handler{userHandler: userHandler}
}

func NewRouter(handler *Handler, _ *redis.RedisClient) *gin.Engine {
	engine := gin.New()

	engine.Use(gin.Logger(), gin.Recovery())

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	userHandler := handler.userHandler
	// 认证接口（无需JWT）
	auth := engine.Group("/api/v1/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.Login)
	}

	// 用户接口（需要JWT）
	user := engine.Group("/api/v1/user")
	user.Use(middleware.JWTAuthMiddleware(userHandler.authService))
	{
		user.GET("/profile", userHandler.GetProfile)
	}

	return engine
}
