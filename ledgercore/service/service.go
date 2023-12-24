package service

import (
	"ledgercore/service/database"
	"ledgercore/service/datadog"
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
		datadog.NewDatadog))
