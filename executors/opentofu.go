package executors

import (
	"fmt"
	"github.com/surajsub/temporal-rest-dsl/models"
	"log"
	"os/exec"
)

type OpenTFExecutor struct {
	*ExecutorBase
}

// Constructor for OpenTFExecutor
func NewOpenTFExecutor(customer, workspace, provider, resource, provisioner, action, operation string) *OpenTFExecutor {
	return &OpenTFExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
	}
}

func (o *OpenTFExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {
	// Initialize the Opentofu executor

	// Ensure Logger is not nil before usage
	if o.Logger == nil {
		return nil, fmt.Errorf("logger is not initialized")
	}
	o.Logger.Infof("Executing OpenTFExecutor with action %s and operation %s", o.Action, step.Operation)

	switch o.Action {
	case "create":
		o.Logger.Debugf("Starting 'create' operation for resource: %s", o.Resource)

		// Initialize, plan, and apply Opentofu
		err := o.Init()
		if err != nil {
			o.Logger.Errorf("error during init: %v", err)
			return nil, fmt.Errorf("error during init: %v", err)
		}

		err = o.Plan()
		if err != nil {
			o.Logger.Errorf("error during plan: %v", err)
			return nil, fmt.Errorf("error during plan: %v", err)
		}

		output, err := o.Apply()
		if err != nil {
			o.Logger.Errorf("error during Apply: %v", err)
			return nil, fmt.Errorf("error during apply: %v", err)
		}
		return output, nil
	case "delete":
		o.Logger.Infof("Starting 'delete' operation for resource: %s", o.Resource)

		// Added the plan for the destroy operation.. else it would encounter a failure
		err := o.Plan()
		if err != nil {
			o.Logger.Errorf("error during Plan for Delete: %v", err)
			return nil, fmt.Errorf("error during plan: %v", err)
		}
		err = o.Destroy()
		if err != nil {
			o.Logger.Errorf("error during Destroy for Delete: %v", err)
			return nil, fmt.Errorf("error during destroy: %v", err)
		}
		return map[string]any{"status": "destroyed"}, nil
	default:
		o.Logger.Errorf("unsupported operation %s for OpenTofuExecutor", step.Operation)
		return nil, fmt.Errorf("unsupported operation %s for OpenTofuExecutor", step.Operation)
	}
}

func (o *OpenTFExecutor) ValidateOperation(step models.Step) error {
	if step.Operation != "deploy" && step.Operation != "destroy" {
		log.Printf("invalid operation %s for OpenTofuExecutor", step.Operation)
	}
	return nil
}

func (o *OpenTFExecutor) apply(payload map[string]any) (map[string]any, error) {
	log.Println("Running OpenTofu apply with variables:", o.Variables)

	return map[string]any{"status": "success"}, nil
}

func (o *OpenTFExecutor) destroy(payload map[string]any) (map[string]any, error) {
	log.Println("Running OpenTofu destroy with variables:", o.Variables)
	// Implement OpenTofu logic here
	return map[string]any{"status": "destroyed"}, nil
}

// Init initializes OpenTofu in the specified workspace
func (o *OpenTFExecutor) Init() error {
	log.Printf("Initializing OpenTofu in workspace: %s", o.Workspace)
	cmd := exec.Command("tofu", "init")
	cmd.Dir = o.Workspace
	return RunCommand(cmd, o.Logger)
}

// Plan runs the OpenTofu plan command
func (o *OpenTFExecutor) Plan() error {
	varArgs := FormatVariables(o.Variables)
	o.Logger.Infof("Running 'Opentofu plan' for resource : %s", o.Resource)
	args := append([]string{"plan", "-input=false"}, varArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = o.Workspace
	return RunCommand(cmd, o.Logger)
}

// Apply applies the OpenTofu configuration and captures outputs
func (o *OpenTFExecutor) Apply() (map[string]any, error) {
	varArgs := FormatVariables(o.Variables)
	args := append([]string{"apply", "-input=false", "-auto-approve"}, varArgs...)
	cmd := exec.Command("tofu", args...)
	o.Logger.Debugf("Executing command: %v in workspace: %s", cmd.Args, o.Workspace)

	cmd.Dir = o.Workspace

	err := RunCommand(cmd, o.Logger)
	if err != nil {
		return nil, err
	}

	outputs, err := CaptureOpenTofuOutputs(o.Workspace, o.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to capture outputs: %w", err)
	}

	return outputs, nil
}

// Destroy destroys the OpenTofu-managed resources
func (o *OpenTFExecutor) Destroy() error {
	varArgs := FormatVariables(o.Variables)
	o.Logger.Infof("Running 'opentofu destroy' for resource: %s", o.Resource)
	args := append([]string{"destroy", "-input=false", "-auto-approve"}, varArgs...)
	cmd := exec.Command("tofu", args...)
	cmd.Dir = o.Workspace

	err := RunCommand(cmd, o.Logger)
	if err != nil {
		return err
	}

	return nil
}

// Utility function to format OpenTofu variable arguments
