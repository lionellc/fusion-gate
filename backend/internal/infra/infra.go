package infra

import (
	"github.com/google/wire"
	"github.com/lionellc/fusion-gate/internal/infra/db"
	"github.com/lionellc/fusion-gate/internal/infra/persistence"
	"github.com/lionellc/fusion-gate/internal/infra/redis"
)

var ProviderSet = wire.NewSet(
	db.NewDB,
	redis.NewRedisClient,

	persistence.NewUserRepo,
	persistence.NewAPIKeyRepo,
)
