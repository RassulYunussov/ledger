package transfer

import (
	"context"
	"errors"
	"fmt"
	"ledgercore/domain/transfers"
	dtransfers "ledgercore/domain/transfers"
	"ledgercore/service/grpcserver/common"
	"shared/logger"
	"shared/protocol/transfer"
	"shared/service/infrastructure/datadog"

	"github.com/gogo/status"
	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type transferServer struct {
	transfer.UnimplementedTransferServer
	log       *zap.Logger
	transfers transfers.Transfers
	dd        datadog.Datadog
}

func (s *transferServer) mapToServiceError(requestMdc *logger.RequestMdc, err error) error {
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		logger.LogError(s.log, requestMdc, "error creating transaction", err)
		return status.Error(codes.ResourceExhausted, err.Error())
	}
	switch err {
	case dtransfers.ErrSameAccount:
		fallthrough
	case dtransfers.ErrCrossCurrency:
		fallthrough
	case dtransfers.ErrNoCurrency:
		fallthrough
	case dtransfers.ErrAccountNotFound:
		fallthrough
	case dtransfers.ErrAccountSuspended:
		fallthrough
	case dtransfers.ErrTransactionDetailsConflict:
		fallthrough
	case dtransfers.ErrInsufficientAmount:
		logger.LogWarn(s.log, requestMdc, "failing transaction", err)
		return status.Error(codes.InvalidArgument, err.Error())
	case dtransfers.ErrInternal:
		fallthrough
	default:
		return common.ErrInternal
	}
}

func (s *transferServer) CreateTransaction(ctx context.Context, request *transfer.TransactionRequest) (*transfer.TransactionResponse, error) {
	subsystemId, err := common.GetSubsystem(ctx)
	if err != nil {
		s.log.Warn("error getting subsystem from request", zap.Error(err))
		return nil, err
	}
	requestMdc := logger.RequestMdc{
		Subsystem: *subsystemId,
		Request:   uuid.UUID(request.Id),
	}
	logger.LogDebug(s.log, &requestMdc, "recieved request")
	storedTransaction, err := s.transfers.CreateTransaction(ctx, &requestMdc, &transfers.TransactionRequest{
		Id:          uuid.UUID(request.Id),
		SubsystemId: *subsystemId,
		Amount:      request.Amount,
		Fractional:  request.Fractional,
		Currency:    request.Currency,
		SourceAccountInfo: transfers.TransferAccountInfoRequest{
			Id:       uuid.UUID(request.Source),
			Currency: request.Currency,
		},
		TargetAccountInfo: transfers.TransferAccountInfoRequest{
			Id:       uuid.UUID(request.Target),
			Currency: request.Currency,
		},
		Description: request.Description,
	})

	if err != nil {
		return nil, s.mapToServiceError(&requestMdc, err)
	}

	logger.LogDebug(s.log, &requestMdc, fmt.Sprintf("processed request %v", storedTransaction.Id))

	return &transfer.TransactionResponse{
		Status: 0,
		Details: &transfer.TransactionDetailsResponse{
			Id: storedTransaction.Id[:],
			Time: &timestamppb.Timestamp{
				Seconds: storedTransaction.Time.Unix(),
			},
			Amount:      storedTransaction.Amount,
			Fractional:  storedTransaction.Fractional,
			Description: storedTransaction.Description,
			Currency:    storedTransaction.Currency,
			SourceOperation: &transfer.OperationDetailsResponse{
				Id:         storedTransaction.SourceOperation.Id,
				Type:       storedTransaction.SourceOperation.Type,
				Status:     storedTransaction.SourceOperation.Status,
				Amount:     storedTransaction.SourceOperation.Amount,
				Fractional: storedTransaction.Fractional,
			},
			TargetOperation: &transfer.OperationDetailsResponse{
				Id:         storedTransaction.TargetOperation.Id,
				Type:       storedTransaction.TargetOperation.Type,
				Status:     storedTransaction.TargetOperation.Status,
				Amount:     storedTransaction.TargetOperation.Amount,
				Fractional: storedTransaction.Fractional,
			},
		},
	}, nil
}

func NewTransferServer(log *zap.Logger, transfers transfers.Transfers, dd datadog.Datadog) transfer.TransferServer {
	return &transferServer{log: log.Named("transfer-server"), transfers: transfers, dd: dd}
}
