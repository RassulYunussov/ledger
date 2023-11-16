package processor

import (
	"balanceupdater/activities"
	"balanceupdater/config"
	"balanceupdater/workflows"
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
	"go.uber.org/zap"
	zapadapter "logur.dev/adapter/zap"
	"logur.dev/logur"
)

type ProcessorParameters struct {
	fx.In
	Log           *zap.Logger
	Lifecycle     fx.Lifecycle
	Configuration config.Configuration
}

type Processor struct {
	log     *zap.Logger
	client  client.Client
	workers []worker.Worker
}

func NewProcessor(p ProcessorParameters) *Processor {

	processor := &Processor{log: p.Log.Named("processor")}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			logger := logur.LoggerToKV(zapadapter.New(processor.log))
			clientOptions := client.Options{
				Namespace: p.Configuration.Temporal.Namespace,
				Logger:    logger,
			}
			temporalClient, err := client.Dial(clientOptions)
			if err != nil {
				processor.log.Error("Unable to create a Temporal Client", zap.Error(err))
				return err
			}
			processor.client = temporalClient
			processor.workers = make([]worker.Worker, p.Configuration.Temporal.Workers)
			for i := 0; i < p.Configuration.Temporal.Workers; i++ {
				// Create a new Worker
				worker := worker.New(temporalClient, p.Configuration.Temporal.TaskQueue, worker.Options{})
				// Register Workflows
				worker.RegisterWorkflow(workflows.UpdateBalance)
				// Register Acivities
				worker.RegisterActivity(activities.UpdateBalance)
				// Start the the Worker Process
				err = worker.Start()
				if err != nil {
					processor.log.Error("Unable to start a Temporal Worker", zap.Error(err))
					return err
				}
				processor.workers[i] = worker
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			for _, worker := range processor.workers {
				worker.Stop()
			}
			processor.client.Close()
			return nil
		},
	})

	return processor
}
