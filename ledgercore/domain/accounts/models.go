package accounts

import (
	"github.com/google/uuid"
)

type AccountCurrency struct {
	Currency    string
	Amount      int64
	CreditLimit *int64
}

type Account struct {
	Id               uuid.UUID
	SubsystemId      uuid.UUID
	AllowAsync       bool
	IsSuspended      bool
	SuspendedDetails *string
}

type AccountCreateRequest struct {
	SubsystemId uuid.UUID
	AllowAsync  bool
	Currencies  []CurrencyCreateRequest
}

type CurrencyCreateRequest struct {
	Currency    string
	CreditLimit *int64
}

type AccountDetails struct {
	Account
	Currencies []AccountCurrency
}
