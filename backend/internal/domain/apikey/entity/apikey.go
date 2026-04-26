package entity

import "time"

// APIKeyStatus API Key状态
type APIKeyStatus string

const (
	APIKeyActive   APIKeyStatus = "active"   // 可用
	APIKeyDisabled APIKeyStatus = "disabled" // 禁用
	APIKeyExpired  APIKeyStatus = "expired"  // 过期
)

// APIKey API密钥实体 - 用于API调用认证
// 区别于JWT Token：API Key是长期凭证，用于调用/v1/chat/completions等服务
type APIKey struct {
	ID     int64        `gorm:"primaryKey" json:"id"`
	Key    string       `gorm:"size:50;uniqueIndex;not null" json:"key"` // sk-xxx格式
	Name   string       `gorm:"size:100" json:"name"`                    // API Key名称
	UserID int64        `gorm:"not null" json:"user_id"`                 // 所属用户
	Status APIKeyStatus `gorm:"size:20;default:active" json:"status"`    // 状态

	// 额度管理
	Quota     float64 `gorm:"type:decimal(10,4);default:0" json:"quota"`      // 总额度（美元）
	UsedQuota float64 `gorm:"type:decimal(10,4);default:0" json:"used_quota"` // 已用额度
	Unlimited bool    `gorm:"default:false" json:"unlimited"`                 // 无限制额度

	// 权限控制
	Models string `gorm:"size:500" json:"models"` // 可用模型列表（逗号分隔）

	// 时间
	ExpiresAt time.Time `gorm:"" json:"expires_at"` // 过期时间
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// IsAvailable 检查API Key是否可用
func (k *APIKey) IsAvailable() bool {
	if k.Status != APIKeyActive {
		return false
	}
	if !k.ExpiresAt.IsZero() && time.Now().After(k.ExpiresAt) {
		return false
	}
	return true
}

// HasQuota 检查是否有足够额度
func (k *APIKey) HasQuota(amount float64) bool {
	if k.Unlimited {
		return true
	}
	return k.Quota-k.UsedQuota >= amount
}

// UseQuota 使用额度
func (k *APIKey) UseQuota(amount float64) {
	if !k.Unlimited {
		k.UsedQuota += amount
	}
}

// RefundQuota 退还额度
func (k *APIKey) RefundQuota(amount float64) {
	if !k.Unlimited {
		k.UsedQuota -= amount
		if k.UsedQuota < 0 {
			k.UsedQuota = 0
		}
	}
}

// RemainingQuota 剩余额度
func (k *APIKey) RemainingQuota() float64 {
	if k.Unlimited {
		return -1 // 表示无限
	}
	return k.Quota - k.UsedQuota
}
