package updater

import (
	"ledgercore/config"
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
		return createInMemoryBalanceUpdater(p.Log,
			p.AccountCurrencyRepository,
			p.Configuration, p.Datadog,
			p.Configuration.InMemory.AccountCurrencies.Strategy)
	} else {
		return createTemporalBalanceUpdater(p.Log, p.Configuration)
	}
}
