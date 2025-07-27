package workflows

import (
	"fmt"
	"regexp"
	"time"

	"github.com/surajsub/temporal-rest-dsl/activities"
	"github.com/surajsub/temporal-rest-dsl/models"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type ActivityInput struct {
	TaskDescription string         `json:"taskDescription"`
	Parameters      map[string]any `json:"parameters"`
}

// RetrySignal is the signal structure
type RetrySignal struct {
	StepID string                 `json:"step_id"`
	Inputs map[string]interface{} `json:"inputs"`
	Action string                 `json:"action"`
}

type WorkflowState struct {
	Results map[string]map[string]any
}

// Regex to match ${dependency.output} placeholders
var variableRegex = regexp.MustCompile(`\${([a-zA-Z0-9_]+)\.([a-zA-Z0-9_]+)}`)

func reverseSteps(steps []models.Step) []models.Step {
	reversed := make([]models.Step, len(steps))
	for i, step := range steps {
		reversed[len(steps)-1-i] = step
	}
	return reversed
}

func TemporalExecutorWorkflow(ctx workflow.Context, input WorkflowInput) (map[string]map[string]interface{}, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NewTemporalExecutorWorkflow")
	signalChan := workflow.GetSignalChannel(ctx, "step_control_signal")

	// Store the workflow state
	state := WorkflowState{
		Results: make(map[string]map[string]interface{}),
	}

	stateFile := fmt.Sprintf("%s-%s-%s", input.Project, input.DeploymentId, input.Account)
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    5 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    5,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	// Handle delete order
	if input.Action == "delete" {
		logger.Info("Loading state for delete")
		if err := workflow.ExecuteActivity(ctx, activities.LoadStateFromStorage, stateFile).Get(ctx, &state.Results); err != nil {
			return nil, fmt.Errorf("failed to load state for delete: %w", err)
		}
		input.Steps = reverseSteps(input.Steps)
	}

	// placeholder for the steps
	//
	pendingSteps := input.Steps
	completedSteps := make(map[string]bool)

	for len(pendingSteps) > 0 {
		nextPending := []models.Step{}

		for _, step := range pendingSteps {
			if !dependenciesMet(step, completedSteps) {
				nextPending = append(nextPending, step)
				continue
			}
			logger.Info("[** Executing Step **]", "[** step ID **]", step.ID)
			step.RoleID = input.RoleID
			step.SecretId = input.SecretId
			if input.Action == "create" || input.Action == "delete" {
				step.Variables = ProcessStepVariables(step, state.Results[step.ID])
			}
			step = PrepareStep(step, input, state.Results)
			stepCtx := workflow.WithValue(ctx, "step", step.ID)

			if err := workflow.ExecuteActivity(stepCtx, activities.DBActivity, step, input.SubmissionID, step.Activity, "STARTED", map[string]any{}).Get(stepCtx, nil); err != nil {
				return nil, fmt.Errorf("db STARTED step %s: %w", step.ID, err)
			}
			var result map[string]any
			execErr := workflow.ExecuteActivity(stepCtx, activities.RunActivity, step).Get(stepCtx, &result)
			if execErr != nil{
				logger.Info("********** Deploy Resource Step failed..Updating the db with status ***********")
				if err := workflow.ExecuteActivity(stepCtx, activities.DBActivity, step, input.SubmissionID, step.Activity, "FAILED", result).Get(stepCtx, nil); err != nil {
					return nil, fmt.Errorf("Failed to update the DB with the right status step %s: %w", step.ID, err)
				}
				
			}

			if execErr != nil {
				logger.Error("[******* Deploy Resource Step failed: %s, %v ", step.Action, execErr)
				for {
					var signal RetrySignal
					signalChan.Receive(ctx, &signal)
					logger.Info("[**************Received signal ***********]]", "signal", signal)

					if signal.StepID != step.ID {
						logger.Debug("The signal and step are NOT THE SAME.. WHY")
						continue
					}

					switch signal.Action {
					case "ignore":
						logger.Info("Step %s ignored via signal", step.ID)
						result = map[string]any{"message": "Step ignored manually"}
					case "retry":
						retryStep := step
						if signal.Inputs != nil {
							retryStep.Variables = deepCopy(signal.Inputs)
						}
						var retryResult map[string]any
						retryErr := workflow.ExecuteActivity(stepCtx, activities.RunActivity, retryStep).Get(stepCtx, &retryResult)
						if retryErr != nil {
							logger.Error("Retry failed again for step %s: %v", step.ID, retryErr)
							continue
						}
						result = retryResult
					default:
						logger.Warn("Unknown signal action received", "action", signal.Action)
						continue
					}
					break
				}
			}

			state.Results[step.ID] = result
			completedSteps[step.ID] = true

			if err := workflow.ExecuteActivity(stepCtx, activities.DBActivity, step, input.SubmissionID, step.Activity, "SUCCESS", result).Get(stepCtx, nil); err != nil {
				return nil, fmt.Errorf("db SUCCESS step %s: %w", step.ID, err)
			}

			logger.Info("Completed step", "stepID", step.ID)
		}
		pendingSteps = nextPending
	}

	if input.Action == "create" {
		logger.Info("Saving workflow state")
		if err := workflow.ExecuteActivity(ctx, activities.SaveStateToStorage, state.Results, stateFile).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to save state: %w", err)
		}
	}

	logger.Info("Workflow complete")
	return state.Results, nil

}

func deepCopy(input map[string]interface{}) map[string]interface{} {
	copy := make(map[string]interface{}, len(input))
	for k, v := range input {
		copy[k] = v
	}
	return copy
}

func dependenciesMet(step models.Step, completed map[string]bool) bool {
	for _, dep := range step.DependsOn {
		if !completed[dep] {
			return false
		}
	}
	return true
}

func prepareStepWithContext(step models.Step, input WorkflowInput, results map[string]map[string]interface{}) models.Step {
	step.RoleID = input.RoleID
	step.SecretId = input.SecretId

	if input.Action == "create" || input.Action == "delete" {
		step.Variables = ProcessStepVariables(step, results[step.ID])
	}

	return PrepareStep(step, input, results)
}

func handleStepFailureWithSignal(ctx workflow.Context, step models.Step, signalChan workflow.ReceiveChannel, logger log.Logger) (map[string]any, error) {
	for {
		var signal RetrySignal
		signalChan.Receive(ctx, &signal)
		logger.Info("Received signal", "signal", signal)

		if signal.StepID != step.ID {
			logger.Debug("Signal does not match step, skipping")
			continue
		}

		switch signal.Action {
		case "ignore":
			logger.Info("Step %s ignored via signal", step.ID)
			return map[string]any{"message": "Step ignored manually"}, nil
		case "retry":
			retryStep := step
			if signal.Inputs != nil {
				retryStep.Variables = deepCopy(signal.Inputs)
			}
			var retryResult map[string]any
			err := workflow.ExecuteActivity(ctx, activities.RunActivity, retryStep).Get(ctx, &retryResult)
			if err != nil {
				logger.Error("Retry failed again for step %s: %v", step.ID, err)
				continue
			}
			return retryResult, nil
		default:
			logger.Warn("Unknown signal action received", "action", signal.Action)
		}
	}
}

// New Function
//

func LTemporalExecutorWorkflow(ctx workflow.Context, input WorkflowInput) (map[string]map[string]interface{}, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting TemporalExecutorWorkflow")

	signalChan := workflow.GetSignalChannel(ctx, "step_control_signal")
	//	var signal RetrySignal

	state := WorkflowState{
		Results: make(map[string]map[string]interface{}),
	}

	stateFile := fmt.Sprintf("%s-%s-%s", input.Project, input.DeploymentId, input.Account)
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    5 * time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    5,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy:         retryPolicy,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Handle delete order
	if input.Action == "delete" {
		logger.Info("Loading state for delete")
		if err := workflow.ExecuteActivity(ctx, activities.LoadStateFromStorage, stateFile).Get(ctx, &state.Results); err != nil {
			return nil, fmt.Errorf("failed to load state for delete: %w", err)
		}
		input.Steps = reverseSteps(input.Steps)
	}

	for _, step := range input.Steps {

		logger.Info("Executing step", "stepID", step.ID)
		// Inject credentials
		step.RoleID = input.RoleID
		step.SecretId = input.SecretId

		if input.Action == "create" || input.Action == "delete" {
			step.Variables = ProcessStepVariables(step, state.Results[step.ID])
		}

		if err := WaitForDependencies(ctx, step, state.Results); err != nil {
			return nil, err
		}

		step = PrepareStep(step, input, state.Results)
		stepCtx := workflow.WithValue(ctx, "step", step.ID)

		// Track STARTED
		//future := workflow.ExecuteActivity(ctx workflow.Context, activity interface{}, args ...interface{})
		if err := workflow.ExecuteActivity(stepCtx, activities.DBActivity, step, input.SubmissionID, step.Activity, "STARTED", map[string]any{}).Get(stepCtx, nil); err != nil {
			return nil, fmt.Errorf("db STARTED step %s: %w", step.ID, err)
		}

		var result map[string]any

		execErr := workflow.ExecuteActivity(stepCtx, activities.RunActivity, step).Get(stepCtx, &result)

		if execErr != nil {

			logger.Error("[******* Deploy Resource Step failed: %s, %v ", step.Action, execErr)
			for {
				var signal RetrySignal
				signalChan.Receive(ctx, &signal)
				logger.Info("[**************Received signal ***********]]", "signal", signal)

				logger.Debug("Received signal for StepID: %s. Current failing StepID: %s", signal.StepID, step.ID)

				logger.Debug("The step that has failed is %s", step.ID)
				logger.Debug("Step is passed in as the signa is %s", signal.StepID)
				if signal.StepID != step.ID {
					logger.Debug("The signal and step are NOT THE SAME.. WHY")
					continue
				}
				logger.Info("Step %s finished", step.Action)

				switch signal.Action {
				case "ignore":
					logger.Info("Step %s ignored via signal", step.ID)
					result = map[string]any{"message": "Step ignored manually"}
				case "retry":

					logger.Debug("Received signal", "stepID", signal.StepID, "action", signal.Action)
					retryStep := step
					logger.Debug("[*******************Signal Inputs are %v ***************]", signal.Inputs)
					logger.Debug("Step is passed in as the signa is %v", step.Variables)
					logger.Debug("******** PRINTING THE RETRY STEP ********** %v", retryStep)
					// Updated code START HERE
					//retryStep = step
					if signal.Inputs != nil {

						retryStep.Variables = deepCopy(signal.Inputs)
					}
					logger.Debug("[******* RETRYING STEP %s with variables %s", retryStep.ID, retryStep.Variables)
					var retryResult map[string]any
					retryErr := workflow.ExecuteActivity(stepCtx, activities.RunActivity, retryStep).Get(stepCtx, &retryResult)
					if retryErr != nil {
						logger.Error("Retry failed again for step %s: %v", step.ID, retryErr)
						continue // Wait for next signal
					}
					result = retryResult
				default:
					logger.Warn("Unknown signal action received", "action", signal.Action)
					continue
				}

				break
			}
		}

		state.Results[step.ID] = result
		// Mark activity as SUCCESS in DB
		if err := workflow.ExecuteActivity(stepCtx, activities.DBActivity, step, input.SubmissionID, step.Activity, "SUCCESS", result).Get(stepCtx, nil); err != nil {
			return nil, fmt.Errorf("db SUCCESS step %s: %w", step.ID, err)
		}

		logger.Info("Completed step", "stepID", step.ID)
	}

	if input.Action == "create" {
		logger.Info("Saving workflow state")
		if err := workflow.ExecuteActivity(ctx, activities.SaveStateToStorage, state.Results, stateFile).Get(ctx, nil); err != nil {
			return nil, fmt.Errorf("failed to save state: %w", err)
		}
	}

	logger.Info("Workflow complete")
	return state.Results, nil
}
