package activities

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/executors"
	"github.com/surajsub/temporal-rest-dsl/models"
)

/*

This file will invoke the function to build the resources
*/

func deployResource(step models.Step, logger *logrus.Logger) (map[string]any, error) {

	logger.Printf("[*********** In the DeployResource Function ***************]")
	logger.Infof("Deploying resource for Cloud Provider: %s, Resource: %s for customer %s", step.Provider, step.Resource, step.Customer)

	// Initialize the executor
	executor, err := initializeExecutor(step, logger)
	if err != nil {
		return nil, err
	}

	// Validate the operation
	err = executor.ValidateOperation(step)
	if err != nil {
		logger.Errorf("invalid operation for executor: %v", err)
		return nil, err
	}

	// Execute the operation
	output, err := executor.Execute(step, step.Executor, step.Variables)
	if err != nil {
		logger.Errorf("Execution failed for %s/%s: %v", step.Action, step.Resource, err)
		return nil, fmt.Errorf("error executing %s for %s: %v", step.Action, step.Resource, err)
	}
	logger.Infof("Resource %s for customer %s deployed successfully using Executor %s", step.Resource, step.Customer, step.Executor)
	logger.Infof("Output: %v for Executor %s", output, step.Executor)
	return output, nil
}

func initializeExecutor(step models.Step, logger *logrus.Logger) (executors.Executor, error) {
	// Prepare the configuration map from the step
	logger.Infof(" Initializing executor for %s with the activity %s", step.Executor, step.Activity)
	config := map[string]any{
		"customer":       step.Customer,
		"workspace":      step.Workspace,
		"provider":       step.Provider,
		"resource":       step.Resource,
		"variables":      step.Variables,
		"provisioner":    step.Provisioner,
		"project":        step.Project,
		"submitter":      step.Submitter,
		"resource_group": step.ResourceGroup,
		"file":           step.File,
		"action":         step.Action,
		"operation":      step.Operation,
	}

	logger.Infof("in the intializeExecutor code with  %s and then %v", step.Executor, config)
	// Fetch and initialize the executor using the registry

	executor, err := executors.GetExecutor(step.Executor, config)
	if err != nil {
		logger.Errorf("failed to initialize executor '%s': %v", step.Executor, err)
		return nil, err
	}
	logger.Infof("Executor initialized successfully for %s", step.Executor)
	return executor, nil
}

func GetDSLActivityLogger(ctx context.Context) *logrus.Logger {
	if loggerInterface := ctx.Value("logger"); loggerInterface != nil {
		return loggerInterface.(*WorkflowLogger).logger
	}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)
	logger.Warn("Using fallback logger as no logger was passed")
	return logger
}
