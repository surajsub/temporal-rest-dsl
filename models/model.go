package models

type Step struct {
	ID              string
	DependsOn       []string
	Provider        string
	Resource        string
	Executor        string
	Operation       string
	Workspace       string
	Variables       map[string]any `yaml:"variables" yaml:"variables"`
	TaskDescription string         `yaml:"-"` // Optional, for logging/UI purposes only
	Customer        string         `yaml:"-"` // Don't need to make it mandatory for now
	Provisioner     string         `yaml:"provisioner,omitempty"`
	Submitter       string         `yaml:"submitter,omitempty"`
	Project         string         `yaml:"project,omitempty"`
	ResourceGroup   string         `yaml:"resource_group,omitempty"`
	File            string         `yaml:"file,omitempty"`
	Action          string         `yaml:"action,omitempty" json:"action,omitempty"`
	Activity        string         `yaml:"activity,omitempty" json:"activity,omitempty"`
	DeploymentName  string         `yaml:"deploymentName,omitempty" json:"deploymentName,omitempty"`
	SecretId        string         `yaml:"secret_id,omitempty" json:"secret_id,omitempty"`
	RoleID          string         `yaml:"role_id,omitempty" json:"role_id,omitempty"`
}

type Credentials struct {
	URL      string
	UserName string
	Password string
	Customer string
	Service  string
}

type RetrySignal struct {
	StepID string                 `json:"step_id"`
	Action string                 `json:"action"`
	Inputs map[string]interface{} `json:"inputs"`
}
