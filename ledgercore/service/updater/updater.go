package updater

import (
	"ledgercore/config"
	"ledgercore/service/updater/inmemory"
	"ledgercore/service/updater/temporal"
	"shared/service/database"
	"shared/service/infrastructure/datadog"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Updater interface {
	UpdateBalance(operation database.OperationStatusTransition)
	Stop()
}

type UpdaterParameters struct {
	fx.In
	Log                       *zap.Logger
	Configuration             config.Configuration
	Lifecycle                 fx.Lifecycle
	AccountCurrencyRepository database.AccountCurrencyRepository
	Datadog                   datadog.Datadog
}

func NewUpdater(p UpdaterParameters) (Updater, error) {
	if p.Configuration.InMemory != nil {
		return inmemory.CreateInMemoryBalanceUpdater(p.Log,
			p.AccountCurrencyRepository,
			p.Configuration, p.Datadog,
			p.Configuration.InMemory.AccountCurrencies.Strategy)
	} else {
		return temporal.CreateTemporalBalanceUpdater(p.Log, p.Configuration)
	}
}
