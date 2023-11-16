package service

import (
	"ledgercore/service/database"
	"ledgercore/service/datadog"
	"ledgercore/service/grpcserver"
	"ledgercore/service/grpcserver/account"
	"ledgercore/service/grpcserver/transfer"
	"ledgercore/service/updater"
	shareddb "shared/service/database"

	"go.uber.org/fx"
)

var ServiceModule = fx.Module("application-services",
	fx.Provide(shareddb.NewPgContext,
		shareddb.NewAccountCurrencyRepository,
		database.NewAccountRepository,
		database.NewTransfersRepository,
		updater.NewUpdater,
		grpcserver.NewGrpcServer,
		transfer.NewTransferServer,
		account.NewAccountsServer,
		datadog.NewDatadog,
		NewShutdownOrchestrator,
	),
	fx.Invoke(func(shutdownOrchestrator *shutdownOrchestrator) {}))
