package account

import (
	"context"

	daccount "ledgercore/domain/accounts"
	"ledgercore/server/grpcserver/common"
	"shared/protocol/account"
	"shared/service/infrastructure/datadog"

	"github.com/gogo/status"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
)

type accountServer struct {
	account.UnimplementedAccountServer
	log      *zap.Logger
	accounts daccount.Accounts
	dd       datadog.Datadog
}

func (s *accountServer) CreateAccount(ctx context.Context, createAccountRequest *account.CreateAccountRequest) (*account.CreateAccountResponse, error) {
	subsystemId, err := common.GetSubsystem(ctx)
	if err != nil {
		return nil, err
	}
	domainCurrencies := make([]daccount.CurrencyCreateRequest, 0)

	for _, requestCurency := range createAccountRequest.Currencies {
		currencyCreateRequest := daccount.CurrencyCreateRequest{
			Currency: requestCurency.Currency,
		}
		if requestCurency.CreditLimit != nil {
			currencyCreateRequest.CreditLimit = &requestCurency.CreditLimit.Value
		}
		domainCurrencies = append(domainCurrencies, currencyCreateRequest)
	}

	domainAcountDetails, err := s.accounts.CreateAccount(ctx, daccount.AccountCreateRequest{
		SubsystemId: *subsystemId,
		AllowAsync:  createAccountRequest.AllowAsync,
		Currencies:  domainCurrencies,
	})

	if err != nil {
		return nil, mapToServiceError(err)
	}

	return &account.CreateAccountResponse{
		Status: 0,
		Details: &account.AccountDetailsResponse{
			Id:          domainAcountDetails.Id[:],
			IsSuspended: domainAcountDetails.IsSuspended,
			AllowAsync:  domainAcountDetails.AllowAsync,
		},
	}, nil
}
func (s *accountServer) GetAccount(ctx context.Context, getAccountRequest *account.GetAccountRequest) (*account.AccountDetailsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccount not implemented")
}

func mapToServiceError(err error) error {
	return err
}

func NewAccountsServer(log *zap.Logger, accounts daccount.Accounts, dd datadog.Datadog) account.AccountServer {
	return &accountServer{log: log.Named("account-server"), accounts: accounts, dd: dd}
}
