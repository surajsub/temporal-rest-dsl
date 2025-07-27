package workflows

import (
	"github.com/surajsub/temporal-rest-dsl/models"
)

type WorkflowInput struct {
	WorkflowName string `yaml:"workflow_name"`
	Account      string `yaml:"account"`
	Submitter    string `yaml:"submitter"`
	Project      string `yaml:"project"`
	Action       string `yaml:"action"`
	SubmissionID string `yaml:"submission_id"`
	DeploymentId string `yaml:"deployment_id,omitempty" json:"deployment_id,omitempty"`
	Steps        []models.Step
	SecretId     string `yaml:"secret_id,omitempty" json:"secret_id,omitempty"`
	RoleID       string `yaml:"role_id,omitempty" json:"role_id,omitempty"`
}

type UpdateInputSignal struct {
	StepID   string         `json:"step_id"`
	NewInput map[string]any `json:"new_input"`
}
