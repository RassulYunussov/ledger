package transfers

import (
	"time"

	"github.com/google/uuid"
)

type TransactionRequest struct {
	Id                uuid.UUID
	SubsystemId       uuid.UUID
	Amount            int64
	Fractional        int32
	Currency          string
	SourceAccountInfo TransferAccountInfoRequest
	TargetAccountInfo TransferAccountInfoRequest
	Description       string
}

type TransferAccountInfoRequest struct {
	Id       uuid.UUID
	Currency string
}

type TransactionResponse struct {
	Id              uuid.UUID
	Time            time.Time
	SourceOperation OperationResponse
	TargetOperation OperationResponse
	SubsystemId     uuid.UUID
	Description     string
	Amount          int64
	Fractional      int32
	Currency        string
}

type OperationResponse struct {
	Id         int64
	Type       string
	Status     string
	Amount     int64
	Fractional int32
}

type accountCurrencyCacheKey struct {
	accountId uuid.UUID
	currency  string
}
