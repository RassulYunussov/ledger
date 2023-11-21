package accountcurrency

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
	"go.uber.org/zap"
)

const UPDATE_BALANCE_TIMEOUT_SECONDS = time.Second * 2

type AccountCurrencyInMemoryUpdater interface {
	UpdateBalance(operation database.OperationStatusTransition, idx int)
	Stop()
}

type accountCurrencyInMemoryUpdater struct {
	log                                   *zap.Logger
	accountCurrencyOperationsChans        [][]chan database.OperationStatusTransition
	workersChan                           chan bool
	accountCurrencyRepository             database.AccountCurrencyRepository
	dd                                    datadog.Datadog
	accountCurrencyOperationsWaitGroup    sync.WaitGroup
	getAccountCurrencyOperationsChannelId func(ost database.OperationStatusTransition, accountCurrencyOperationsChans [][]chan database.OperationStatusTransition) (uint, uint)
}

func (inMemory *accountCurrencyInMemoryUpdater) Stop() {
	inMemory.log.Info("stop account currency updater initiated")
	for j := 0; j < len(inMemory.accountCurrencyOperationsChans); j++ {
		for i := 0; i < len(inMemory.accountCurrencyOperationsChans[j]); i++ {
			close(inMemory.accountCurrencyOperationsChans[j][i])
		}
	}
	inMemory.accountCurrencyOperationsWaitGroup.Wait()
	inMemory.log.Info("account currency updater finished processing, closing workers channel")
	close(inMemory.workersChan)
}

func (inMemory *accountCurrencyInMemoryUpdater) UpdateBalance(operation database.OperationStatusTransition, idx int) {
	current := time.Now()
	diff := current.UnixNano()/int64(time.Millisecond) - operation.Timestamp.UnixNano()/int64(time.Millisecond)
	inMemory.dd.Timing("operation.dequeue", time.Duration(diff), fmt.Sprintf("channel:%d", idx))
	subsystemId, channelId := inMemory.getAccountCurrencyOperationsChannelId(operation, inMemory.accountCurrencyOperationsChans)
	inMemory.log.Debug(fmt.Sprintf("enqueue operation %d form channel %d into accunt currencies channel %d", operation.Id, idx, channelId))
	inMemory.dd.Increment("acoperation.enqueue", fmt.Sprintf("channel:%d", channelId), fmt.Sprintf("subsystem:%d", subsystemId))
	operation.Timestamp = current
	inMemory.accountCurrencyOperationsChans[subsystemId][channelId] <- operation
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

func selectSubsystem(subsystemId uuid.UUID, cap uint) uint {
	return getUIntFromUuid(subsystemId) % cap
}

func getUIntFromUuid(id uuid.UUID) uint {
	bytes := [16]byte(id)
	intValue, _ := binary.Varint(bytes[9:])
	return uint(intValue)
}

func NewAccountCurrencyInMemoryUpdater(log *zap.Logger,
	dd datadog.Datadog,
	configuration config.Configuration,
	accountCurrencyRepository database.AccountCurrencyRepository) AccountCurrencyInMemoryUpdater {
	updater := accountCurrencyInMemoryUpdater{
		log:                       log.Named("account-currency-inmemory-updater"),
		dd:                        dd,
		accountCurrencyRepository: accountCurrencyRepository,
	}

	updater.accountCurrencyOperationsChans = make([][]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Subsystems)
	for j := 0; j < configuration.InMemory.AccountCurrencies.Subsystems; j++ {
		updater.accountCurrencyOperationsChans[j] = make([]chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Queues)
		for i := 0; i < configuration.InMemory.AccountCurrencies.Queues; i++ {
			updater.accountCurrencyOperationsChans[j][i] = make(chan database.OperationStatusTransition, configuration.InMemory.AccountCurrencies.Buffer)
		}
	}

	switch configuration.InMemory.AccountCurrencies.Strategy {
	case "weighted":
		updater.getAccountCurrencyOperationsChannelId = getAccountCurrencyOperationsChannelWeighted
	case "round-robin":
		updater.getAccountCurrencyOperationsChannelId = getAccountCurrencyOperationsChannelRoundRobin
	default:
		updater.getAccountCurrencyOperationsChannelId = getAccountCurrencyOperationsChannelByUuid
	}

	updater.workersChan = make(chan bool, configuration.InMemory.Workers)

	updater.log.Info("start account currency updater")
	for j := 0; j < len(updater.accountCurrencyOperationsChans); j++ {
		updater.accountCurrencyOperationsWaitGroup.Add(len(updater.accountCurrencyOperationsChans[j]))
	}
	for subsystemIdx, chans := range updater.accountCurrencyOperationsChans {
		for idx, ch := range chans {
			updater.log.Info(fmt.Sprintf("start listening from sybsystem %d account currency channel %d", subsystemIdx, idx))
			go updater.processAccountCurrencyOperations(ch, subsystemIdx, idx)
		}
	}

	return &updater
}
