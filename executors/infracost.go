package executors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/surajsub/temporal-rest-dsl/models"
)

type InfraCostExecutor struct {
	*ExecutorBase
	Submitter string
	Project   string
}

// Constructor for TerraformExecutor
func NewInfraCostExecutor(customer, workspace, provider, resource, provisioner, action, operation, submitter, project string) *InfraCostExecutor {
	return &InfraCostExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
		Submitter:    submitter,
		Project:      project,
	}
}

func (ice *InfraCostExecutor) EstimateCost(operation, workspace string) (map[string]any, any) {
	ice.Logger.Infof("Estimating cost with %s in workspace: %s with %s", operation, workspace, ice.Provisioner)
	//log.Printf("Estimating cost with %s in workspace: %s with %s", operation, workspace, ice.Provisioner)

	// Show the plan in JSON format
	cmd := exec.Command(operation, "breakdown", "--path", "plan.json", "--fields", "all", "--format", "json", "--out-file", "output.json")
	cmd.Dir = workspace

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run infracost breakdown: %v. Command: %s, Workspace: %s", err, operation, workspace)
	}
	log.Printf("Infracost breakdown completed successfully. Output file: %s/output.json", workspace)

	// Read the generated file
	outputFile := fmt.Sprintf("%s/output.json", workspace)
	content, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read infracost output file: %v", err)
	}

	// Parse JSON content

	var parsedJSON map[string]any
	err = json.Unmarshal(content, &parsedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON content: %v", err)
	}

	fmt.Printf("the parsed json is %s\n", parsedJSON["totalMonthlyCost"])

	fmt.Sprintf("the total cost of the resource is %s %s\n", parsedJSON["totalMonthlyCost"], parsedJSON["currency"])
	var totalCost = fmt.Sprintf("%s %s", parsedJSON["totalMonthlyCost"], parsedJSON["currency"])

	// Return the extracted data
	return map[string]any{
		"estimated_cost": totalCost,
	}, nil
}

func (ice *InfraCostExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {

	log.Printf("Executing InfraCostExecutor with action %s", ice.Action)
	//var costEstimate map[string]interface{}
	switch step.Operation {
	case "cost_estimate":

		ice.Logger.Infof("the executor is: %s \n", executor)
		ice.Logger.Infof("the payload is: %v \n", payload)

		//var provisioningExecutor = payload["provisioner"].(string)
		ice.Logger.Infof("Executing Cost Estimate for resource %s using the executor %s", ice.Resource, executor)
		ice.Logger.Infof("Starting 'deploy' operation for resource: %s", ice.Resource)
		err := ice.Init(executor)
		if err != nil {
			return nil, fmt.Errorf("error during init: %v", err)
		}

		err = ice.PlanOut(executor)
		if err != nil {
			return nil, fmt.Errorf("error during planout: %v", err)
		}

		err = ice.Show(executor)
		if err != nil {
			return nil, fmt.Errorf("error during Terraform show: %v", err)
		}

		costEstimate, _ := ice.EstimateCost(executor, ice.Workspace)
		ice.Logger.Infof("Cost Estimate for the resource: %v", costEstimate)
		return costEstimate, nil

	default:
		return nil, fmt.Errorf("unsupported operation %s for InfracostExecutor", step.Operation)
	}

}

func (ice *InfraCostExecutor) ValidateOperation(step models.Step) error {
	if step.Operation != "cost_estimate" {
		return fmt.Errorf("invalid operation %s for InfraCostExecutor", step.Operation)
	}
	return nil
}

// Init initializes in the specified workspace
func (ice *InfraCostExecutor) Init(executor string) error {
	//log.Printf("Initializing  init .. %s in workspace: %s with provisioner %s", executor, ice.Workspace, ice.Provisioner)
	ice.Logger.Infof("Initializing %s in workspace: %s", executor, ice.Workspace)
	args := append([]string{"init"})
	cmd := exec.Command(ice.Provisioner, args...)
	cmd.Dir = ice.Workspace
	return RunCommand(cmd, ice.Logger)
}

// PlanOut Run the plan and out runs the Terraform plan command
func (ice *InfraCostExecutor) PlanOut(executor string) error {
	varArgs := FormatVariables(ice.Variables)
	ice.Logger.Infof("Running '%s plan and out' for resource: %s", ice.Provisioner, ice.Resource)
	args := append([]string{"plan", "-input=false", "-out=plan.binary"}, varArgs...)
	cmd := exec.Command(ice.Provisioner, args...)
	cmd.Dir = ice.Workspace
	return RunCommand(cmd, ice.Logger)
}

// Show Run the plan to convert to json
func (ice *InfraCostExecutor) Show(executor string) error {
	ice.Logger.Infof("Initializing %s in workspace: %s", executor, ice.Workspace)

	// Define the command
	cmd := exec.Command(ice.Provisioner, "show", "-json", "plan.binary")

	// Set the working directory
	cmd.Dir = ice.Workspace

	// Create or open the file for writing
	outputFilePath := filepath.Join(ice.Workspace, "plan.json")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create plan.json file: %w", err)
	}
	defer outputFile.Close()

	// Redirect the command's stdout to the file
	cmd.Stdout = outputFile

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run terraform show: %s:  %w", stderr.String(), err)
	}

	ice.Logger.Debugf("%s plan successfully exported to: %s", executor, outputFilePath)
	return nil
}
