package transfers

import (
	"context"
	"errors"
	"fmt"
	"ledgercore/config"
	"ledgercore/domain/transfers/currency"
	lcdatabase "ledgercore/service/database"
	"ledgercore/service/updater"
	"shared/logger"
	"shared/service/database"
	"shared/service/infrastructure/datadog"
	"time"

	"github.com/google/uuid"
	"github.com/jellydator/ttlcache/v3"
	"github.com/sony/gobreaker"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Transfers interface {
	CreateTransaction(ctx context.Context, requestMdc *logger.RequestMdc, transaction *TransactionRequest) (*TransactionResponse, error)
}

type TransfersParameters struct {
	fx.In
	Log                 *zap.Logger
	TransfersRepository lcdatabase.TransfersRepository
	AccountRepository   lcdatabase.AccountRepository
	Updater             updater.Updater
	Datadog             datadog.Datadog
	Configuration       config.Configuration
}

type transfers struct {
	log                        *zap.Logger
	transfersRepository        lcdatabase.TransfersRepository
	accountRepository          lcdatabase.AccountRepository
	updater                    updater.Updater
	dd                         datadog.Datadog
	createTransactionTimeoutMs time.Duration
	cb                         *gobreaker.CircuitBreaker
	accountsCache              *ttlcache.Cache[uuid.UUID, *lcdatabase.Account]
	accountCurrenciesCache     *ttlcache.Cache[accountCurrencyCacheKey, *lcdatabase.AccountCurrency]
}

func (t *transfers) CreateTransaction(ctx context.Context, requestMdc *logger.RequestMdc, transaction *TransactionRequest) (*TransactionResponse, error) {
	timedCtx, cancel := context.WithTimeout(ctx, t.createTransactionTimeoutMs)
	defer cancel()
	result, err := t.cb.Execute(func() (interface{}, error) {
		return t.createTransaction(timedCtx, requestMdc, transaction)
	})
	if err != nil {
		return nil, err
	}
	return result.(*TransactionResponse), nil
}

func (t *transfers) createTransaction(ctx context.Context, requestMdc *logger.RequestMdc, transaction *TransactionRequest) (*TransactionResponse, error) {
	sourceAccount, targetAccount, err := t.getValidatedAccounts(ctx, requestMdc, transaction)
	if err != nil {
		return nil, err
	}
	sourceAccountCurrency, targetAccountCurrency, err := t.getValidatedAccountCurrencies(ctx, requestMdc, transaction, sourceAccount)
	if err != nil {
		return nil, err
	}

	// just because we don't allow cross-currency now - we put the value everywhere
	// otherwise we need - to obtain values for source/target using currency conversion
	transactionAmountInFractional := currency.GetAmountInFractional(transaction.Amount, transaction.Fractional, transaction.Currency)

	transactionCreateRequest := lcdatabase.TransactionCreateRequest{
		Id:          transaction.Id,
		SubsystemId: transaction.SubsystemId,
		Description: transaction.Description,
		Amount:      transactionAmountInFractional,
		Currency:    transaction.Currency,
		SourceOperation: lcdatabase.TransferAccountCurrency{
			AccountCurrencyId: sourceAccountCurrency.Id,
			Amount:            -transactionAmountInFractional,
			OperationType:     database.OPERATION_TYPE_OUT,
			OperationStatus:   getOperationStatus(sourceAccount.AllowAsync),
			IsAsync:           sourceAccount.AllowAsync,
		},
		TargetOperation: lcdatabase.TransferAccountCurrency{
			AccountCurrencyId: targetAccountCurrency.Id,
			Amount:            transactionAmountInFractional,
			OperationType:     database.OPERATION_TYPE_IN,
			OperationStatus:   database.OPERATION_STATUS_CREATED,
			IsAsync:           true,
		},
	}

	dbTransaction, transactionError := t.transfersRepository.CreateTransaction(ctx, requestMdc, &transactionCreateRequest)

	if transactionError != nil {

		if errors.Is(transactionError, lcdatabase.ErrInsufficientAmount) {
			return nil, ErrInsufficientAmount
		}
		if errors.Is(transactionError, lcdatabase.ErrDuplicateTransaction) {
			return t.getTransactionDetails(ctx, requestMdc, transaction, &transactionCreateRequest, sourceAccountCurrency, targetAccountCurrency)
		}
		if errors.Is(transactionError, context.DeadlineExceeded) {
			logger.LogError(t.log, requestMdc, "error creating transaction", transactionError)
			increment(t.dd, "error", requestMdc.Subsystem, "reason:timeout", "operation:create")
			return nil, ErrInternal
		}

		logger.LogError(t.log, requestMdc, "error creating transaction", transactionError)
		increment(t.dd, "error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", transactionError), "operation:create")
		return nil, ErrInternal

	}
	// init workflows to update account balance using transaction.SourceOperation.Status/transaction.TargetOperation.Status
	// fire & forget way
	go t.initBalanceUpdate(dbTransaction, sourceAccount.SubsystemId, targetAccount.SubsystemId)
	increment(t.dd, "success", transaction.SubsystemId)

	return t.mapToTransactionResponse(ctx, requestMdc, transaction, dbTransaction, sourceAccountCurrency, targetAccountCurrency)
}

func (t *transfers) getTransactionDetails(ctx context.Context, requestMdc *logger.RequestMdc,
	transaction *TransactionRequest,
	transactionCreateRequest *lcdatabase.TransactionCreateRequest,
	sourceAccountCurrency *lcdatabase.AccountCurrency,
	targetAccountCurrency *lcdatabase.AccountCurrency) (*TransactionResponse, error) {
	dbTransaction, err := t.transfersRepository.GetTransactionDetails(ctx, requestMdc, transaction.Id)
	if err == nil {
		if transactionsEqual(dbTransaction, transactionCreateRequest) {
			return t.mapToTransactionResponse(ctx, requestMdc, transaction, dbTransaction, sourceAccountCurrency, targetAccountCurrency)
		}
		logger.LogInfo(t.log, requestMdc, "transaction details conflict")
		increment(t.dd, "fail", requestMdc.Subsystem, "reason:conflict")
		return nil, ErrTransactionDetailsConflict
	}
	if errors.Is(err, context.DeadlineExceeded) {
		logger.LogError(t.log, requestMdc, "error getting transaction", err)
		increment(t.dd, "error", requestMdc.Subsystem, "reason:timeout", "operation:get")
	} else {
		logger.LogError(t.log, requestMdc, "error getting transaction details", err)
		increment(t.dd, "error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err), "operation:get")
	}
	return nil, ErrInternal
}

func (t *transfers) mapToTransactionResponse(ctx context.Context, requestMdc *logger.RequestMdc,
	transaction *TransactionRequest,
	dbTransaction *lcdatabase.Transaction,
	sourceAccountCurrency *lcdatabase.AccountCurrency,
	targetAccountCurrency *lcdatabase.AccountCurrency) (*TransactionResponse, error) {

	transactionAmount, transactionAmountFractional := currency.GetAmountFromFractional(dbTransaction.Amount, dbTransaction.Currency)
	sourceAmount, sourceFractional := currency.GetAmountFromFractional(dbTransaction.SourceOperation.Amount, sourceAccountCurrency.Currency)
	targetAmount, targetFractional := currency.GetAmountFromFractional(dbTransaction.TargetOperation.Amount, targetAccountCurrency.Currency)

	return &TransactionResponse{
		Id:          dbTransaction.Id,
		Time:        dbTransaction.Time,
		Amount:      transactionAmount,
		Fractional:  transactionAmountFractional,
		Currency:    dbTransaction.Currency,
		SubsystemId: dbTransaction.SubsystemId,
		Description: dbTransaction.Description,
		SourceOperation: OperationResponse{
			Id:         dbTransaction.SourceOperation.Id,
			Type:       dbTransaction.SourceOperation.Type,
			Status:     dbTransaction.SourceOperation.Status,
			Amount:     sourceAmount,
			Fractional: sourceFractional,
		},
		TargetOperation: OperationResponse{
			Id:         dbTransaction.TargetOperation.Id,
			Type:       dbTransaction.TargetOperation.Type,
			Status:     dbTransaction.TargetOperation.Status,
			Amount:     targetAmount,
			Fractional: targetFractional,
		}}, nil
}

func (t *transfers) initBalanceUpdate(dbTransaction *lcdatabase.Transaction, sourceAccountSubsystemId uuid.UUID, targetAccountSubsystemId uuid.UUID) {
	if dbTransaction.SourceOperation.Status == database.OPERATION_STATUS_CREATED {
		t.updater.UpdateBalance(database.OperationStatusTransition{
			AccountSubsystemId: sourceAccountSubsystemId,
			Id:                 dbTransaction.SourceOperation.Id,
			AccountCurrencyId:  dbTransaction.SourceOperation.AccountCurrencyId,
			From:               database.OPERATION_STATUS_CREATED,
			To:                 database.OPERATION_STATUS_POSTED,
			OperationType:      dbTransaction.SourceOperation.Type,
			Timestamp:          time.Now(),
		})
	}

	t.updater.UpdateBalance(database.OperationStatusTransition{
		AccountSubsystemId: targetAccountSubsystemId,
		Id:                 dbTransaction.TargetOperation.Id,
		AccountCurrencyId:  dbTransaction.TargetOperation.AccountCurrencyId,
		From:               database.OPERATION_STATUS_CREATED,
		To:                 database.OPERATION_STATUS_POSTED,
		OperationType:      dbTransaction.TargetOperation.Type,
		Timestamp:          time.Now(),
	})
}

func (t *transfers) getAccount(ctx context.Context, requestMdc *logger.RequestMdc, accountId uuid.UUID) (*lcdatabase.Account, error) {

	if cached := t.accountsCache.Get(accountId); cached != nil {
		increment(t.dd, "account.cache", requestMdc.Subsystem, "result:ok")
		return cached.Value(), nil
	}
	increment(t.dd, "account.cache", requestMdc.Subsystem, "result:not")

	account, err := t.accountRepository.GetAccount(ctx, accountId)
	if err != nil {
		if err == lcdatabase.ErrNotFound {
			logger.LogInfo(t.log, requestMdc, fmt.Sprintf("account not found: %v, failing transaction", accountId))
			increment(t.dd, "fail", requestMdc.Subsystem, "reason:not_found")
			return nil, ErrAccountNotFound
		}
		logger.LogError(t.log, requestMdc, "error from db", err)
		increment(t.dd, "error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, ErrInternal
	}
	if account.IsSuspended {
		logger.LogInfo(t.log, requestMdc, fmt.Sprintf("account suspended: %v, failing transaction", accountId))
		increment(t.dd, "fail", requestMdc.Subsystem, "reason:suspended")
		return nil, ErrAccountSuspended
	}

	t.accountsCache.Set(accountId, account, ttlcache.DefaultTTL)
	return account, nil
}

func (t *transfers) getValidatedAccounts(ctx context.Context, requestMdc *logger.RequestMdc, transaction *TransactionRequest) (*lcdatabase.Account, *lcdatabase.Account, error) {

	if transaction.SourceAccountInfo.Id == transaction.TargetAccountInfo.Id {
		logger.LogInfo(t.log, requestMdc, fmt.Sprintf("source account %v is equal to target account, failing transaction", transaction.SourceAccountInfo.Id))
		increment(t.dd, "fail", requestMdc.Subsystem, "reason:source_equal_target")
		return nil, nil, ErrSameAccount
	}

	// support only same currency transfers for this moment
	if !(transaction.Currency == transaction.SourceAccountInfo.Currency && transaction.Currency == transaction.TargetAccountInfo.Currency) {
		logger.LogInfo(t.log, requestMdc, fmt.Sprintf("cross currency (%s, %s, %s), failing transaction between %v %v",
			transaction.Currency,
			transaction.SourceAccountInfo.Currency,
			transaction.TargetAccountInfo.Currency,
			transaction.SourceAccountInfo.Id,
			transaction.TargetAccountInfo.Id,
		))
		increment(t.dd, "fail", requestMdc.Subsystem, "reason:cross_currency")
		return nil, nil, ErrCrossCurrency
	}

	sourceAccount, err := t.getAccount(ctx, requestMdc, transaction.SourceAccountInfo.Id)
	if err != nil {
		return nil, nil, err
	}
	targetAccount, err := t.getAccount(ctx, requestMdc, transaction.TargetAccountInfo.Id)
	if err != nil {
		return nil, nil, err
	}

	return sourceAccount, targetAccount, nil

}

func (t *transfers) getAccountCurrency(ctx context.Context, requestMdc *logger.RequestMdc, accountId uuid.UUID, currency string, allowFromCache bool) (*lcdatabase.AccountCurrency, error) {

	key := accountCurrencyCacheKey{accountId, currency}
	if allowFromCache {
		if cachedAccountCurrency := t.accountCurrenciesCache.Get(key); cachedAccountCurrency != nil {
			increment(t.dd, "account-currency.cache", requestMdc.Subsystem, "result:ok")
			return cachedAccountCurrency.Value(), nil
		}
	}
	increment(t.dd, "account-currency.cache", requestMdc.Subsystem, "result:not")

	accountCurrency, err := t.accountRepository.GetAccountCurrency(ctx, accountId, currency)
	if err != nil {
		if err == lcdatabase.ErrNotFound {
			logger.LogInfo(t.log, requestMdc, fmt.Sprintf("account %v has no currency: %s, failing transaction", accountId, currency))
			increment(t.dd, "fail", requestMdc.Subsystem, "reason:no_currency")
			return nil, ErrNoCurrency
		}
		logger.LogError(t.log, requestMdc, "error getting currency", err)
		increment(t.dd, "error", requestMdc.Subsystem, fmt.Sprintf("reason:%v", err))
		return nil, ErrInternal
	}

	t.accountCurrenciesCache.Set(key, accountCurrency, ttlcache.DefaultTTL)
	return accountCurrency, nil
}

func (t *transfers) getValidatedAccountCurrencies(ctx context.Context, requestMdc *logger.RequestMdc, transaction *TransactionRequest, source *lcdatabase.Account) (*lcdatabase.AccountCurrency, *lcdatabase.AccountCurrency, error) {

	sourceAccountCurrency, err := t.getAccountCurrency(ctx, requestMdc, transaction.SourceAccountInfo.Id, transaction.SourceAccountInfo.Currency, source.AllowAsync)
	if err != nil {
		return nil, nil, err
	}

	if sourceAccountCurrency.CreditLimit != nil && sourceAccountCurrency.Amount-transaction.Amount < *sourceAccountCurrency.CreditLimit {
		logger.LogInfo(t.log, requestMdc, fmt.Sprintf("account %v reached credit limit: %d, failing transaction", transaction.SourceAccountInfo.Id, *sourceAccountCurrency.CreditLimit))
		increment(t.dd, "fail", requestMdc.Subsystem, "reason:credit_limit")
		return nil, nil, ErrInsufficientAmount
	}

	targetAccountCurrency, err := t.getAccountCurrency(ctx, requestMdc, transaction.TargetAccountInfo.Id, transaction.TargetAccountInfo.Currency, true)
	if err != nil {
		return nil, nil, err
	}

	return sourceAccountCurrency, targetAccountCurrency, nil
}

func getOperationStatus(allowAsync bool) string {
	if allowAsync {
		return database.OPERATION_STATUS_CREATED
	}
	return database.OPERATION_STATUS_POSTED
}

func transactionsEqual(dbTransaction *lcdatabase.Transaction, requestTransaction *lcdatabase.TransactionCreateRequest) bool {
	return requestTransaction.SubsystemId == dbTransaction.SubsystemId &&
		requestTransaction.SourceOperation.AccountCurrencyId == dbTransaction.SourceOperation.AccountCurrencyId &&
		requestTransaction.TargetOperation.AccountCurrencyId == dbTransaction.TargetOperation.AccountCurrencyId &&
		requestTransaction.Description == dbTransaction.Description &&
		requestTransaction.SourceOperation.OperationType == dbTransaction.SourceOperation.Type &&
		requestTransaction.TargetOperation.OperationType == dbTransaction.TargetOperation.Type &&
		requestTransaction.Amount == dbTransaction.Amount &&
		requestTransaction.Currency == dbTransaction.Currency
}

func NewTransfers(p TransfersParameters) Transfers {

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        "Create Transaction Circuit Breaker",
		MaxRequests: p.Configuration.Domain.Transfer.CreateTransaction.CircuitBreaker.MaxRequest,
		Interval:    time.Duration(p.Configuration.Domain.Transfer.CreateTransaction.CircuitBreaker.IntervalMs * time.Millisecond),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > p.Configuration.Domain.Transfer.CreateTransaction.CircuitBreaker.ConsecutiveFailures
		},
		IsSuccessful: func(err error) bool {
			return !errors.Is(err, ErrInternal)
		},
		Timeout: time.Duration(p.Configuration.Domain.Transfer.CreateTransaction.CircuitBreaker.TimeoutMs * time.Millisecond),
		OnStateChange: func(name string, from, to gobreaker.State) {
			p.Log.Debug(fmt.Sprintf("%s changed state from %v to %v", name, from, to))
			p.Datadog.Increment("transfer.transaction.cicruit-breaker", fmt.Sprintf("state:%v", to))
		},
	})

	accountsCache := ttlcache.New[uuid.UUID, *lcdatabase.Account](
		ttlcache.WithTTL[uuid.UUID, *lcdatabase.Account](p.Configuration.Domain.Transfer.AccountsCache.TtlSeconds*time.Second),
		ttlcache.WithCapacity[uuid.UUID, *lcdatabase.Account](p.Configuration.Domain.Transfer.AccountsCache.Capacity),
		ttlcache.WithDisableTouchOnHit[uuid.UUID, *lcdatabase.Account](),
	)

	accountCurrenciesCache := ttlcache.New[accountCurrencyCacheKey, *lcdatabase.AccountCurrency](
		ttlcache.WithTTL[accountCurrencyCacheKey, *lcdatabase.AccountCurrency](p.Configuration.Domain.Transfer.AccountsCache.TtlSeconds*time.Second),
		ttlcache.WithCapacity[accountCurrencyCacheKey, *lcdatabase.AccountCurrency](p.Configuration.Domain.Transfer.AccountsCache.Capacity),
		ttlcache.WithDisableTouchOnHit[accountCurrencyCacheKey, *lcdatabase.AccountCurrency](),
	)

	go accountsCache.Start()
	go accountCurrenciesCache.Start()

	createTransactionTimeoutMs := time.Duration(p.Configuration.Domain.Transfer.CreateTransaction.TimeoutMs * time.Millisecond)
	transfers := transfers{p.Log.Named("transfers"),
		p.TransfersRepository,
		p.AccountRepository,
		p.Updater,
		p.Datadog,
		createTransactionTimeoutMs,
		cb,
		accountsCache,
		accountCurrenciesCache,
	}
	return &transfers
}
