package domain

import (
	"ledgercore/config"
	"ledgercore/domain/accounts"
	"ledgercore/domain/transfers"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var DomainModule = fx.Module("application-domain",
	fx.Provide(transfers.NewTransfers, accounts.NewAccounts),
	fx.Invoke(func(log *zap.Logger, cfg config.Configuration) {
		log.Info("Starting Application",
			zap.String("name", cfg.Application.Name),
			zap.String("environment", cfg.Application.Environment),
			zap.String("version", config.APPLICATION_VERSION),
		)
	}),
)
