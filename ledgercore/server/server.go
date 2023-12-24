package server

import (
	"context"
	"ledgercore/domain"
	"ledgercore/server/grpcserver"
	"ledgercore/service"
	"ledgercore/service/updater"

	shareddb "shared/service/database"
	shdatadog "shared/service/infrastructure/datadog"

	"go.uber.org/fx"
)

// ServerModule is an aggregation of servers provided by application
// It is responsible for orchestrating startup/shutdown sequence of services
var ServerModule = fx.Module("application-servers",
	domain.DomainModule,
	service.ServiceModule,
	grpcserver.GrpcServerModule,
	fx.Invoke(func(lc fx.Lifecycle,
		updater updater.Updater,
		pgContext shareddb.PgContext,
		grpcserver grpcserver.GrpcServer,
		dd shdatadog.Datadog) {
		lc.Append(fx.Hook{
			OnStart: func(ctx context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				grpcserver.Stop()
				updater.Stop()
				pgContext.Stop()
				shdatadog.ShutdownDatadogClient(dd)
				shdatadog.DatadogTracerStop()
				return nil
			},
		})
	}))
