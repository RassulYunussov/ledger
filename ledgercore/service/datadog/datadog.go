package datadog

import (
	"context"
	"fmt"
	"ledgercore/config"
	"shared/service/infrastructure/datadog"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewDatadog(log *zap.Logger, lc fx.Lifecycle, cfg config.Configuration) (datadog.Datadog, error) {
	if cfg.Datadog.Enabled {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				datadog.DatadogTracerStart(cfg.Application.Environment, cfg.Application.Name, config.APPLICATION_VERSION)
				return nil
			},
		})
		return datadog.CreateDatadogClient(log.Named("datadog-client"),
			cfg.Datadog.Rate,
			cfg.Application.Name,
			cfg.Application.Environment,
			config.APPLICATION_VERSION,
			fmt.Sprintf("%s:%d", cfg.Datadog.Host, cfg.Datadog.Port),
		)
	}
	return datadog.CreateNoopDatadogClient(log.Named("datadog-client"),
		cfg.Datadog.Rate,
		cfg.Application.Name,
		cfg.Application.Environment,
		config.APPLICATION_VERSION,
		fmt.Sprintf("%s:%d", cfg.Datadog.Host, cfg.Datadog.Port),
	)
}
