package infra

import (
	"github.com/google/wire"
	userclient "github.com/lionellc/fusion-gate/internal/domain/user/client"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"github.com/lionellc/fusion-gate/internal/infra/redis"
	"github.com/lionellc/fusion-gate/internal/repo"
)

var ProviderSet = wire.NewSet(
	db.NewDB,
	redis.NewRedisClient,

	// User module
	repo.NewUserRepo,
	wire.Bind(new(userclient.UserClient), new(*repo.UserRepo)),
)
