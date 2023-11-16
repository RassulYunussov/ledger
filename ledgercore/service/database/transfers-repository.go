package database

import (
	"context"
	"errors"
	"fmt"
	"shared/logger"
	"shared/service/database"
	"shared/service/infrastructure/datadog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type TransfersRepository interface {
	CreateTransaction(ctx context.Context, requestMdc *logger.RequestMdc, transactionCreateRequest *TransactionCreateRequest) (*Transaction, error)
	GetTransactionDetails(ctx context.Context, requestMdc *logger.RequestMdc, transactionId uuid.UUID) (*Transaction, error)
}

type transfersRepository struct {
	log       *zap.Logger
	pgContext database.PgContext
	dd        datadog.Datadog
}

func (r *transfersRepository) createTransaction(ctx context.Context, tx pgx.Tx, requestMdc *logger.RequestMdc, transactionCreateRequest *TransactionCreateRequest) (*Transaction, error) {

	if !transactionCreateRequest.SourceOperation.IsAsync {
		var amount int64
		var creditLimit *int64
		err := tx.QueryRow(ctx, "update account_currency set amount = amount+$1 where id = $2 returning amount, credit_limit",
			transactionCreateRequest.SourceOperation.Amount,
			transactionCreateRequest.SourceOperation.AccountCurrencyId).Scan(&amount, &creditLimit)
		if err != nil {
			logger.LogWarn(r.log, requestMdc, "error updating account currency", err)
			increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
			return nil, err
		}
		if creditLimit != nil && amount < *creditLimit {
			logger.LogInfo(r.log, requestMdc, fmt.Sprintf("source account currency %v - insufficient amount money", transactionCreateRequest.SourceOperation.AccountCurrencyId))
			increment(r.dd, "transaction.fail", requestMdc.Subsystem, "reason:source_insufficient_amount")
			return nil, ErrInsufficientAmount
		}
	}

	var transactionDate time.Time
	err := tx.QueryRow(ctx, `insert into transaction (id, 
										subsystem_id, 
										source_account_currency_id, 
										target_account_currency_id,
										amount,
										currency_id,
										transaction_date,
										description)
								values ($1, $2, $3, $4, $5, $6, now(), $7) returning transaction_date`,
		transactionCreateRequest.Id,
		transactionCreateRequest.SubsystemId,
		transactionCreateRequest.SourceOperation.AccountCurrencyId,
		transactionCreateRequest.TargetOperation.AccountCurrencyId,
		transactionCreateRequest.Amount,
		transactionCreateRequest.Currency,
		transactionCreateRequest.Description).Scan(&transactionDate)

	if err != nil {
		logger.LogWarn(r.log, requestMdc, "error inserting into transcation", err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:pgerr_%s", pgErr.Code))
			if pgErr.Code == "23505" {
				return nil, ErrDuplicateTransaction
			}
		}
		increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, err
	}

	logger.LogDebug(r.log, requestMdc, fmt.Sprintf("stored transaction with date %v", transactionDate))
	var sourceOperationId int64
	err = tx.QueryRow(ctx, `insert into account_operation (account_currency_id, 
														operation_type_id,
														operation_status_id, 
														amount, 
														transaction_id)
										values($1, $2, $3, $4, $5) returning id`,
		transactionCreateRequest.SourceOperation.AccountCurrencyId,
		transactionCreateRequest.SourceOperation.OperationType,
		transactionCreateRequest.SourceOperation.OperationStatus,
		transactionCreateRequest.SourceOperation.Amount,
		transactionCreateRequest.Id,
	).Scan(&sourceOperationId)

	if err != nil {
		logger.LogWarn(r.log, requestMdc, "error storing source operation", err)
		increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, err
	}

	var targetOperationId int64
	err = tx.QueryRow(ctx, `insert into account_operation (account_currency_id, 
														operation_type_id,
														operation_status_id, 
														amount, 
														transaction_id)
										values($1, $2, $3, $4, $5) returning id`,
		transactionCreateRequest.TargetOperation.AccountCurrencyId,
		transactionCreateRequest.TargetOperation.OperationType,
		transactionCreateRequest.TargetOperation.OperationStatus,
		transactionCreateRequest.TargetOperation.Amount,
		transactionCreateRequest.Id,
	).Scan(&targetOperationId)

	if err != nil {
		logger.LogWarn(r.log, requestMdc, "error storing target operation", err)
		increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, err
	}

	// store into outpost
	if err := r.storeOutpost(ctx, tx, requestMdc, &(transactionCreateRequest.SourceOperation), sourceOperationId); err != nil {
		return nil, err
	}
	if err := r.storeOutpost(ctx, tx, requestMdc, &(transactionCreateRequest.TargetOperation), targetOperationId); err != nil {
		return nil, err
	}

	return &Transaction{transactionCreateRequest.Id,
		transactionDate,
		Operation{sourceOperationId,
			transactionCreateRequest.SourceOperation.AccountCurrencyId,
			transactionCreateRequest.SourceOperation.OperationType,
			transactionCreateRequest.SourceOperation.OperationStatus,
			transactionCreateRequest.SourceOperation.Amount},
		Operation{targetOperationId,
			transactionCreateRequest.TargetOperation.AccountCurrencyId,
			transactionCreateRequest.TargetOperation.OperationType,
			transactionCreateRequest.TargetOperation.OperationStatus,
			transactionCreateRequest.TargetOperation.Amount},
		transactionCreateRequest.SubsystemId,
		transactionCreateRequest.Description,
		transactionCreateRequest.Amount,
		transactionCreateRequest.Currency}, nil

}

func (r *transfersRepository) storeOutpost(ctx context.Context, tx pgx.Tx, requestMdc *logger.RequestMdc, tac *TransferAccountCurrency, operationId int64) error {
	if tac.IsAsync {
		cmdTag, err := tx.Exec(ctx, "insert into account_operation_outpost (id, created_date) values($1, now())", operationId)
		if err != nil {
			logger.LogWarn(r.log, requestMdc, "error storing operation to outpost", err)
			increment(r.dd, "transaction.error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
			return err
		}
		r.log.Debug(fmt.Sprintf("operation outpost store result %v", cmdTag))
	}
	return nil
}

func (r *transfersRepository) CreateTransaction(ctx context.Context, requestMdc *logger.RequestMdc, transactionCreateRequest *TransactionCreateRequest) (*Transaction, error) {
	transaction, err := database.ExecuteTransactionally(r.pgContext.GetPool(), ctx, r.log, requestMdc,
		func(tx pgx.Tx) (any, error) {
			return r.createTransaction(ctx, tx, requestMdc, transactionCreateRequest)
		},
	)
	if err != nil {
		return nil, err
	}
	return transaction.(*Transaction), err
}

func (r *transfersRepository) getOperation(ctx context.Context, requestMdc *logger.RequestMdc, transactionId uuid.UUID, accountCurrencyId uuid.UUID) (*Operation, error) {
	var operation Operation

	err := r.pgContext.GetPool().QueryRow(ctx, `select 
									id,
									account_currency_id,
									operation_type_id,
									operation_status_id,
									amount
								from account_operation
								where transaction_id = $1 and account_currency_id = $2`, transactionId, accountCurrencyId).
		Scan(&operation.Id,
			&operation.AccountCurrencyId,
			&operation.Type,
			&operation.Status,
			&operation.Amount)

	if err != nil {
		return nil, err
	}
	return &operation, nil
}

func (r *transfersRepository) GetTransactionDetails(ctx context.Context, requestMdc *logger.RequestMdc, transactionId uuid.UUID) (*Transaction, error) {

	var subsystemId uuid.UUID
	var sourceAccountCurrencyId uuid.UUID
	var targetAccountCurrencyId uuid.UUID
	var transactionDate time.Time
	var description string
	var amount int64
	var currency string

	err := r.pgContext.GetPool().QueryRow(ctx, `select 
									subsystem_id,
									source_account_currency_id,
									target_account_currency_id,
									transaction_date,
									description,
									amount,
									currency_id
								from transaction
								where id = $1`, transactionId).
		Scan(&subsystemId,
			&sourceAccountCurrencyId,
			&targetAccountCurrencyId,
			&transactionDate,
			&description,
			&amount,
			&currency)
	if err != nil {
		logger.LogError(r.log, requestMdc, "error selecting transaction", err)
		return nil, err
	}

	sourceOperation, err := r.getOperation(ctx, requestMdc, transactionId, sourceAccountCurrencyId)

	if err != nil {
		logger.LogError(r.log, requestMdc, fmt.Sprintf("error selecting source operation by account currency %v", sourceAccountCurrencyId), err)
		increment(r.dd, "transaction.err", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, err
	}

	targetOperation, err := r.getOperation(ctx, requestMdc, transactionId, targetAccountCurrencyId)

	if err != nil {
		logger.LogError(r.log, requestMdc, fmt.Sprintf("error selecting source operation by account currency %v", targetAccountCurrencyId), err)
		increment(r.dd, "transaction.err", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, err
	}

	return &Transaction{transactionId,
		transactionDate,
		*sourceOperation,
		*targetOperation,
		subsystemId,
		description,
		amount,
		currency}, nil
}

func NewTransfersRepository(log *zap.Logger, pgContext database.PgContext, dd datadog.Datadog) TransfersRepository {
	return &transfersRepository{log.Named("transfers-repository"), pgContext, dd}
}
