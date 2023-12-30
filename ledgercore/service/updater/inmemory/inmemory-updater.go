package inmemory

import (
	"context"
	"fmt"
	"ledgercore/config"
	"ledgercore/service/updater/inmemory/accountcurrency"
	"shared/service/database"
	"shared/service/infrastructure/datadog"
	"sync"
	"time"

	"go.uber.org/zap"
)

const outpostReadTimeout = time.Minute

type inMemoryUpdater struct {
	accountCurrencyInMemoryUpdater accountcurrency.AccountCurrencyInMemoryUpdater
	log                            *zap.Logger
	operationsChans                []chan database.OperationStatusTransition
	dd                             datadog.Datadog
	operationsWaitGroup            sync.WaitGroup
	accountCurrencyRepository      database.AccountCurrencyRepository
}

func (inMemory *inMemoryUpdater) UpdateBalance(operation database.OperationStatusTransition) {
	inMemory.log.Debug("initiated update balance")
	channelId := operation.Id % int64(len(inMemory.operationsChans))
	inMemory.dd.Increment("operation.enqueue", fmt.Sprintf("channel:%d", channelId))
	inMemory.operationsChans[channelId] <- operation
}

func (inMemory *inMemoryUpdater) Stop() {
	inMemory.log.Info("stop updater initiated")
	inMemory.log.Info("closing operations channels")
	for i := 0; i < len(inMemory.operationsChans); i++ {
		close(inMemory.operationsChans[i])
	}
	inMemory.operationsWaitGroup.Wait()
	inMemory.accountCurrencyInMemoryUpdater.Stop()
}

func (inMemory *inMemoryUpdater) startPeriodicUpdater(intervalSeconds int, offsetSeconds int) {
	inMemory.log.Info(fmt.Sprintf("starting account operations outpost processing with %ds interval and %ds time offset", intervalSeconds, offsetSeconds))
	for {
		time.Sleep(time.Second * time.Duration(intervalSeconds))
		inMemory.log.Debug("processing account operations outpost")
		secondsBeforeCurrentTime := offsetSeconds
		minuteContext, cancel := context.WithTimeout(context.Background(), outpostReadTimeout)
		operationIds, err := inMemory.accountCurrencyRepository.GetOperationsFromOutpost(minuteContext, secondsBeforeCurrentTime)
		if err != nil {
			inMemory.log.Warn("error fetching operations from account operation outpost", zap.Error(err))
			cancel()
			continue
		}
		inMemory.log.Debug(fmt.Sprintf("fetched %d to process", len(operationIds)))
		for _, operationId := range operationIds {
			secondContext, scancel := context.WithTimeout(minuteContext, accountcurrency.UPDATE_BALANCE_TIMEOUT_SECONDS)
			err := inMemory.accountCurrencyRepository.UpdateBalance(secondContext, database.OperationStatusTransition{
				Id:   operationId,
				From: database.OPERATION_STATUS_CREATED,
				To:   database.OPERATION_STATUS_POSTED,
			})
			scancel()
			if err != nil {
				//handle error
				inMemory.log.Info(fmt.Sprintf("error updating balance for opperation %d", operationId), zap.Error(err))
			}
		}
		cancel()
	}
}

func (inMemory *inMemoryUpdater) processOperations(channel <-chan database.OperationStatusTransition, idx int) {
	for operation := range channel {
		inMemory.accountCurrencyInMemoryUpdater.UpdateBalance(operation, idx)
	}
	inMemory.operationsWaitGroup.Done()
}

func CreateInMemoryBalanceUpdater(log *zap.Logger,
	accountCurrencyRepository database.AccountCurrencyRepository,
	configuration config.Configuration,
	dd datadog.Datadog) (*inMemoryUpdater, error) {

	operationsChans := make([]chan database.OperationStatusTransition, configuration.InMemory.Operations.Queues)
	for i := 0; i < configuration.InMemory.Operations.Queues; i++ {
		operationsChans[i] = make(chan database.OperationStatusTransition, configuration.InMemory.Operations.Buffer)
	}

	updater := inMemoryUpdater{
		accountcurrency.CreateAccountCurrencyInMemoryUpdater(log, dd, configuration, accountCurrencyRepository),
		log.Named("inmemory updater"),
		operationsChans,
		dd,
		sync.WaitGroup{},
		accountCurrencyRepository,
	}

	updater.log.Info("start updater")
	updater.operationsWaitGroup.Add(len(updater.operationsChans))
	for idx, ch := range updater.operationsChans {
		updater.log.Info(fmt.Sprintf("start listening from operation channel %d", idx))
		go updater.processOperations(ch, idx)
	}
	go updater.startPeriodicUpdater(configuration.InMemory.Outpost.IntervalSeconds, configuration.InMemory.Outpost.OffsetSeconds)

	return &updater, nil
}
