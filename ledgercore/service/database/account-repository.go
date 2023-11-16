package database

import (
	"context"
	"fmt"
	"shared/logger"
	"shared/service/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, accountCreateRequest AccountCreateRequest) (*uuid.UUID, error)
	GetAccount(ctx context.Context, id uuid.UUID) (*Account, error)
	GetAccountCurrency(ctx context.Context, id uuid.UUID, currency string) (*AccountCurrency, error)
	SetAccountCreditLimit(ctx context.Context, accountId uuid.UUID, currency string, creditLimit int64) error
	SuspendAccount(ctx context.Context, accountId uuid.UUID, reason string) error
	ResumeAccount(ctx context.Context, accountId uuid.UUID) error
}

type accountRepository struct {
	log       *zap.Logger
	pgContext database.PgContext
}

func (r *accountRepository) createAccountTransactionally(ctx context.Context, tx pgx.Tx, accountCreateRequest *AccountCreateRequest) (*uuid.UUID, error) {
	accountId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	cmdTag, operationError := tx.Exec(ctx, `insert into account(id, subsystem_id, allow_async, is_suspended, created_date)
						values($1,$2,$3,false,now())`, accountId, accountCreateRequest.SubsystemId, accountCreateRequest.AllowAsync)
	if operationError != nil {
		return nil, operationError
	}
	r.log.Debug(fmt.Sprintf("created account record %v", cmdTag))

	for _, currency := range accountCreateRequest.Currencies {
		currencyId, err := uuid.NewRandom()
		if err != nil {
			return nil, err
		}
		cmdTag, operationError = tx.Exec(ctx, `insert into account_currency(id, account_id, currency_id, amount, credit_limit)
						values($1,$2,$3,0,$4)`, currencyId, accountId, currency.Currency, currency.CreditLimit)
		if operationError != nil {
			return nil, operationError
		}
		r.log.Debug(fmt.Sprintf("created account currency record %v", cmdTag))
	}
	return &accountId, nil
}

func (r *accountRepository) CreateAccount(ctx context.Context, accountCreateRequest AccountCreateRequest) (*uuid.UUID, error) {
	accountId, err := database.ExecuteTransactionally(r.pgContext.GetPool(), ctx, r.log, &logger.RequestMdc{Subsystem: accountCreateRequest.SubsystemId},
		func(tx pgx.Tx) (any, error) {
			return r.createAccountTransactionally(ctx, tx, &accountCreateRequest)
		},
	)
	if err != nil {
		return nil, err
	}
	return accountId.(*uuid.UUID), nil
}
func (r *accountRepository) GetAccount(ctx context.Context, id uuid.UUID) (*Account, error) {
	var account Account
	err := r.pgContext.GetPool().QueryRow(ctx, "select subsystem_id, allow_async, is_suspended from account where id = $1", id).
		Scan(&account.SubsystemId, &account.AllowAsync, &account.IsSuspended)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	account.Id = id
	return &account, nil
}

func (r *accountRepository) GetAccountCurrency(ctx context.Context, account_id uuid.UUID, currency string) (*AccountCurrency, error) {
	var accountCurrency AccountCurrency
	err := r.pgContext.GetPool().QueryRow(ctx, "select id, currency_id, amount, credit_limit from account_currency where account_id = $1 and currency_id=$2",
		account_id, currency).
		Scan(&accountCurrency.Id, &accountCurrency.Currency, &accountCurrency.Amount, &accountCurrency.CreditLimit)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &accountCurrency, nil
}

func (r *accountRepository) SetAccountCreditLimit(ctx context.Context, accountId uuid.UUID, currency string, creditLimit int64) error {
	cmdTag, err := r.pgContext.GetPool().Exec(ctx, "update account_currency set credit_limit = $1 where account_id =$2 and currency_id =$3", creditLimit, accountId, currency)
	if err != nil {
		return err
	}
	r.log.Info(fmt.Sprintf("updated account credit limit %v for account %v and currency %s, %v", creditLimit, accountId, currency, cmdTag))
	return nil
}

func (r *accountRepository) SuspendAccount(ctx context.Context, accountId uuid.UUID, reason string) error {
	cmdTag, err := r.pgContext.GetPool().Exec(ctx, "update account set is_suspended = true, suspended_date=now(), suspended_reason=$2 where account_id =$1", accountId, reason)
	if err != nil {
		return err
	}
	r.log.Info(fmt.Sprintf("suspended account %v, %v", accountId, cmdTag))
	return nil
}

func (r *accountRepository) ResumeAccount(ctx context.Context, accountId uuid.UUID) error {
	cmdTag, err := r.pgContext.GetPool().Exec(ctx, "update account set is_suspended = false, suspended_date=null, suspended_reason=null where account_id =$1", accountId)
	if err != nil {
		return err
	}
	r.log.Info(fmt.Sprintf("resumed account %v, %v", accountId, cmdTag))
	return nil
}

func NewAccountRepository(log *zap.Logger, pgContext database.PgContext) AccountRepository {
	return &accountRepository{log.Named("account-repository"), pgContext}
}
