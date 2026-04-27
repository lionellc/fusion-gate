package repo

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/channel/entity"
)

type ChannelRepo interface {
	// CRUD
	Create(ctx context.Context, channel *entity.Channel) error
	GetByID(ctx context.Context, id int64) (*entity.Channel, error)
	GetByName(ctx context.Context, name string) (*entity.Channel, error)
	List(ctx context.Context) ([]*entity.Channel, error)
	ListByProvider(ctx context.Context, provider entity.Provider) ([]*entity.Channel, error)
	Update(ctx context.Context, channel *entity.Channel) error
	Delete(ctx context.Context, id int64) error

	// 状态查询
	ListAvailable(ctx context.Context, model string) ([]*entity.Channel, error)
	GetHighestPriority(ctx context.Context, model string) (*entity.Channel, error)

	// 健康检查与熔断
	UpdateHealth(ctx context.Context, id int64, status entity.HealthStatus, responseTime int) error
	RecordSuccess(ctx context.Context, id int64, responseTime int) error
	RecordFailure(ctx context.Context, id int64) error
	ResetCircuit(ctx context.Context, id int64) error

	// 配额操作
	GetQuota(ctx context.Context, id int64) (quota, used float64, err error)
	UseQuota(ctx context.Context, id int64, amount float64) error
	RefundQuota(ctx context.Context, id int64, amount float64) error
}
