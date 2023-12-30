package accountcurrency

import (
	"encoding/binary"
	"math"
	"math/rand"
	"shared/service/database"

	"github.com/google/uuid"
)

const (
	weighted   = "weighted"
	roundRobin = "round-robin"
	random     = "random"
	_default   = "default"
)

var strategies = []string{weighted, roundRobin, _default}

func getAccountCurrencyOperationsChannelSelector(strategy string) (func(ost database.OperationStatusTransition, subsystemChannels []subsystemChannel) (uint, uint), string) {
	switch strategy {
	case random:
		return getAccountCurrencyOperationsChannelSelector(strategies[rand.Intn(3)])
	case weighted:
		return getAccountCurrencyOperationsChannelWeighted, weighted
	case roundRobin:
		return getAccountCurrencyOperationsChannelRoundRobin, roundRobin
	default:
		return getAccountCurrencyOperationsChannelByUuid, _default
	}
}

func getAccountCurrencyOperationsChannelWeighted(ost database.OperationStatusTransition, subsystemChannels []subsystemChannel) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(subsystemChannels)))
	min := math.MaxInt32
	channelId := uint(0)
	for idx, ch := range subsystemChannels[subsystemId].accountCurrencyChannels {
		chLen := len(ch)
		if chLen < min {
			min = chLen
			channelId = uint(idx)
		}
	}
	return subsystemId, channelId
}

func getAccountCurrencyOperationsChannelRoundRobin(ost database.OperationStatusTransition, subsystemChannels []subsystemChannel) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(subsystemChannels)))
	return subsystemId, uint(ost.Id) % uint(len(subsystemChannels[subsystemId].accountCurrencyChannels))
}

func getAccountCurrencyOperationsChannelByUuid(ost database.OperationStatusTransition, subsystemChannels []subsystemChannel) (uint, uint) {
	subsystemId := selectSubsystem(ost.AccountSubsystemId, uint(len(subsystemChannels)))
	return subsystemId, getUIntFromUuid(ost.AccountCurrencyId) % uint(len(subsystemChannels[subsystemId].accountCurrencyChannels))
}

func selectSubsystem(subsystemId uuid.UUID, cap uint) uint {
	return getUIntFromUuid(subsystemId) % cap
}

func getUIntFromUuid(id uuid.UUID) uint {
	bytes := [16]byte(id)
	intValue, _ := binary.Varint(bytes[9:])
	return uint(intValue)
}
