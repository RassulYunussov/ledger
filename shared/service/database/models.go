package database

import (
	"time"

	"github.com/google/uuid"
)

const (
	OPERATION_STATUS_CREATED = "CREATED"
	OPERATION_STATUS_POSTED  = "POSTED"

	OPERATION_TYPE_OUT = "OUT"
	OPERATION_TYPE_IN  = "IN"
)

type OperationStatusTransition struct {
	AccountSubsystemId uuid.UUID
	Id                 int64
	AccountCurrencyId  uuid.UUID
	From               string
	To                 string
	OperationType      string
	Timestamp          time.Time // observability purpose
}

type Currency struct {
	IsoName       string
	NumberToBasic int32
}

func GetOperationTypes() []string {
	return []string{OPERATION_TYPE_OUT, OPERATION_TYPE_IN}
}
