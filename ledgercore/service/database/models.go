package database

import (
	"time"

	"github.com/google/uuid"
)

type AccountCreateRequest struct {
	SubsystemId uuid.UUID
	AllowAsync  bool
	Currencies  []CurrencyCreateRequest
}
type CurrencyCreateRequest struct {
	Currency    string
	CreditLimit *int64
}

type TransactionCreateRequest struct {
	Id              uuid.UUID
	SubsystemId     uuid.UUID
	Description     string
	Amount          int64
	Currency        string
	SourceOperation TransferAccountCurrency
	TargetOperation TransferAccountCurrency
}

type TransferAccountCurrency struct {
	AccountCurrencyId uuid.UUID
	Amount            int64
	OperationType     string
	OperationStatus   string
	IsAsync           bool
}

type Transaction struct {
	Id              uuid.UUID
	Time            time.Time
	SourceOperation Operation
	TargetOperation Operation
	SubsystemId     uuid.UUID
	Description     string
	Amount          int64
	Currency        string
}

type Operation struct {
	Id                int64
	AccountCurrencyId uuid.UUID
	Type              string
	Status            string
	Amount            int64
}

type Account struct {
	Id          uuid.UUID
	SubsystemId uuid.UUID
	AllowAsync  bool
	IsSuspended bool
}

type AccountCurrency struct {
	Id          uuid.UUID
	Currency    string
	Amount      int64
	CreditLimit *int64
}
