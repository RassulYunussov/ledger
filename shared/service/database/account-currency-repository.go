package database

import (
	"context"
	"fmt"
	"shared/logger"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AccountCurrencyRepository interface {
	UpdateBalance(ctx context.Context, operation OperationStatusTransition) error
	GetOperationsFromOutpost(ctx context.Context, upToCurrentTimeMinusSeconds int) ([]int64, error)
}

type accountCurrencyRepository struct {
	log       *zap.Logger
	pgContext PgContext
}

func (repo *accountCurrencyRepository) processTransactionally(ctx context.Context, tx pgx.Tx, operation OperationStatusTransition) error {

	var accountCurrencyId uuid.UUID
	var operationTypeId string
	var operationStatusId string
	var amount int64

	err := tx.QueryRow(ctx, `select 
									account_currency_id, 
									operation_type_id, 
									operation_status_id, 
									amount 
								from account_operation 
								where id = $1 for update`, operation.Id).
		Scan(&accountCurrencyId,
			&operationTypeId,
			&operationStatusId,
			&amount)

	if err != nil {
		repo.log.Error("error fetching operation data", zap.Int64("operation", operation.Id), zap.Error(err))
		return err
	}

	repo.log.Debug(fmt.Sprintf("selected account operation %d", operation.Id))

	var operationOutpostDate time.Time
	err = tx.QueryRow(ctx, `select created_date from account_operation_outpost where id = $1 for update`, operation.Id).
		Scan(&operationOutpostDate)

	if err != nil {
		repo.log.Error("error fetching operation outpost date", zap.Int64("operation", operation.Id), zap.Error(err))
		return err
	}

	repo.log.Debug(fmt.Sprintf("selected date from operation outpost %v for operation %d", operationOutpostDate, operation.Id))

	if operationStatusId != operation.From {
		repo.log.Warn("operation status conflict", zap.Int64("operation", operation.Id), zap.String("status", operationStatusId))
		return ErrOperationStatusConflict
	}

	commandTag, err := tx.Exec(ctx, "update account_currency set amount = amount+$1 where id = $2", amount, accountCurrencyId)

	if err != nil {
		repo.log.Error("error updating account currency amount", zap.Int64("operation", operation.Id))
		return err
	}

	repo.log.Debug(fmt.Sprintf("account currency affected rows %d for operation %d", commandTag.RowsAffected(), operation.Id))

	commandTag, err = tx.Exec(ctx, "update account_operation set operation_status_id = $1 where id = $2", operation.To, operation.Id)

	if err != nil {
		repo.log.Error("error updating account operation status", zap.Int64("operation", operation.Id), zap.String("desired status", operation.To))
		return err
	}

	repo.log.Debug(fmt.Sprintf("account operation status affected rows %d for operation %d", commandTag.RowsAffected(), operation.Id))

	commandTag, err = tx.Exec(ctx, "delete from account_operation_outpost where id = $1", operation.Id)

	if err != nil {
		repo.log.Error("error deleteing data from account operation outpost", zap.Int64("operation", operation.Id), zap.Error(err))
		return err
	}

	repo.log.Debug(fmt.Sprintf("deleted %d rows from account operation outpost for peration %d", commandTag.RowsAffected(), operation.Id))

	return nil

}

func (repo *accountCurrencyRepository) GetOperationsFromOutpost(ctx context.Context, upToCurrentTimeMinusSeconds int) ([]int64, error) {
	ids := make([]int64, 0)
	rows, err := repo.pgContext.GetPool().Query(ctx, fmt.Sprintf("select id from account_operation_outpost where created_date < now() - interval '%d seconds'", upToCurrentTimeMinusSeconds))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (repo *accountCurrencyRepository) UpdateBalance(ctx context.Context, operation OperationStatusTransition) error {
	_, err := ExecuteTransactionally(repo.pgContext.GetPool(), ctx, repo.log, &logger.RequestMdc{Subsystem: operation.AccountSubsystemId},
		func(tx pgx.Tx) (any, error) {
			return nil, repo.processTransactionally(ctx, tx, operation)
		},
	)
	return err
}

func NewAccountCurrencyRepository(log *zap.Logger, pgContext PgContext) AccountCurrencyRepository {
	return &accountCurrencyRepository{log.Named("account-currency-repository"), pgContext}
}
