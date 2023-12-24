package grpcserver

import (
	"context"
	"fmt"
	"ledgercore/config"
	"ledgercore/server/grpcserver/account"
	"ledgercore/server/grpcserver/transfer"
	"net"
	pbaccount "shared/protocol/account"
	pbtransfer "shared/protocol/transfer"

	grpctrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/google.golang.org/grpc"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type GrpcServer interface {
	Stop()
}

type grpcServer struct {
	server *grpc.Server
	log    *zap.Logger
}

func (s *grpcServer) Stop() {
	s.log.Info("stop grpc server")
	s.server.Stop()
}

type GrpcServerParameters struct {
	fx.In
	Log            *zap.Logger
	Configuration  config.Configuration
	Lifecycle      fx.Lifecycle
	TransferServer pbtransfer.TransferServer
	AccountServer  pbaccount.AccountServer
}

func NewGrpcServer(p GrpcServerParameters) (GrpcServer, error) {
	grpcServer := grpcServer{
		log: p.Log.Named("grpc server"),
	}

	if p.Configuration.Datadog.Enabled {
		ui := grpctrace.UnaryServerInterceptor(
			grpctrace.WithServiceName(p.Configuration.Application.Name),
			grpctrace.WithCustomTag("version", config.APPLICATION_VERSION),
		)
		grpcServer.server = grpc.NewServer(grpc.UnaryInterceptor(ui))
	} else {
		grpcServer.server = grpc.NewServer()
	}

	pbtransfer.RegisterTransferServer(grpcServer.server, p.TransferServer)
	pbaccount.RegisterAccountServer(grpcServer.server, p.AccountServer)

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			lis, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Configuration.Server.Port))
			if err != nil {
				grpcServer.log.Error("failed to listen", zap.Error(err))
				return err
			}
			grpcServer.log.Info(fmt.Sprintf("server listening at %v", lis.Addr()))
			go func() {
				if err := grpcServer.server.Serve(lis); err != nil {
					grpcServer.log.Error("failed to serve", zap.Error(err))
				}
			}()
			return nil
		},
	})
	return &grpcServer, nil
}

// GrpcServer is an aggregation of gRPC services provided by application
var GrpcServerModule = fx.Module("grpc-server", fx.Provide(account.NewAccountsServer, transfer.NewTransferServer, NewGrpcServer))
