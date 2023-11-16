# generation command
## generate transfer service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative shared/protocol/transfer/transfer.proto 

## generate account service
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative shared/protocol/account/account.proto 

# test 
## create transaction
grpcurl -plaintext -d '{"id": "SkMd/s+6SoSUma1AM78/tg==", "source": "SkMd/s+6SoSUma1AM78/tg==", "currency":"KZT", "amount":100, "target": "ZxEguj/fQuSqurP4xvObVg==", "description":"gift"}' -H 'subsystem: 9667453c-e685-4c2a-a1ba-98f2efda9b0a' -proto "shared/protocol/transfer/transfer.proto" localhost:50051 transfer.Transfer/CreateTransaction

## create account
grpcurl -plaintext -d '{"allowAsync": true, "currencies": [{"currency":"KZT"}]}' -H 'subsystem: 9667453c-e685-4c2a-a1ba-98f2efda9b0a' -proto "shared/protocol/account/account.proto" localhost:50051 account.Account/CreateAccount