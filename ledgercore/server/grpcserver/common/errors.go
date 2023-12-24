package common

import (
	"github.com/gogo/status"
	"google.golang.org/grpc/codes"
)

var ErrNoMeta = status.Error(codes.FailedPrecondition, "no metadata in request context")
var ErrNoSubsystem = status.Error(codes.Unauthenticated, "no subsystem in request metadata")
var ErrUnparseableUuid = status.Error(codes.InvalidArgument, "could not parse UUID")

var ErrInternal = status.Error(codes.Internal, "internal error")
