package repo

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/apikey/entity"
)

// APIKeyRepo API Key仓储接口
type APIKeyRepo interface {
	// CRUD
	Create(ctx context.Context, apiKey *entity.APIKey) error
	GetByID(ctx context.Context, id int64) (*entity.APIKey, error)
	GetByKey(ctx context.Context, key string) (*entity.APIKey, error) // 根据sk-xxx查询
	GetByUserID(ctx context.Context, userID int64) ([]*entity.APIKey, error)
	Update(ctx context.Context, apiKey *entity.APIKey) error
	Delete(ctx context.Context, id int64) error

	// 额度操作（用于计费）
	UseQuota(ctx context.Context, id int64, amount float64) error
	RefundQuota(ctx context.Context, id int64, amount float64) error
	GetQuota(ctx context.Context, id int64) (quota, used float64, err error)
}
