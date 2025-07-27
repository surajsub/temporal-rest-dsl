package executors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/surajsub/temporal-rest-dsl/models"
	"log"
	"os/exec"
)

type BicepExecutor struct {
	*ExecutorBase
	File           string
	ResourceGroup  string
	DeploymentName string `json:"deploymentName"`
}

// Constructor for BicepExecutor
func NewBicepExecutor(customer, workspace, provider, resource, provisioner, action, operation, file, resourcegroup, deploymentname string) *BicepExecutor {
	return &BicepExecutor{
		ExecutorBase:   NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
		File:           file,
		ResourceGroup:  resourcegroup,
		DeploymentName: deploymentname,
	}
}

func (b *BicepExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {
	b.Logger.Infof("Executing BicepExecutor with action %s and operation %s", b.Action, step.Operation)

	switch b.Action {
	case "create":
		output, nil := b.ExecuteCreateOperation(step)
		return output, nil
	case "delete":
		log.Printf("Starting 'destroy' operation for resource: %s", b.Resource)
		err := b.Destroy()
		if err != nil {
			return nil, fmt.Errorf("error during destroy: %v", err)
		}
		return map[string]any{"status": "destroyed"}, nil
	default:
		return nil, fmt.Errorf("unsupported operation [ %s ] for Bicep", step.Operation)
	}
}

func (t *BicepExecutor) ValidateOperation(step models.Step) error {

	return nil
}

// Apply applies the Bicep configuration and captures outputs
func (t *BicepExecutor) Apply() (map[string]any, error) {
	varArgs := FormatBicepVariables(t.Variables)

	args := []string{
		"deployment", "group", "create",
		"--resource-group", t.ResourceGroup,
		"--template-file", t.File,
		"--name", t.DeploymentName,
	}

	args = append(args, varArgs...)
	cmd := exec.Command("az", args...)
	log.Printf("Executing command: %v in workspace: %s", cmd.Args, t.Workspace)
	cmd.Dir = t.Workspace

	// Capture the command output
	var outBuffer, errBuffer bytes.Buffer
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		log.Printf("Command failed: %v\nStderr: %s", err, errBuffer.String())
		return nil, fmt.Errorf("command failed: %w", err)
	}

	log.Printf("Command succeeded. Output: %s", outBuffer.String())

	// Parse the JSON output from the command
	var outputs map[string]any
	err = json.Unmarshal(outBuffer.Bytes(), &outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return outputs, nil
}

// Destroy destroys the Terraform-managed resources
func (t *BicepExecutor) Destroy() error {
	varArgs := FormatBicepVariables(t.Variables)
	log.Printf("Running 'terraform destroy' for resource: %s", t.Resource)
	args := append([]string{"destroy", "-input=false", "-auto-approve"}, varArgs...)
	cmd := exec.Command("terraform", args...)
	cmd.Dir = t.Workspace
	return RunCommand(cmd, t.Logger)
}

func (b *BicepExecutor) ExecuteCreateOperation(step models.Step) (map[string]any, error) {
	log.Printf("Executing Bicep [ ****** Execute Create ******** ]  for resource %s", b.Resource)
	log.Printf("Starting 'deploy' operation for resource: %s", b.Resource)

	log.Printf("Running the deploy with the following payload %v", b)

	output, err := b.Apply()
	if err != nil {
		return nil, fmt.Errorf("error during apply: %v", err)
	}
	return output, nil
}

func CaptureBicepOutputs(workspace string) (map[string]any, error) {
	log.Printf("Capturing Bicep outputs for workspace: %s", workspace)
	cmd := exec.Command("terraform", "output", "-json")
	cmd.Dir = workspace

	outputBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to capture Terraform outputs: %w", err)
	}

	var rawOutputs map[string]map[string]any
	if err := json.Unmarshal(outputBytes, &rawOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse Terraform outputs: %w", err)
	}

	outputs := make(map[string]any)
	for key, details := range rawOutputs {
		if value, ok := details["value"]; ok {
			outputs[key] = value
		}
	}

	return outputs, nil
}
