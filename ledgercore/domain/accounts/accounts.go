package accounts

import (
	"context"
	"ledgercore/service/database"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Accounts interface {
	CreateAccount(ctx context.Context, accountCreateRequest AccountCreateRequest) (*AccountDetails, error)
	SetAccountCreditLimit(ctx context.Context, id uuid.UUID, currency string, creditLimit int64) error
	SuspendAccount(ctx context.Context, id uuid.UUID, reason string) error
	ResumeAccount(ctx context.Context, id uuid.UUID) error
	GetAccount(ctx context.Context, id uuid.UUID) (*AccountDetails, error)
}

type accounts struct {
	log               *zap.Logger
	accountRepository database.AccountRepository
}

func (a *accounts) CreateAccount(ctx context.Context, accountCreateRequest AccountCreateRequest) (*AccountDetails, error) {
	if accountCreateRequest.Currencies == nil {
		return nil, ErrNoCurrencySupplied
	}
	currencies := make([]database.CurrencyCreateRequest, 0)
	for _, c := range accountCreateRequest.Currencies {
		currencies = append(currencies, database.CurrencyCreateRequest{
			Currency:    c.Currency,
			CreditLimit: c.CreditLimit,
		})
	}
	accountId, err := a.accountRepository.CreateAccount(ctx, database.AccountCreateRequest{
		SubsystemId: accountCreateRequest.SubsystemId,
		AllowAsync:  accountCreateRequest.AllowAsync,
		Currencies:  currencies,
	})
	// TODO map db error to domain error
	return &AccountDetails{
		Account{
			Id:          *accountId,
			AllowAsync:  accountCreateRequest.AllowAsync,
			IsSuspended: false,
		},
		[]AccountCurrency{},
	}, err
}
func (a *accounts) SetAccountCreditLimit(ctx context.Context, id uuid.UUID, currency string, creditLimit int64) error {
	err := a.accountRepository.SetAccountCreditLimit(ctx, id, currency, creditLimit)
	// TODO map db error to domain error
	return err
}
func (a *accounts) SuspendAccount(ctx context.Context, id uuid.UUID, reason string) error {
	err := a.accountRepository.SuspendAccount(ctx, id, reason)
	// TODO map db error to domain error
	return err
}
func (a *accounts) ResumeAccount(ctx context.Context, id uuid.UUID) error {
	err := a.accountRepository.ResumeAccount(ctx, id)
	// TODO map db error to domain error
	return err
}
func (a *accounts) GetAccount(ctx context.Context, id uuid.UUID) (*AccountDetails, error) {
	dbAccount, err := a.accountRepository.GetAccount(ctx, id)
	// TODO map db error to domain error
	if err != nil {
		return nil, err
	}
	return &AccountDetails{
		Account{
			Id:          dbAccount.Id,
			SubsystemId: dbAccount.SubsystemId,
			AllowAsync:  dbAccount.AllowAsync,
			IsSuspended: dbAccount.IsSuspended,
		},
		[]AccountCurrency{},
	}, nil
}

type AccountsParameters struct {
	fx.In
	Log               *zap.Logger
	AccountRepository database.AccountRepository
}

func NewAccounts(p AccountsParameters) Accounts {
	accounts := accounts{p.Log.Named("accounts"), p.AccountRepository}
	return &accounts
}
