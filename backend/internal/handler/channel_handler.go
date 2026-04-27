package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
	"github.com/lionellc/fusion-gate/internal/domain/channel/service"
)

type ChannelHandler struct {
	channelService service.ChannelService
}

func NewChannelHandler(channelService service.ChannelService) *ChannelHandler {
	return &ChannelHandler{channelService: channelService}
}

// Create 创建渠道
func (h *ChannelHandler) Create(c *gin.Context) {
	var req struct {
		Name             string   `json:"name" binding:"required"`
		Provider         string   `json:"provider" binding:"required"`
		BaseURL          string   `json:"base_url"`
		APIKey           string   `json:"api_key" binding:"required"`
		OrgID            string   `json:"org_id"`
		Models           []string `json:"models" binding:"required"`
		Priority         int      `json:"priority"`
		Weight           int      `json:"weight"`
		FailureThreshold int      `json:"failure_threshold"`
		RecoveryInterval int      `json:"recovery_interval"`
		Quota            float64  `json:"quota"`
		Unlimited        bool     `json:"unlimited"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证Provider
	provider := entity.Provider(req.Provider)
	validProviders := []entity.Provider{
		entity.ProviderOpenAI,
		entity.ProviderClaude,
		entity.ProviderGemini,
		entity.ProviderCustom,
	}
	valid := false
	for _, p := range validProviders {
		if provider == p {
			valid = true
			break
		}
	}
	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid provider"})
		return
	}

	channel := &entity.Channel{
		Name:             req.Name,
		Provider:         provider,
		BaseURL:          req.BaseURL,
		APIKey:           req.APIKey,
		OrgID:            req.OrgID,
		Models:           strings.Join(req.Models, ","),
		Priority:         req.Priority,
		Weight:           req.Weight,
		FailureThreshold: req.FailureThreshold,
		RecoveryInterval: req.RecoveryInterval,
		Quota:            req.Quota,
		Unlimited:        req.Unlimited,
	}

	if err := h.channelService.Create(c.Request.Context(), channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       channel.ID,
		"name":     channel.Name,
		"provider": channel.Provider,
		"models":   channel.GetModelList(),
		"status":   channel.Status,
	})
}

// List 获取渠道列表
func (h *ChannelHandler) List(c *gin.Context) {
	channels, err := h.channelService.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 构建响应（隐藏APIKey）
	result := make([]gin.H, 0, len(channels))
	for _, ch := range channels {
		result = append(result, gin.H{
			"id":            ch.ID,
			"name":          ch.Name,
			"provider":      ch.Provider,
			"models":        ch.GetModelList(),
			"status":        ch.Status,
			"health_status": ch.HealthStatus,
			"circuit_state": ch.CircuitState,
			"priority":      ch.Priority,
			"weight":        ch.Weight,
			"response_time": ch.ResponseTime,
			"last_check_at": ch.LastCheckAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"channels": result})
}

// Get 获取单个渠道
func (h *ChannelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":              channel.ID,
		"name":            channel.Name,
		"provider":        channel.Provider,
		"base_url":        channel.BaseURL,
		"models":          channel.GetModelList(),
		"status":          channel.Status,
		"health_status":   channel.HealthStatus,
		"circuit_state":   channel.CircuitState,
		"priority":        channel.Priority,
		"weight":          channel.Weight,
		"response_time":   channel.ResponseTime,
		"failure_count":   channel.FailureCount,
		"last_check_at":   channel.LastCheckAt,
		"last_failure_at": channel.LastFailureAt,
	})
}

// Update 更新渠道
func (h *ChannelHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	channel, err := h.channelService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if channel == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		return
	}

	var req struct {
		Name             string   `json:"name"`
		BaseURL          string   `json:"base_url"`
		APIKey           string   `json:"api_key"`
		Models           []string `json:"models"`
		Priority         int      `json:"priority"`
		Weight           int      `json:"weight"`
		FailureThreshold int      `json:"failure_threshold"`
		RecoveryInterval int      `json:"recovery_interval"`
		Quota            float64  `json:"quota"`
		Unlimited        bool     `json:"unlimited"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新字段
	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.BaseURL != "" {
		channel.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		channel.APIKey = req.APIKey
	}
	if len(req.Models) > 0 {
		channel.Models = strings.Join(req.Models, ",")
	}
	if req.Priority > 0 {
		channel.Priority = req.Priority
	}
	if req.Weight > 0 {
		channel.Weight = req.Weight
	}
	if req.FailureThreshold > 0 {
		channel.FailureThreshold = req.FailureThreshold
	}
	if req.RecoveryInterval > 0 {
		channel.RecoveryInterval = req.RecoveryInterval
	}
	if req.Quota > 0 {
		channel.Quota = req.Quota
	}
	channel.Unlimited = req.Unlimited

	if err := h.channelService.Update(c.Request.Context(), channel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// Delete 删除渠道
func (h *ChannelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.channelService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Test 测试渠道
func (h *ChannelHandler) Test(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.channelService.Test(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 获取更新后的状态
	status, err := h.channelService.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"health_status": status["health_status"],
		"response_time": status["response_time"],
	})
}

// Status 获取渠道状态详情
func (h *ChannelHandler) Status(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	status, err := h.channelService.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

// Enable 启用渠道
func (h *ChannelHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.channelService.Enable(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "enabled"})
}

// Disable 禁用渠道
func (h *ChannelHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.channelService.Disable(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disabled"})
}

// ResetCircuit 重置熔断器
func (h *ChannelHandler) ResetCircuit(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.channelService.ResetCircuit(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "circuit reset"})
}
