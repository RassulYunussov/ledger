package transfers

import "errors"

var ErrSameAccount = errors.New("transfer is not allowed for same account")
var ErrCrossCurrency = errors.New("cross curency transfers not allowed")
var ErrNoCurrency = errors.New("account does not have selected currency")
var ErrAccountNotFound = errors.New("account not found")
var ErrAccountSuspended = errors.New("account suspended")

var ErrTransactionDetailsConflict = errors.New("Transaction details conflict")
var ErrInsufficientAmount = errors.New("Insufficent amount")

var ErrInternal = errors.New("Internal error")
