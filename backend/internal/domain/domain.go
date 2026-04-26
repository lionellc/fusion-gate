package domain

import (
	"github.com/google/wire"
	apikeyService "github.com/lionellc/fusion-gate/internal/domain/apikey/service"
	"github.com/lionellc/fusion-gate/internal/domain/user/service"
)

var ProviderSet = wire.NewSet(
	service.NewAuthService,
	apikeyService.NewAPIKeyService,
)
