package persistence

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/user/entity"
	"github.com/lionellc/fusion-gate/internal/domain/user/repo"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"gorm.io/gorm"
)

type userRepo struct {
	db *db.DB
}

func NewUserRepo(db *db.DB) repo.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *userRepo) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepo) GetBalance(ctx context.Context, id int64) (float64, error) {
	user, err := r.GetById(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (r *userRepo) AddBalance(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *userRepo) SubBalance(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Where("balance >= ?", amount).
		Update("balance", gorm.Expr("balance - ?", amount)).Error
}
