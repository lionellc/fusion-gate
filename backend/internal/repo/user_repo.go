package repo

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/user/entity"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *db.DB
}

func NewUserRepo(db *db.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) GetById(ctx context.Context, id int64) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepo) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *UserRepo) GetBalance(ctx context.Context, id int64) (float64, error) {
	user, err := r.GetById(ctx, id)
	if err != nil {
		return 0, err
	}
	return user.Balance, nil
}

func (r *UserRepo) AddBalance(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *UserRepo) SubBalance(ctx context.Context, id int64, amount float64) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).
		Where("id = ?", id).
		Where("balance >= ?", amount).
		Update("balance", gorm.Expr("balance - ?", amount)).Error
}
