package repo

import (
	"context"

	"github.com/lionellc/fusion-gate/internal/domain/user/entity"
)

type UserRepo interface {
	Create(ctx context.Context, user *entity.User) error
	GetById(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error

	// balance operate
	GetBalance(ctx context.Context, id int64) (float64, error)
	AddBalance(ctx context.Context, id int64, amount float64) error
	SubBalance(ctx context.Context, id int64, amount float64) error
}
