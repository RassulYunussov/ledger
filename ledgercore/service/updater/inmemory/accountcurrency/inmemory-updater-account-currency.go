package accountcurrency

import (
	"context"
	"fmt"
	"ledgercore/config"
	"math/rand"
	"shared/service/database"
	"shared/service/infrastructure/datadog"
	"sync"
	"time"

	"go.uber.org/zap"
)

const UPDATE_BALANCE_TIMEOUT_SECONDS = time.Second * 2

type AccountCurrencyInMemoryUpdater interface {
	UpdateBalance(operation database.OperationStatusTransition, idx int)
	Stop()
}

type accountCurrencyInMemoryUpdater struct {
	log *zap.Logger
	// operation type -> subsystem -> channel
	accountCurrencyOperationsChans        [][][]chan database.OperationStatusTransition
	workersChan                           chan bool
	accountCurrencyRepository             database.AccountCurrencyRepository
	dd                                    datadog.Datadog
	accountCurrencyOperationsWaitGroup    sync.WaitGroup
	getAccountCurrencyOperationsChannelId func(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint)
}

func (inMemory *accountCurrencyInMemoryUpdater) Stop() {
	inMemory.log.Info("stop account currency updater initiated")
	for k := 0; k < len(inMemory.accountCurrencyOperationsChans); k++ {
		for j := 0; j < len(inMemory.accountCurrencyOperationsChans[k]); j++ {
			for i := 0; i < len(inMemory.accountCurrencyOperationsChans[k][j]); i++ {
				close(inMemory.accountCurrencyOperationsChans[k][j][i])
			}
		}
	}
	inMemory.accountCurrencyOperationsWaitGroup.Wait()
	inMemory.log.Info("account currency updater finished processing, closing workers channel")
	close(inMemory.workersChan)
}

func (inMemory *accountCurrencyInMemoryUpdater) getOperationTypeChanelId(operationType string) uint {
	switch operationType {
	case database.OPERATION_TYPE_OUT:
		return 0
	case database.OPERATION_TYPE_IN:
		return 1
	default:
		return uint(rand.Intn(len(database.GetOperationTypes())))
	}
}

func (inMemory *accountCurrencyInMemoryUpdater) UpdateBalance(operation database.OperationStatusTransition, idx int) {
	current := time.Now()
	diff := current.UnixNano()/int64(time.Millisecond) - operation.Timestamp.UnixNano()/int64(time.Millisecond)
	inMemory.dd.Timing("operation.dequeue", time.Duration(diff), fmt.Sprintf("channel:%d", idx))
	operationTypeChanelId := inMemory.getOperationTypeChanelId(operation.OperationType)
	subsystemId, channelId := inMemory.getAccountCurrencyOperationsChannelId(operation, inMemory.accountCurrencyOperationsChans[operationTypeChanelId])
	inMemory.log.Debug(fmt.Sprintf("enqueue operation %d from channel %d into accunt currencies channel %d %d %d", operation.Id, idx, operationTypeChanelId, subsystemId, channelId))
	inMemory.dd.Increment("acoperation.enqueue", fmt.Sprintf("channel:%d", channelId), fmt.Sprintf("subsystem:%d", subsystemId), fmt.Sprintf("type:%d", operationTypeChanelId))
	operation.Timestamp = current
	inMemory.accountCurrencyOperationsChans[operationTypeChanelId][subsystemId][channelId] <- operation
}

func (inMemory *accountCurrencyInMemoryUpdater) processAccountCurrencyOperations(channel <-chan database.OperationStatusTransition, subsystemIdx int, idx int) {
	for operation := range channel {
		current := time.Now()
		diff := time.Duration(current.UnixNano()/int64(time.Millisecond) - operation.Timestamp.UnixNano()/int64(time.Millisecond))
		inMemory.dd.Timing("acoperation.dequeue", time.Duration(diff), fmt.Sprintf("channel:%d", idx), fmt.Sprintf("subsystem:%d", subsystemIdx))
		inMemory.log.Debug(fmt.Sprintf("processing account currency operation %d form channel %d subsystem %d", operation.Id, idx, subsystemIdx))
		inMemory.workersChan <- true
		go inMemory.update(operation)
	}
	inMemory.accountCurrencyOperationsWaitGroup.Done()
}

func (inMemory *accountCurrencyInMemoryUpdater) update(operation database.OperationStatusTransition) {
	start := time.Now().UnixNano() / int64(time.Millisecond)
	secondContext, cancel := context.WithTimeout(context.Background(), UPDATE_BALANCE_TIMEOUT_SECONDS)
	defer cancel()
	err := inMemory.accountCurrencyRepository.UpdateBalance(secondContext, operation)
	if err == nil {
		end := time.Now().UnixNano() / int64(time.Millisecond)
		diff := end - start
		inMemory.dd.Timing("operation.update.time", time.Duration(diff), fmt.Sprintf("ac:%v", operation.AccountCurrencyId), fmt.Sprintf("subsystem:%v", operation.AccountSubsystemId))
	} else {
		inMemory.dd.Increment("operation.update.error", fmt.Sprintf("reason:%v", err), fmt.Sprintf("subsystem:%v", operation.AccountSubsystemId))
		inMemory.log.Info("error updating balance", zap.Error(err))
	}
	<-inMemory.workersChan
}

func CreateAccountCurrencyInMemoryUpdater(log *zap.Logger,
	dd datadog.Datadog,
	configuration config.Configuration,
	accountCurrencyRepository database.AccountCurrencyRepository) AccountCurrencyInMemoryUpdater {
	updater := accountCurrencyInMemoryUpdater{
		log:                       log.Named("account-currency-inmemory-updater"),
		dd:                        dd,
		accountCurrencyRepository: accountCurrencyRepository,
	}

	updater.accountCurrencyOperationsChans = make([][][]chan database.OperationStatusTransition, len(database.GetOperationTypes()))
	for k := 0; k < len(database.GetOperationTypes()); k++ {
		updater.accountCurrencyOperationsChans[k] = make([][]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Subsystems)
		for j := 0; j < configuration.InMemory.AccountCurrencies.Subsystems; j++ {
			updater.accountCurrencyOperationsChans[k][j] = make([]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Queues)
			for i := 0; i < configuration.InMemory.AccountCurrencies.Queues; i++ {
				updater.accountCurrencyOperationsChans[k][j][i] = make(chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Buffer)
			}
		}
	}
	var selectedStrategy string
	updater.getAccountCurrencyOperationsChannelId, selectedStrategy = getAccountCurrencyOperationsChannelSelector(configuration.InMemory.AccountCurrencies.Strategy)
	updater.workersChan = make(chan bool, configuration.InMemory.Workers)

	updater.log.Info(fmt.Sprintf("start account currency updater with selector %s", selectedStrategy))
	for j := 0; j < len(updater.accountCurrencyOperationsChans); j++ {
		updater.accountCurrencyOperationsWaitGroup.Add(len(updater.accountCurrencyOperationsChans[j]))
	}
	for operationTypeIdx, operationTypeChanel := range updater.accountCurrencyOperationsChans {
		for subsystemIdx, subsystemChans := range operationTypeChanel {
			for idx, ch := range subsystemChans {
				updater.log.Info(fmt.Sprintf("start listening from operation type %d, sybsystem %d, account currency channel %d", subsystemIdx, operationTypeIdx, idx))
				go updater.processAccountCurrencyOperations(ch, subsystemIdx, idx)
			}
		}
	}
	return &updater
}
