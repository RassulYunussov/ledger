package workflows

import (
	"balanceupdater/activities"
	"shared/service/database"
	"time"

	"go.temporal.io/sdk/workflow"
)

func UpdateBalance(ctx workflow.Context, operation database.OperationStatusTransition) error {
	// Define the Activity Execution options
	// StartToCloseTimeout or ScheduleToCloseTimeout must be set
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	// Execute the Activity synchronously (wait for the result before proceeding)
	err := workflow.ExecuteActivity(ctx, activities.UpdateBalance, operation).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
