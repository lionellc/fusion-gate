package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/domain/apikey/service"
)

func APIKeyAuthMiddleware(apiKeyService service.APIKeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization"})
			c.Abort()
			return
		}

		// API Key格式：Bearer sk-xxx 或直接 sk-xxx
		key := strings.TrimPrefix(authHeader, "Bearer ")
		if key == authHeader && !strings.HasPrefix(authHeader, "sk-") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid format"})
			c.Abort()
			return
		}

		apiKey, err := apiKeyService.Validate(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set("api_key_id", apiKey.ID)
		c.Set("user_id", apiKey.UserID)
		c.Set("api_key", apiKey.Key)
		c.Next()
	}
}
