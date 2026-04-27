package service

import (
	"sync"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
)

type CircuitBreakerCache interface {
	Get(channelID int64, channel *entity.Channel) CircuitBreaker
	Remove(channelID int64)
	Clear()
}

// CircuitBreakerCache 熔断器缓存（避免频繁查DB）
type circuitBreakerCache struct {
	mu       sync.RWMutex
	breakers map[int64]CircuitBreaker
}

func NewCircuitBreakerCache() CircuitBreakerCache {
	return &circuitBreakerCache{
		breakers: make(map[int64]CircuitBreaker),
	}
}

// Get 获取渠道的熔断器
func (c *circuitBreakerCache) Get(channelID int64, channel *entity.Channel) CircuitBreaker {
	c.mu.RLock()
	cb, exists := c.breakers[channelID]
	c.mu.RUnlock()

	if exists {
		return cb
	}

	// 创建新熔断器
	cb = FromChannel(channel)

	c.mu.Lock()
	c.breakers[channelID] = cb
	c.mu.Unlock()

	return cb
}

// Remove 移除熔断器
func (c *circuitBreakerCache) Remove(channelID int64) {
	c.mu.Lock()
	delete(c.breakers, channelID)
	c.mu.Unlock()
}

// Clear 清空缓存
func (c *circuitBreakerCache) Clear() {
	c.mu.Lock()
	c.breakers = make(map[int64]CircuitBreaker)
	c.mu.Unlock()
}
