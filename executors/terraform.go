package executors

import (
	"bytes"
	"fmt"
	"github.com/surajsub/temporal-rest-dsl/models"

	"os"
	"os/exec"
	"path/filepath"
)

type TerraformExecutor struct {
	*ExecutorBase
}

// Constructor for TerraformExecutor
func NewTerraformExecutor(customer, workspace, provider, resource, provisioner, action, operation string) *TerraformExecutor {
	return &TerraformExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
	}
}

func (t *TerraformExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {
	// Initialize the Terraform executor

	// Ensure Logger is not nil before usage
	if t.Logger == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	t.Logger.Infof("Executing TerraformExecutor with action %s and operation %s", t.Action, step.Operation)

	switch t.Action {
	case "create":
		t.Logger.Debugf("Starting 'create' operation for resource: %s", t.Resource)

		// Initialize, plan, and apply Terraform
		err := t.Init()
		if err != nil {
			t.Logger.Errorf("error during init: %v", err)
			return nil, fmt.Errorf("error during init: %w", err)
		}

		err = t.Plan()
		if err != nil {
			t.Logger.Errorf("error during plan: %w", err)
			return nil, fmt.Errorf("error during plan: %w", err)
		}

		output, err := t.Apply()
		if err != nil {
			t.Logger.Errorf("error during Apply: %w", err)
			return nil, fmt.Errorf("error during apply: %w", err)
		}
		return output, nil
	case "delete":
		t.Logger.Infof("Starting 'delete' operation for resource: %s", t.Resource)

		// Added the plan for the destroy operation.. else it would encounter a failure
		err := t.Plan()
		if err != nil {
			t.Logger.Errorf("error during Plan for Delete: %v", err)
			return nil, fmt.Errorf("error during plan: %v", err)
		}
		err = t.Destroy()
		if err != nil {
			t.Logger.Errorf("error during Destroy for Delete: %v", err)
			return nil, fmt.Errorf("error during destroy: %v", err)
		}
		return map[string]any{"status": "destroyed"}, nil
	default:
		t.Logger.Errorf("unsupported operation %s for TerraformExecutor", step.Operation)
		return nil, fmt.Errorf("unsupported operation %s for TerraformExecutor", step.Operation)
	}
}

func (t *TerraformExecutor) ValidateOperation(step models.Step) error {

	// The terraform executor has no need to support any operations since they are supported by the actions
	// we should do this for the opentofu and bicep executors as well
	return nil
}

// Init initializes Terraform in the specified workspace
func (t *TerraformExecutor) Init() error {
	t.Logger.Infof("Initializing Terraform in workspace: %s", t.Workspace)
	cmd := exec.Command("terraform", "init")
	cmd.Dir = t.Workspace
	return RunCommand(cmd, t.Logger)

}

// Plan runs the Terraform plan command
func (t *TerraformExecutor) Plan() error {
	varArgs := FormatVariables(t.Variables)
	t.Logger.Infof("Running 'terraform plan' for resource : %s", t.Resource)
	args := append([]string{"plan", "-input=false"}, varArgs...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = t.Workspace
	
	return RunCommand(cmd, t.Logger)
}

// Run the plan and out runs the Terraform plan command
func (t *TerraformExecutor) PlanOut() error {
	varArgs := FormatVariables(t.Variables)
	t.Logger.Infof("Running 'terraform plan and out' for resource: %s", t.Resource)
	args := append([]string{"plan", "-input=false", "-out=plan.binary"}, varArgs...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = t.Workspace
	return RunCommand(cmd, t.Logger)
}

// Run the plan to conver to json
func (t *TerraformExecutor) Show() error {
	t.Logger.Infof("Running 'terraform show and out' for resource: %s", t.Workspace)

	// Define the command
	cmd := exec.Command("terraform", "show", "-json", "plan.binary")

	// Set the working directory
	cmd.Dir = t.Workspace

	// Create or open the file for writing
	outputFilePath := filepath.Join(t.Workspace, "plan.json")
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		t.Logger.Errorf("failed to create plan.json file: %v", err)
		return fmt.Errorf("failed to create plan.json file: %w", err)
	}
	defer outputFile.Close()

	// Redirect the command's stdout to the file
	cmd.Stdout = outputFile

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		t.Logger.Errorf("failed to run terraform show: %s:  %v", stderr.String(), err)
		return fmt.Errorf("failed to run terraform show: %s:  %w", stderr.String(), err)
	}

	t.Logger.Debugf("Terraform plan successfully exported to: %s", outputFilePath)
	return nil
}

// Apply applies the Terraform configuration and captures outputs
func (t *TerraformExecutor) Apply() (map[string]any, error) {
	varArgs := FormatVariables(t.Variables)
	args := append([]string{"apply", "-input=false", "-auto-approve"}, varArgs...)
	cmd := exec.Command("terraform", args...)
	t.Logger.Debugf("Executing command: %v in workspace: %s", cmd.Args, t.Workspace)

	cmd.Dir = t.Workspace

	err := RunCommand(cmd, t.Logger)
	if err != nil {
		return nil, err
	}

	outputs, err := CaptureTerraformOutputs(t.Workspace, t.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to capture outputs: %w", err)
	}

	return outputs, nil
}

// Destroy destroys the Terraform-managed resources
func (t *TerraformExecutor) Destroy() error {
	varArgs := FormatVariables(t.Variables)
	t.Logger.Infof("Running 'terraform destroy' for resource: %s", t.Resource)
	args := append([]string{"destroy", "-input=false", "-auto-approve"}, varArgs...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = t.Workspace

	err := RunCommand(cmd, t.Logger)
	if err != nil {
		return err
	}

	return nil
}
