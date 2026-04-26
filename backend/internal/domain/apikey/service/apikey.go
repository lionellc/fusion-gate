package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/lionellc/fusion-gate/internal/domain/apikey/entity"
	"github.com/lionellc/fusion-gate/internal/domain/apikey/repo"
)

type APIKeyService interface {
	Create(ctx context.Context, userID int64, name string, quota float64, unlimited bool, models string, expiresAt time.Time) (*entity.APIKey, error)
	Validate(ctx context.Context, key string) (*entity.APIKey, error)
	GetByUserID(ctx context.Context, userID int64) ([]*entity.APIKey, error)
	Delete(ctx context.Context, id int64, userID int64) error
	CheckQuota(ctx context.Context, id int64, amount float64) (bool, error)
	UseQuota(ctx context.Context, id int64, amount float64) error
	RefundQuota(ctx context.Context, id int64, amount float64) error
}

type apiKeyService struct {
	apiKeyClient repo.APIKeyRepo
}

func NewAPIKeyService(apiKeyClient repo.APIKeyRepo) APIKeyService {
	return &apiKeyService{apiKeyClient: apiKeyClient}
}

// Create 创建API Key
func (s *apiKeyService) Create(ctx context.Context, userID int64, name string, quota float64, unlimited bool, models string, expiresAt time.Time) (*entity.APIKey, error) {
	key := "sk-" + generateKey(32)

	apiKey := &entity.APIKey{
		Key:       key,
		Name:      name,
		UserID:    userID,
		Quota:     quota,
		Unlimited: unlimited,
		Models:    models,
		ExpiresAt: expiresAt,
		Status:    entity.APIKeyActive,
	}

	if err := s.apiKeyClient.Create(ctx, apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

// Validate 验证API Key（用于/v1/chat/completions等API调用）
func (s *apiKeyService) Validate(ctx context.Context, key string) (*entity.APIKey, error) {
	apiKey, err := s.apiKeyClient.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, errors.New("invalid api key")
	}
	if !apiKey.IsAvailable() {
		return nil, errors.New("api key is disabled or expired")
	}
	return apiKey, nil
}

// GetByUserID 获取用户的API Key列表
func (s *apiKeyService) GetByUserID(ctx context.Context, userID int64) ([]*entity.APIKey, error) {
	return s.apiKeyClient.GetByUserID(ctx, userID)
}

// Delete 删除API Key
func (s *apiKeyService) Delete(ctx context.Context, id int64, userID int64) error {
	apiKey, err := s.apiKeyClient.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if apiKey == nil || apiKey.UserID != userID {
		return errors.New("not authorized")
	}
	return s.apiKeyClient.Delete(ctx, id)
}

// CheckQuota 检查额度（用于计费预扣）
func (s *apiKeyService) CheckQuota(ctx context.Context, id int64, amount float64) (bool, error) {
	apiKey, err := s.apiKeyClient.GetByID(ctx, id)
	if err != nil {
		return false, err
	}
	return apiKey.HasQuota(amount), nil
}

// UseQuota 使用额度（计费PreConsume）
func (s *apiKeyService) UseQuota(ctx context.Context, id int64, amount float64) error {
	return s.apiKeyClient.UseQuota(ctx, id, amount)
}

// RefundQuota 退还额度（计费Refund）
func (s *apiKeyService) RefundQuota(ctx context.Context, id int64, amount float64) error {
	return s.apiKeyClient.RefundQuota(ctx, id, amount)
}

func generateKey(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
