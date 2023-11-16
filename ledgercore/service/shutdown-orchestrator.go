package service

import (
	"context"
	"ledgercore/service/grpcserver"
	"ledgercore/service/updater"
	"shared/service/database"
	"shared/service/infrastructure/datadog"

	"go.uber.org/fx"
)

type shutdownOrchestrator struct{}

func NewShutdownOrchestrator(
	lc fx.Lifecycle,
	updater updater.Updater,
	pgContext database.PgContext,
	grpcserver grpcserver.GrpcServer,
	dd datadog.Datadog) *shutdownOrchestrator {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			grpcserver.Stop()
			updater.Stop()
			pgContext.Stop()
			datadog.ShutdownDatadogClient(dd)
			datadog.DatadogTracerStop()
			return nil
		},
	})
	return &shutdownOrchestrator{}
}
