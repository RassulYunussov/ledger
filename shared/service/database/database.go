package database

import (
	"context"
	"fmt"
	"ledgercore/config"
	"shared/service/infrastructure/datadog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const pingTimeout = time.Second * 2
const poolRotationTimeout = time.Second * 2

type PgContext interface {
	GetPool() *pgxpool.Pool
	Stop()
}

type PgContextParameters struct {
	fx.In
	Log           *zap.Logger
	Lifecycle     fx.Lifecycle
	Configuration config.Configuration
	Datadog       datadog.Datadog
}

type pgContext struct {
	pool         *pgxpool.Pool
	fallbackPool *pgxpool.Pool
	isActive     bool
	log          *zap.Logger
	dd           datadog.Datadog
}

func (p *pgContext) GetPool() *pgxpool.Pool {
	return p.pool
}

func (p *pgContext) Stop() {
	p.isActive = false
	p.log.Info("close database pool")
	p.pool.Close()
	p.log.Info("close database fallback pool")
	p.fallbackPool.Close()
}

func (p *pgContext) startPoolRotationRoutine(connString string) {
	for p.isActive {
		timedContext, cancel := context.WithTimeout(context.Background(), pingTimeout)
		if err := p.pool.Ping(timedContext); err != nil {
			p.log.Error("pool is dead, replacing by fallbackPool", zap.Error(err))
			oldPool := p.pool
			p.pool = p.fallbackPool
			poolRotationTimedContext, cancelPoolRotation := context.WithTimeout(context.Background(), poolRotationTimeout)
			newFallbackPool, err := pgxpool.New(poolRotationTimedContext, connString)
			if err != nil {
				if oldPool != p.pool {
					p.log.Info("closing old pool")
					go oldPool.Close()
				} else {
					p.log.Info("not closing old pool, the only pool we have")
				}
				p.log.Error("new fallbackPool couldnt't be created", zap.Error(err))
				p.dd.Increment("database.pool.replacement", "result:fail")
			} else {
				go oldPool.Close()
				p.fallbackPool = newFallbackPool
				p.dd.Increment("database.pool.replacement", "result:success")
			}
			cancelPoolRotation()
		}
		cancel()
		time.Sleep(time.Second)
	}
}

func NewPgContext(p PgContextParameters) PgContext {

	databaseContext := pgContext{log: p.Log.Named("database-context"), dd: p.Datadog}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			databaseConfig := &p.Configuration.Database
			connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?pool_max_conns=%d", databaseConfig.User,
				databaseConfig.Password, databaseConfig.Host, databaseConfig.Port, databaseConfig.Name, databaseConfig.PoolSize)
			pool, err := pgxpool.New(ctx, connString)
			if err != nil {
				return err
			}
			err = pool.Ping(ctx)
			if err != nil {
				return err
			}
			fallbackPool, err := pgxpool.New(ctx, connString)
			if err != nil {
				databaseContext.log.Error("fallbackPool couldn't be created", zap.Error(err))
				return err
			}
			databaseContext.isActive = true
			databaseContext.pool = pool
			databaseContext.fallbackPool = fallbackPool
			// that is for resiliency, noticed the death of pool without recuperating for a long period!
			go databaseContext.startPoolRotationRoutine(connString)
			return nil
		},
	})

	return &databaseContext
}
