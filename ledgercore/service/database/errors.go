package database

import "errors"

var ErrNotFound = errors.New("Record not found")
var ErrInsufficientAmount = errors.New("Insufficent amount")
var ErrDuplicateTransaction = errors.New("Duplicate transaction")
