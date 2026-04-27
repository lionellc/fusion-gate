package persistence

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
	"github.com/lionellc/fusion-gate/internal/domain/channel/repo"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"gorm.io/gorm"
)

type channelRepo struct {
	db *db.DB
}

func NewChannelRepo(db *db.DB) repo.ChannelRepo {
	return &channelRepo{db: db}
}

func (r *channelRepo) Create(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Create(channel).Error
}

func (r *channelRepo) GetByID(ctx context.Context, id int64) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.WithContext(ctx).First(&channel, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &channel, err
}

func (r *channelRepo) GetByName(ctx context.Context, name string) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&channel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &channel, err
}

func (r *channelRepo) List(ctx context.Context) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.WithContext(ctx).Order("priority desc, weight desc").Find(&channels).Error
	return channels, err
}

func (r *channelRepo) ListByProvider(ctx context.Context, provider entity.Provider) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.WithContext(ctx).
		Where("provider = ?", provider).
		Order("priority desc, weight desc").
		Find(&channels).Error
	return channels, err
}

func (r *channelRepo) Update(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *channelRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.Channel{}, id).Error
}

// ListAvailable 获取支持某模型的所有可用渠道
func (r *channelRepo) ListAvailable(ctx context.Context, model string) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Where("models LIKE ?", "%"+model+"%").
		Where("circuit_state != ? OR circuit_state = ? AND last_failure_at < NOW() - INTERVAL '1 minute' * recovery_interval",
			entity.CircuitOpen, entity.CircuitOpen).
		Order("priority desc, weight desc").
		Find(&channels).Error
	return channels, err
}

// GetHighestPriority 获取支持某模型的最高优先级渠道
func (r *channelRepo) GetHighestPriority(ctx context.Context, model string) (*entity.Channel, error) {
	var channel entity.Channel
	err := r.db.WithContext(ctx).
		Where("status = ?", entity.StatusActive).
		Where("models LIKE ?", "%"+model+"%").
		Where("circuit_state = ? OR circuit_state = ?", entity.CircuitClosed, entity.CircuitHalfOpen).
		Order("priority desc, weight desc").
		First(&channel).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &channel, err
}

// UpdateHealth 更新健康状态
func (r *channelRepo) UpdateHealth(ctx context.Context, id int64, status entity.HealthStatus, responseTime int) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"health_status": status,
			"response_time": responseTime,
			"last_check_at": gorm.Expr("NOW()"),
		}).Error
}

// RecordSuccess 记录成功
func (r *channelRepo) RecordSuccess(ctx context.Context, id int64, responseTime int) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failure_count": 0,
			"circuit_state": entity.CircuitClosed,
			"health_status": entity.HealthHealthy,
			"response_time": responseTime,
			"status":        entity.StatusActive,
			"last_check_at": gorm.Expr("NOW()"),
		}).Error
}

// RecordFailure 记录失败
func (r *channelRepo) RecordFailure(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failure_count":   gorm.Expr("failure_count + 1"),
			"last_failure_at": gorm.Expr("NOW()"),
			"health_status":   entity.HealthUnhealthy,
		}).Error
}

// ResetCircuit 重置熔断器
func (r *channelRepo) ResetCircuit(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failure_count": 0,
			"circuit_state": entity.CircuitClosed,
			"status":        entity.StatusActive,
		}).Error
}

// 配额操作
func (r *channelRepo) GetQuota(ctx context.Context, id int64) (quota, used float64, err error) {
	channel, err := r.GetByID(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	return channel.Quota, channel.UsedQuota, nil
}

func (r *channelRepo) UseQuota(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Where("quota - used_quota >= ? OR unlimited = ?", amount, true).
		Update("used_quota", gorm.Expr("used_quota + ?", amount)).Error
}

func (r *channelRepo) RefundQuota(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.Channel{}).
		Where("id = ?", id).
		Where("used_quota >= ?", amount).
		Update("used_quota", gorm.Expr("used_quota - ?", amount)).Error
}
