package service

import (
	"context"
	"errors"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
	"github.com/lionellc/fusion-gate/internal/domain/channel/repo"
)

type ChannelService interface {
	Create(ctx context.Context, channel *entity.Channel) error
	GetByID(ctx context.Context, id int64) (*entity.Channel, error)
	List(ctx context.Context) ([]*entity.Channel, error)
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id int64) error
	Test(ctx context.Context, id int64) error
	Enable(ctx context.Context, id int64) error
	Disable(ctx context.Context, id int64) error
	ResetCircuit(ctx context.Context, id int64) error
	SelectChannel(ctx context.Context, model string) (*entity.Channel, error)
	GetStatus(ctx context.Context, id int64) (map[string]interface{}, error)
}

type channelService struct {
	channelRepo  repo.ChannelRepo
	healthCheck  HealthCheckService
	circuitCache CircuitBreakerCache
}

func NewChannelService(
	channelRepo repo.ChannelRepo,
	healthCheck HealthCheckService,
) ChannelService {
	return &channelService{
		channelRepo:  channelRepo,
		healthCheck:  healthCheck,
		circuitCache: NewCircuitBreakerCache(),
	}
}

// Create 创建渠道
func (s *channelService) Create(ctx context.Context, channel *entity.Channel) error {
	// 验证必填字段
	if channel.Name == "" {
		return errors.New("channel name is required")
	}
	if channel.Provider == "" {
		return errors.New("channel provider is required")
	}
	if channel.APIKey == "" {
		return errors.New("channel api key is required")
	}
	if channel.Models == "" {
		return errors.New("channel models is required")
	}

	// 检查名称是否重复
	existing, err := s.channelRepo.GetByName(ctx, channel.Name)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("channel name already exists")
	}

	// 设置默认值
	if channel.Status == "" {
		channel.Status = entity.StatusActive
	}
	if channel.HealthStatus == "" {
		channel.HealthStatus = entity.HealthUnknown
	}
	if channel.CircuitState == "" {
		channel.CircuitState = entity.CircuitClosed
	}
	if channel.FailureThreshold == 0 {
		channel.FailureThreshold = 5
	}
	if channel.RecoveryInterval == 0 {
		channel.RecoveryInterval = 60
	}
	if channel.Weight == 0 {
		channel.Weight = 100
	}

	// 创建渠道
	if err := s.channelRepo.Create(ctx, channel); err != nil {
		return err
	}

	// 立即执行健康检查
	go func() {
		checkCtx := context.Background()
		s.healthCheck.CheckChannel(checkCtx, channel)
	}()

	return nil
}

// GetByID 获取渠道
func (s *channelService) GetByID(ctx context.Context, id int64) (*entity.Channel, error) {
	return s.channelRepo.GetByID(ctx, id)
}

// List 获取渠道列表
func (s *channelService) List(ctx context.Context) ([]*entity.Channel, error) {
	return s.channelRepo.List(ctx)
}

// Update 更新渠道
func (s *channelService) Update(ctx context.Context, channel *entity.Channel) error {
	// 检查渠道是否存在
	existing, err := s.channelRepo.GetByID(ctx, channel.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("channel not found")
	}

	return s.channelRepo.Update(ctx, channel)
}

// Delete 删除渠道
func (s *channelService) Delete(ctx context.Context, id int64) error {
	// 清除熔断器缓存
	s.circuitCache.Remove(id)

	return s.channelRepo.Delete(ctx, id)
}

// Test 测试渠道（手动触发健康检查）
func (s *channelService) Test(ctx context.Context, id int64) error {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if channel == nil {
		return errors.New("channel not found")
	}

	return s.healthCheck.CheckChannel(ctx, channel)
}

// Enable 启用渠道
func (s *channelService) Enable(ctx context.Context, id int64) error {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if channel == nil {
		return errors.New("channel not found")
	}

	channel.Status = entity.StatusActive
	return s.channelRepo.Update(ctx, channel)
}

// Disable 禁用渠道
func (s *channelService) Disable(ctx context.Context, id int64) error {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if channel == nil {
		return errors.New("channel not found")
	}

	channel.Status = entity.StatusDisabled
	return s.channelRepo.Update(ctx, channel)
}

// ResetCircuit 重置熔断器
func (s *channelService) ResetCircuit(ctx context.Context, id int64) error {
	// 清除缓存
	s.circuitCache.Remove(id)

	// 重置数据库状态
	return s.channelRepo.ResetCircuit(ctx, id)
}

// SelectChannel 选择渠道（供Relay使用）
func (s *channelService) SelectChannel(ctx context.Context, model string) (*entity.Channel, error) {
	// 获取支持该模型且可用的最高优先级渠道
	channel, err := s.channelRepo.GetHighestPriority(ctx, model)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("no available channel for model: " + model)
	}

	// 检查熔断器缓存
	cb := s.circuitCache.Get(channel.ID, channel)
	if cb.IsOpen() {
		// 熔断器开启，尝试获取下一个渠道
		channels, err := s.channelRepo.ListAvailable(ctx, model)
		if err != nil {
			return nil, err
		}

		for _, ch := range channels {
			if ch.ID == channel.ID {
				continue // 跳过已熔断的渠道
			}
			cb := s.circuitCache.Get(ch.ID, ch)
			if !cb.IsOpen() {
				return ch, nil
			}
		}

		return nil, errors.New("all channels are fused for model: " + model)
	}

	return channel, nil
}

// GetStatus 获取渠道状态详情
func (s *channelService) GetStatus(ctx context.Context, id int64) (map[string]interface{}, error) {
	channel, err := s.channelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if channel == nil {
		return nil, errors.New("channel not found")
	}

	cb := s.circuitCache.Get(channel.ID, channel)

	return map[string]interface{}{
		"id":              channel.ID,
		"name":            channel.Name,
		"provider":        channel.Provider,
		"status":          channel.Status,
		"health_status":   channel.HealthStatus,
		"circuit_state":   cb.GetState(),
		"failure_count":   channel.FailureCount,
		"last_check_at":   channel.LastCheckAt,
		"last_failure_at": channel.LastFailureAt,
		"response_time":   channel.ResponseTime,
		"models":          channel.GetModelList(),
	}, nil
}
