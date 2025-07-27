package handlers

type Step struct {
	ID        string         `yaml:"id"`
	Executor  string         `yaml:"executor"`
	Provider  string         `yaml:"provider"`
	Resource  string         `yaml:"resource"`
	Workspace string         `yaml:"workspace"`
	Operation string         `yaml:"operation"`
	DependsOn []string       `yaml:"depends_on"`
	Variables map[string]any `yaml:"variables"`
}

type WorkflowYAML struct {
	WorkflowName string `yaml:"workflow_name"`
	Account      string `yaml:"account"`
	Submitter    string `yaml:"submitter"`
	Project      string `yaml:"project"`
	Action       string `yaml:"action"`
	DeploymentID string `yaml:"deployment_id" json:"deployment_id,omitempty"`
	Steps        []Step `yaml:"steps"`
}

//type WorkflowInput struct {
//	WorkflowName string `yaml:"workflow_name"`
//	Account      string `yaml:"account"`
//	Submitter    string `yaml:"submitter"`
//	Project      string `yaml:"project"`
//	Action       string `yaml:"action"`
//	DeploymentId string `yaml:"deployment_id,omitempty" json:"deployment_id,omitempty"`
//	Steps        []models.Step
//}
