package main

import (
	"balanceupdater/config"
	"balanceupdater/domain"
	"shared/logger"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(logger.NewLogger),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		config.ConfigurationModule,
		domain.DomainModule,
	).Run()
}
