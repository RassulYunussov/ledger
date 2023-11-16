package updater

import (
	"balanceupdater/workflows"
	"context"
	"ledgercore/config"
	"shared/service/database"
	"time"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

type temporalBalanceUpdater struct {
	log                    *zap.Logger
	client                 client.Client
	balanceUpdateQueueName string
}

func (t *temporalBalanceUpdater) Stop() {
	t.log.Info("close temporal client")
	t.client.Close()
}

func (t *temporalBalanceUpdater) UpdateBalance(operation database.OperationStatusTransition) {

	options := client.StartWorkflowOptions{
		TaskQueue: t.balanceUpdateQueueName,
	}
	secondContext, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	we, err := t.client.ExecuteWorkflow(secondContext, options, workflows.UpdateBalance, operation)
	if err != nil {
		t.log.Error("Unable to start the Workflow", zap.Error(err))
	}
	t.log.Info("initiated update balance workflow", zap.String("worklow id", we.GetID()))

}

func createTemporalBalanceUpdater(log *zap.Logger, configuration config.Configuration) (Updater, error) {

	logger := logur.LoggerToKV(zapadapter.New(log))
	clientOptions := client.Options{
		Namespace: configuration.Temporal.Namespace,
		Logger:    logger,
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		return nil, err
	}
	temporal := temporalBalanceUpdater{log.Named("temporal"), c, configuration.Temporal.TaskQueue}
	return &temporal, nil
}
