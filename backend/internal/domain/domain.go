package domain

import (
	"github.com/google/wire"
	apikeyService "github.com/lionellc/fusion-gate/internal/domain/apikey/service"
	channelService "github.com/lionellc/fusion-gate/internal/domain/channel/service"
	authService "github.com/lionellc/fusion-gate/internal/domain/user/service"
)

var ProviderSet = wire.NewSet(
	authService.NewAuthService,
	apikeyService.NewAPIKeyService,
	channelService.NewHealthCheckService,
	channelService.NewChannelService,
)
