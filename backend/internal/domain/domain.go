package domain

import (
	"github.com/google/wire"
	"github.com/lionellc/fusion-gate/internal/domain/user/service"
)

var ProviderSet = wire.NewSet(
	service.NewAuthService,
)
