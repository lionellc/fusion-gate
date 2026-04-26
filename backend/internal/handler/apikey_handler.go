package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/constant"
	"github.com/lionellc/fusion-gate/internal/domain/apikey/service"
	"github.com/lionellc/fusion-gate/internal/types"
)

type APIKeyHandler struct {
	apiKeyService service.APIKeyService
}

func NewAPIKeyHandler(apiKeyService service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

// Create 创建API Key（使用JWT认证的管理接口）
func (h *APIKeyHandler) Create(c *gin.Context) {
	userID := c.GetInt64(constant.HeaderXUserId)

	var req types.CreateAPIKeyReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var expiresAt time.Time
	if req.ExpiresAt != "" {
		expiresAt, _ = time.Parse("2006-01-02", req.ExpiresAt)
	}

	apiKey, err := h.apiKeyService.Create(c.Request.Context(), userID, req.Name, req.Quota, req.Unlimited, req.Models, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.CreateAPIKeyResp{
		ID:        apiKey.ID,
		Key:       apiKey.Key,
		Name:      apiKey.Name,
		Quota:     apiKey.Quota,
		Unlimited: apiKey.Unlimited,
		Models:    apiKey.Models,
		ExpiresAt: apiKey.ExpiresAt.Format("2006-01-02 15:04:05"),
	})
}

// List 获取API Key列表
func (h *APIKeyHandler) List(c *gin.Context) {
	userID := c.GetInt64(constant.HeaderXUserId)

	apiKeys, err := h.apiKeyService.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": apiKeys})
}

// Delete 删除API Key
func (h *APIKeyHandler) Delete(c *gin.Context) {
	userID := c.GetInt64(constant.HeaderXUserId)
	apiKeyID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.apiKeyService.Delete(c.Request.Context(), apiKeyID, userID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
