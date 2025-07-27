package activities

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/models"
	"go.temporal.io/sdk/activity"
	_ "go.temporal.io/sdk/workflow"
)

type WorkflowLogger struct {
	logger *logrus.Logger
}

// RunActivity /*
func RunActivity(ctx context.Context, step models.Step) (map[string]any, error) {
	logger := GetDSLActivityLogger(ctx)

	logger.Infof("[********* Running activity: %s *********]", step.Activity)

	logger.Infof("Executing  %s for %s in workspace %s with variables: %v", step.Operation, step.Resource, step.Workspace, step.Variables)

	// Call the deployResource function and get the output

	activityInfo := activity.GetInfo(ctx)
	logger.Infof("Running activity: %s (Step: %s, WorkflowID: %s)", step.Activity, step.ID, activityInfo.WorkflowExecution.ID)

	// ✅ Send a heartbeat so that Temporal UI updates the progress
	activity.RecordHeartbeat(ctx, fmt.Sprintf("Executing step: %s (Activity: %s)", step.ID, step.Activity))

	// This is the top level action in the yaml. We support only two actions
	if step.Action == "create" || step.Action == "delete" {
		logger.Infof("Running: %s for: %s ", step.Activity, step.Resource)
		output, err := deployResource(step, logger)
		if err != nil {
			logger.Errorf("Error in deployResource: %v", err)

			return nil, err
		}
		return output, nil
	} else {
		logger.Infof("NO ACTION PROVIDED IN THE YAML FILE- HENCE IGNORING Running Update")
		return nil, nil
	}

}
