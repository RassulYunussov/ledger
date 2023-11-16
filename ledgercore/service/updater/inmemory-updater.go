package updater

import (
	"context"
	"encoding/binary"
	"fmt"
	"ledgercore/config"
	"math"
	"shared/service/database"
	"shared/service/infrastructure/datadog"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const updateBalanceTimeoutSeconds = time.Second * 2
const outpostReadTimeout = time.Minute

type inMemoryUpdater struct {
	log                                   *zap.Logger
	operationsChans                       []chan database.OperationStatusTransition
	accountCurrencyOperationsChans        [][]chan database.OperationStatusTransition
	workersChan                           chan bool
	finished                              chan bool
	accountCurrencyRepository             database.AccountCurrencyRepository
	dd                                    datadog.Datadog
	operationsWaitGroup                   sync.WaitGroup
	accountCurrencyOperationsWaitGroup    sync.WaitGroup
	getAccountCurrencyOperationsChannelId func(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint)
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
	inMemory.log.Info("closing account currency operations channels")
	for j := 0; j < len(inMemory.accountCurrencyOperationsChans); j++ {
		for i := 0; i < len(inMemory.accountCurrencyOperationsChans[j]); i++ {
			close(inMemory.accountCurrencyOperationsChans[j][i])
		}
	}
	inMemory.accountCurrencyOperationsWaitGroup.Wait()
	<-inMemory.finished
	inMemory.log.Info("updater finished processing, closing workers channel")
	close(inMemory.workersChan)
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
			secondContext, scancel := context.WithTimeout(minuteContext, updateBalanceTimeoutSeconds)
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

func (inMemory *inMemoryUpdater) start() {
	inMemory.log.Info("start updater")
	for j := 0; j < len(inMemory.accountCurrencyOperationsChans); j++ {
		inMemory.accountCurrencyOperationsWaitGroup.Add(len(inMemory.accountCurrencyOperationsChans[j]))
	}
	for subsystemIdx, chans := range inMemory.accountCurrencyOperationsChans {
		for idx, ch := range chans {
			inMemory.log.Info(fmt.Sprintf("start listening from sybsystem %d account currency channel %d", subsystemIdx, idx))
			go inMemory.processAccountCurrencyOperations(ch, subsystemIdx, idx)
		}
	}
	inMemory.operationsWaitGroup.Add(len(inMemory.operationsChans))
	for idx, ch := range inMemory.operationsChans {
		inMemory.log.Info(fmt.Sprintf("start listening from operation channel %d", idx))
		go inMemory.processOperations(ch, idx)
	}
	inMemory.finished <- true
}

func (inMemory *inMemoryUpdater) processAccountCurrencyOperations(channel <-chan database.OperationStatusTransition, subsystemIdx int, idx int) {
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

func (inMemory *inMemoryUpdater) processOperations(channel <-chan database.OperationStatusTransition, idx int) {
	for operation := range channel {
		current := time.Now()
		diff := current.UnixNano()/int64(time.Millisecond) - operation.Timestamp.UnixNano()/int64(time.Millisecond)
		inMemory.dd.Timing("operation.dequeue", time.Duration(diff), fmt.Sprintf("channel:%d", idx))
		subsystemId, channelId := inMemory.getAccountCurrencyOperationsChannelId(operation, inMemory.accountCurrencyOperationsChans)
		inMemory.log.Debug(fmt.Sprintf("enqueue operation %d form channel %d into accunt currencies channel %d", operation.Id, idx, channelId))
		inMemory.dd.Increment("acoperation.enqueue", fmt.Sprintf("channel:%d", channelId), fmt.Sprintf("subsystem:%d", subsystemId))
		operation.Timestamp = current
		inMemory.accountCurrencyOperationsChans[subsystemId][channelId] <- operation
	}
	inMemory.operationsWaitGroup.Done()
}

func (inMemory *inMemoryUpdater) update(operation database.OperationStatusTransition) {
	start := time.Now().UnixNano() / int64(time.Millisecond)
	secondContext, cancel := context.WithTimeout(context.Background(), updateBalanceTimeoutSeconds)
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

func getUIntFromUuid(id uuid.UUID) uint {
	bytes := [16]byte(id)
	intValue, _ := binary.Varint(bytes[9:])
	return uint(intValue)
}

func selectSubsystem(subsystemId uuid.UUID, cap uint) uint {
	return getUIntFromUuid(subsystemId) % cap
}

func getAccountCurrencyOperationsChannelWeighted(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(accountCurrencyOperationsChans)))
	min := math.MaxInt32
	channelId := uint(0)
	for idx, ch := range accountCurrencyOperationsChans[subsystemId] {
		chLen := len(ch)
		if chLen < min {
			min = chLen
			channelId = uint(idx)
		}
	}
	return subsystemId, channelId
}

func getAccountCurrencyOperationsChannelRoundRobin(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(accountCurrencyOperationsChans)))
	return subsystemId, uint(ost.Id) % uint(len(accountCurrencyOperationsChans[subsystemId]))
}

func getAccountCurrencyOperationsChannelByUuid(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(accountCurrencyOperationsChans)))
	return subsystemId, getUIntFromUuid(ost.AccountCurrencyId) % uint(len(accountCurrencyOperationsChans[subsystemId]))
}

func createInMemoryBalanceUpdater(lifecycle fx.Lifecycle,
	log *zap.Logger,
	accountCurrencyRepository database.AccountCurrencyRepository,
	configuration config.Configuration,
	dd datadog.Datadog,
	strategy string) (Updater, error) {

	operationsChans := make([]chan database.OperationStatusTransition, configuration.InMemory.Operations.Queues)
	for i := 0; i < configuration.InMemory.Operations.Queues; i++ {
		operationsChans[i] = make(chan database.OperationStatusTransition, configuration.InMemory.Operations.Buffer)
	}

	accountCurrenciesChans := make([][]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Subsystems)
	for j := 0; j < configuration.InMemory.AccountCurrencies.Subsystems; j++ {
		accountCurrenciesChans[j] = make([]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Queues)
		for i := 0; i < configuration.InMemory.AccountCurrencies.Queues; i++ {
			accountCurrenciesChans[j][i] = make(chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Buffer)
		}
	}

	workersChan := make(chan bool, configuration.InMemory.Workers)
	finished := make(chan bool)

	var selector func(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint)

	switch strategy {
	case "weighted":
		selector = getAccountCurrencyOperationsChannelWeighted
	case "round-robin":
		selector = getAccountCurrencyOperationsChannelRoundRobin
	default:
		selector = getAccountCurrencyOperationsChannelByUuid
	}

	updater := inMemoryUpdater{log.Named("inmemory updater"),
		operationsChans,
		accountCurrenciesChans,
		workersChan,
		finished,
		accountCurrencyRepository,
		dd,
		sync.WaitGroup{},
		sync.WaitGroup{},
		selector,
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go updater.start()
			go updater.startPeriodicUpdater(configuration.InMemory.Outpost.IntervalSeconds, configuration.InMemory.Outpost.OffsetSeconds)
			return nil
		},
	})

	return &updater, nil

}
