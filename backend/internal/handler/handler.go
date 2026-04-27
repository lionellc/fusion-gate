package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/infra/redis"
	"github.com/lionellc/fusion-gate/internal/middleware"
)

type Handler struct {
	userHandler    *UserHandler
	apikeyHandler  *APIKeyHandler
	channelHandler *ChannelHandler
}

func NewHandler(userHandler *UserHandler, apikeyHandler *APIKeyHandler, channelHandler *ChannelHandler) *Handler {
	return &Handler{userHandler: userHandler, apikeyHandler: apikeyHandler, channelHandler: channelHandler}
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

	apikeyHandler := handler.apikeyHandler
	apikey := engine.Group("/api/v1/apikey")
	apikey.Use(middleware.JWTAuthMiddleware(userHandler.authService))
	{
		apikey.POST("/create", apikeyHandler.Create)
		apikey.GET("/list", apikeyHandler.List)
		apikey.DELETE("/:id", apikeyHandler.Delete)
	}

	// 渠道管理接口（需要JWT，管理员）
	channelHandler := handler.channelHandler
	channels := engine.Group("/api/v1/channels")
	channels.Use(middleware.JWTAuthMiddleware(userHandler.authService))
	{
		channels.POST("", channelHandler.Create)
		channels.GET("", channelHandler.List)
		channels.GET("/:id", channelHandler.Get)
		channels.PUT("/:id", channelHandler.Update)
		channels.DELETE("/:id", channelHandler.Delete)
		channels.POST("/:id/test", channelHandler.Test)
		channels.GET("/:id/status", channelHandler.Status)
		channels.POST("/:id/enable", channelHandler.Enable)
		channels.POST("/:id/disable", channelHandler.Disable)
		channels.POST("/:id/reset-circuit", channelHandler.ResetCircuit)
	}
	return engine
}
