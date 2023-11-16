package domain

import (
	"balanceupdater/config"
	"balanceupdater/domain/processor"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var DomainModule = fx.Module("application-domain",
	fx.Provide(processor.NewProcessor),
	fx.Invoke(func(log *zap.Logger, cfg config.Configuration, p *processor.Processor) {
		log.Info("Starting Application",
			zap.String("name", cfg.Application.Name),
			zap.String("environment", cfg.Application.Environment),
			zap.String("version", config.APPLICATION_VERSION),
		)
	}),
)
