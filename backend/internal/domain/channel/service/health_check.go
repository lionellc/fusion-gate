package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
	"github.com/lionellc/fusion-gate/internal/domain/channel/repo"
)

type HealthCheckService interface {
	SetCheckFunc(fn func(ctx context.Context, channel *entity.Channel) error)
	CheckChannel(ctx context.Context, channel *entity.Channel) error
	CheckAllChannels(ctx context.Context) error
	CheckFusedChannels(ctx context.Context) error
	StartBackgroundCheck(ctx context.Context)
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Interval      time.Duration // 检查间隔（默认5分钟）
	Timeout       time.Duration // 单次检查超时（默认10秒）
	RecoveryCheck time.Duration // 熔断恢复检查间隔（默认30秒）
}

// HealthCheckService 健康检查服务
type healthCheckService struct {
	channelRepo repo.ChannelRepo
	config      HealthCheckConfig
	logger      *slog.Logger

	// 健康检查函数（外部提供）
	checkFunc func(ctx context.Context, channel *entity.Channel) error
}

func NewHealthCheckService(channelRepo repo.ChannelRepo, logger *slog.Logger) HealthCheckService {
	return &healthCheckService{
		channelRepo: channelRepo,
		config: HealthCheckConfig{
			Interval:      5 * time.Minute,
			Timeout:       10 * time.Second,
			RecoveryCheck: 30 * time.Second,
		},
		logger: logger,
	}
}

// SetCheckFunc 设置检查函数（由AdaptorFactory注入）
func (s *healthCheckService) SetCheckFunc(fn func(ctx context.Context, channel *entity.Channel) error) {
	s.checkFunc = fn
}

// CheckChannel 检查单个渠道
func (s *healthCheckService) CheckChannel(ctx context.Context, channel *entity.Channel) error {
	if s.checkFunc == nil {
		// 默认检查：简单HTTP请求
		return s.defaultCheck(ctx, channel)
	}

	checkCtx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	startTime := time.Now()
	err := s.checkFunc(checkCtx, channel)
	responseTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		s.logger.Warn("channel health check failed",
			"channel_id", channel.ID,
			"channel_name", channel.Name,
			"error", err)

		// 记录失败
		if err := s.channelRepo.RecordFailure(ctx, channel.ID); err != nil {
			s.logger.Error("failed to record failure", "error", err)
		}
		return err
	}

	// 记录成功
	if err := s.channelRepo.RecordSuccess(ctx, channel.ID, responseTime); err != nil {
		s.logger.Error("failed to record success", "error", err)
	}

	s.logger.Info("channel health check passed",
		"channel_id", channel.ID,
		"channel_name", channel.Name,
		"response_time", responseTime)

	return nil
}

// CheckAllChannels 检查所有渠道
func (s *healthCheckService) CheckAllChannels(ctx context.Context) error {
	channels, err := s.channelRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		if channel.Status == entity.StatusDisabled {
			continue // 跳过手动禁用的渠道
		}

		// 检查单个渠道
		go func(ch *entity.Channel) {
			checkCtx := context.Background()
			s.CheckChannel(checkCtx, ch)
		}(channel)
	}

	return nil
}

// CheckFusedChannels 检查熔断渠道（尝试恢复）
func (s *healthCheckService) CheckFusedChannels(ctx context.Context) error {
	channels, err := s.channelRepo.List(ctx)
	if err != nil {
		return err
	}

	for _, channel := range channels {
		// 只检查熔断状态的渠道
		if channel.CircuitState != entity.CircuitOpen {
			continue
		}

		// 检查是否到达恢复时间
		if time.Since(channel.LastFailureAt).Seconds() < float64(channel.RecoveryInterval) {
			continue
		}

		// 尝试恢复
		s.logger.Info("attempting to recover fused channel",
			"channel_id", channel.ID,
			"channel_name", channel.Name)

		// 设置为半开状态，允许试探性请求
		channel.CircuitState = entity.CircuitHalfOpen
		s.channelRepo.Update(ctx, channel)

		// 执行检查
		go func(ch *entity.Channel) {
			checkCtx := context.Background()
			s.CheckChannel(checkCtx, ch)
		}(channel)
	}

	return nil
}

// StartBackgroundCheck 启动后台健康检查
func (s *healthCheckService) StartBackgroundCheck(ctx context.Context) {
	// 定期检查所有渠道
	checkTicker := time.NewTicker(s.config.Interval)

	// 定期检查熔断渠道
	recoveryTicker := time.NewTicker(s.config.RecoveryCheck)

	go func() {
		for {
			select {
			case <-checkTicker.C:
				s.CheckAllChannels(ctx)
			case <-recoveryTicker.C:
				s.CheckFusedChannels(ctx)
			case <-ctx.Done():
				checkTicker.Stop()
				recoveryTicker.Stop()
				return
			}
		}
	}()
}

// defaultCheck 默认健康检查
func (s *healthCheckService) defaultCheck(ctx context.Context, channel *entity.Channel) error {
	// 下一步完成前，使用简单的HTTP HEAD请求检查
	// 这里只是占位，实际由Adaptor实现
	return nil
}
