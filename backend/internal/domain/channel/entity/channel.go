package entity

import (
	"strings"
	"time"
)

// Provider 渠道提供商类型
type Provider string

const (
	ProviderOpenAI Provider = "openai"
	ProviderClaude Provider = "claude"
	ProviderGemini Provider = "gemini"
	ProviderCustom Provider = "custom"
)

// ChannelStatus 渠道状态
type ChannelStatus string

const (
	StatusActive   ChannelStatus = "active"   // 正常可用
	StatusDisabled ChannelStatus = "disabled" // 手动禁用
	StatusFailed   ChannelStatus = "failed"   // 健康检查失败
	StatusFused    ChannelStatus = "fused"    // 熔断状态
)

// HealthStatus 健康检查状态
type HealthStatus string

const (
	HealthHealthy   HealthStatus = "healthy"   // 健康
	HealthUnhealthy HealthStatus = "unhealthy" // 不健康
	HealthUnknown   HealthStatus = "unknown"   // 未检查
)

// CircuitState 熔断器状态
type CircuitState string

const (
	CircuitClosed   CircuitState = "closed"    // 正常（允许请求）
	CircuitOpen     CircuitState = "open"      // 熔断（拒绝请求）
	CircuitHalfOpen CircuitState = "half-open" // 半开（试探性允许）
)

// Channel 渠道实体 - 代表一个API提供商的配置
type Channel struct {
	ID       int64         `gorm:"primaryKey" json:"id"`
	Name     string        `gorm:"size:100;not null" json:"name"`        // 渠道名称
	Provider Provider      `gorm:"size:20;not null" json:"provider"`     // 提供商类型
	Status   ChannelStatus `gorm:"size:20;default:active" json:"status"` // 渠道状态

	// API配置
	BaseURL string `gorm:"size:200" json:"base_url"`   // API地址（可选覆盖）
	APIKey  string `gorm:"size:200;not null" json:"-"` // API密钥（不暴露）
	OrgID   string `gorm:"size:100" json:"org_id"`     // 组织ID（可选）

	// 支持的模型
	Models string `gorm:"size:500;not null" json:"models"` // 支持的模型列表（逗号分隔）

	// 优先级与权重
	Priority int `gorm:"default:0" json:"priority"` // 优先级（越高越优先）
	Weight   int `gorm:"default:100" json:"weight"` // 权重（负载均衡用）

	// 健康检查
	HealthStatus HealthStatus `gorm:"size:20;default:unknown" json:"health_status"`
	LastCheckAt  time.Time    `gorm:"" json:"last_check_at"`          // 上次检查时间
	ResponseTime int          `gorm:"default:0" json:"response_time"` // 响应时间(ms)

	// 熔断器
	CircuitState     CircuitState `gorm:"size:20;default:closed" json:"circuit_state"`
	FailureCount     int          `gorm:"default:0" json:"failure_count"`      // 连续失败次数
	FailureThreshold int          `gorm:"default:5" json:"failure_threshold"`  // 熔断阈值
	RecoveryInterval int          `gorm:"default:60" json:"recovery_interval"` // 恢复间隔(秒)
	LastFailureAt    time.Time    `gorm:"" json:"last_failure_at"`             // 上次失败时间

	// 配额管理（可选）
	Quota     float64 `gorm:"type:decimal(10,4);default:0" json:"quota"`      // 配额
	UsedQuota float64 `gorm:"type:decimal(10,4);default:0" json:"used_quota"` // 已用配额
	Unlimited bool    `gorm:"default:true" json:"unlimited"`                  // 无限制

	// 时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// IsAvailable 检查渠道是否可用（考虑状态和熔断）
func (c *Channel) IsAvailable() bool {
	if c.Status != StatusActive {
		return false
	}
	// 熔断状态检查
	if c.CircuitState == CircuitOpen {
		// 检查是否到达恢复时间
		if time.Since(c.LastFailureAt).Seconds() < float64(c.RecoveryInterval) {
			return false
		}
		// 到达恢复时间，转为半开状态
		c.CircuitState = CircuitHalfOpen
	}
	return true
}

// RecordSuccess 记录成功（重置熔断计数）
func (c *Channel) RecordSuccess(responseTime int) {
	c.FailureCount = 0
	c.CircuitState = CircuitClosed
	c.HealthStatus = HealthHealthy
	c.ResponseTime = responseTime
	c.LastCheckAt = time.Now()
}

// RecordFailure 记录失败（累加熔断计数）
func (c *Channel) RecordFailure() {
	c.FailureCount++
	c.LastFailureAt = time.Now()
	c.HealthStatus = HealthUnhealthy

	if c.FailureCount >= c.FailureThreshold {
		c.CircuitState = CircuitOpen
		c.Status = StatusFused
	}
}

// GetModelList 获取模型列表
func (c *Channel) GetModelList() []string {
	if c.Models == "" {
		return []string{}
	}
	// 分割逗号分隔的模型列表
	models := strings.Split(c.Models, ",")
	result := make([]string, 0, len(models))
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m != "" {
			result = append(result, m)
		}
	}
	return result
}

// SupportsModel 检查是否支持某模型
func (c *Channel) SupportsModel(model string) bool {
	models := c.GetModelList()
	for _, m := range models {
		if m == model {
			return true
		}
	}
	return false
}
