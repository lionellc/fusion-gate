package persistence

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/apikey/entity"
	"github.com/lionellc/fusion-gate/internal/domain/apikey/repo"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"gorm.io/gorm"
)

type apiKeyRepo struct {
	db *db.DB
}

func NewAPIKeyRepo(db *db.DB) repo.APIKeyRepo {
	return &apiKeyRepo{db: db}
}

func (r *apiKeyRepo) Create(ctx context.Context, apiKey *entity.APIKey) error {
	return r.db.WithContext(ctx).Create(apiKey).Error
}

func (r *apiKeyRepo) GetByID(ctx context.Context, id int64) (*entity.APIKey, error) {
	var apiKey entity.APIKey
	err := r.db.WithContext(ctx).First(&apiKey, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &apiKey, err
}

func (r *apiKeyRepo) GetByKey(ctx context.Context, key string) (*entity.APIKey, error) {
	var apiKey entity.APIKey
	err := r.db.WithContext(ctx).Where("key = ?", key).First(&apiKey).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &apiKey, err
}

func (r *apiKeyRepo) GetByUserID(ctx context.Context, userID int64) ([]*entity.APIKey, error) {
	var apiKeys []*entity.APIKey
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&apiKeys).Error
	return apiKeys, err
}

func (r *apiKeyRepo) Update(ctx context.Context, apiKey *entity.APIKey) error {
	return r.db.WithContext(ctx).Save(apiKey).Error
}

func (r *apiKeyRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&entity.APIKey{}, id).Error
}

func (r *apiKeyRepo) UseQuota(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.APIKey{}).
		Where("id = ?", id).
		Where("quota - used_quota >= ? OR unlimited = ?", amount, true).
		Update("used_quota", gorm.Expr("used_quota + ?", amount)).Error
}

func (r *apiKeyRepo) RefundQuota(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.APIKey{}).
		Where("id = ?", id).
		Where("used_quota >= ?", amount).
		Update("used_quota", gorm.Expr("used_quota - ?", amount)).Error
}

func (r *apiKeyRepo) GetQuota(ctx context.Context, id int64) (quota, used float64, err error) {
	apiKey, err := r.GetByID(ctx, id)
	if err != nil {
		return 0, 0, err
	}
	return apiKey.Quota, apiKey.UsedQuota, nil
}
