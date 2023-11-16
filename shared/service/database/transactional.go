package database

import (
	"context"
	"errors"
	"fmt"
	"shared/logger"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const rollbackTimeoutSeconds = time.Second * 5

func rollbackTransaction(tx pgx.Tx, log *zap.Logger, mdc *logger.RequestMdc, operationError error) {
	if operationError == nil || tx.Conn().IsClosed() {
		logger.LogDebug(log, mdc, "nothing to rollback")
		return
	}
	timedContext, cancel := context.WithTimeout(context.Background(), rollbackTimeoutSeconds)
	defer cancel()
	if txErr := tx.Rollback(timedContext); txErr != nil {
		if !errors.Is(txErr, pgx.ErrTxClosed) {
			logger.LogWarn(log, mdc, "could not roll back transaction", txErr)
		}
	} else {
		logger.LogInfo(log, mdc, fmt.Sprintf("rolled back transaction, reason: %v", operationError))
	}
}

func ExecuteTransactionally(pool *pgxpool.Pool, ctx context.Context, log *zap.Logger, mdc *logger.RequestMdc, operation func(tx pgx.Tx) (any, error)) (any, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	var operationError error
	defer rollbackTransaction(tx, log, mdc, operationError)
	result, operationError := operation(tx)
	if operationError != nil {
		return nil, operationError
	}
	operationError = tx.Commit(ctx)
	if operationError != nil {
		return nil, operationError
	}
	return result, nil
}
