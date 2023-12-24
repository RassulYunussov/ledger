package common

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func GetSubsystem(ctx context.Context) (*uuid.UUID, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoMeta
	}
	subsystemIds, ok := meta["subsystem"]
	if !ok || len(subsystemIds) != 1 {
		return nil, ErrNoSubsystem
	}
	subsystemId, err := uuid.Parse(subsystemIds[0])
	if err != nil {
		return nil, ErrUnparseableUuid
	}
	return &subsystemId, nil
}
