//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/lionellc/fusion-gate/internal/config"
	"github.com/lionellc/fusion-gate/internal/handler"
	"github.com/lionellc/fusion-gate/internal/infra"
)

func wireApp(cfg *config.Config) (*gin.Engine, func(), error) {
	panic(wire.Build(
		infra.ProviderSet,
		handler.NewRouter,
	))
}
