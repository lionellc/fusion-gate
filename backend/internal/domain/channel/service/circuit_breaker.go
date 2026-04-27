package service

import (
	"sync"
	"time"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
)

type CircuitBreaker interface {
	IsOpen() bool
	RecordSuccess()
	RecordFailure()
	RecordHalfOpenRequest()
	GetState() entity.CircuitState
	Reset()
}

// CircuitBreaker 独立的熔断器（用于内存缓存场景）
type circuitBreaker struct {
	mu               sync.RWMutex
	state            entity.CircuitState
	failureCount     int
	failureThreshold int
	recoveryInterval time.Duration
	lastFailureTime  time.Time
	lastSuccessTime  time.Time
	halfOpenRequests int
	halfOpenMax      int
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	FailureThreshold int           // 失败阈值（默认5）
	RecoveryInterval time.Duration // 恢复间隔（默认60秒）
	HalfOpenMax      int           // 半开状态最大试探请求（默认3）
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) CircuitBreaker {
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if config.RecoveryInterval == 0 {
		config.RecoveryInterval = 60 * time.Second
	}
	if config.HalfOpenMax == 0 {
		config.HalfOpenMax = 3
	}

	return &circuitBreaker{
		state:            entity.CircuitClosed,
		failureThreshold: config.FailureThreshold,
		recoveryInterval: config.RecoveryInterval,
		halfOpenMax:      config.HalfOpenMax,
	}
}

// IsOpen 检查是否熔断（拒绝请求）
func (cb *circuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == entity.CircuitClosed {
		return false // 正常，允许请求
	}

	if cb.state == entity.CircuitOpen {
		// 检查是否到达恢复时间
		if time.Since(cb.lastFailureTime) >= cb.recoveryInterval {
			// 转为半开状态
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = entity.CircuitHalfOpen
			cb.halfOpenRequests = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return false // 半开状态，允许试探请求
		}
		return true // 熔断状态，拒绝请求
	}

	// 半开状态，检查试探请求数量
	if cb.halfOpenRequests >= cb.halfOpenMax {
		return true // 已达最大试探请求，拒绝
	}

	return false // 半开状态，允许试探请求
}

// RecordSuccess 记录成功
func (cb *circuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount = 0
	cb.lastSuccessTime = time.Now()

	if cb.state == entity.CircuitHalfOpen {
		// 半开状态下成功，恢复正常
		cb.state = entity.CircuitClosed
		cb.halfOpenRequests = 0
	}
}

// RecordFailure 记录失败
func (cb *circuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == entity.CircuitHalfOpen {
		// 半开状态下失败，立即熔断
		cb.state = entity.CircuitOpen
		cb.halfOpenRequests = 0
	} else if cb.failureCount >= cb.failureThreshold {
		// 达到阈值，熔断
		cb.state = entity.CircuitOpen
	}
}

// RecordHalfOpenRequest 记录半开试探请求
func (cb *circuitBreaker) RecordHalfOpenRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == entity.CircuitHalfOpen {
		cb.halfOpenRequests++
	}
}

// GetState 获取当前状态
func (cb *circuitBreaker) GetState() entity.CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset 重置熔断器
func (cb *circuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = entity.CircuitClosed
	cb.failureCount = 0
	cb.halfOpenRequests = 0
}

// FromChannel 从Channel实体创建熔断器
func FromChannel(channel *entity.Channel) CircuitBreaker {
	return NewCircuitBreaker(CircuitBreakerConfig{
		FailureThreshold: channel.FailureThreshold,
		RecoveryInterval: time.Duration(channel.RecoveryInterval) * time.Second,
		HalfOpenMax:      3,
	})
}
