package executors

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

const (
	TERRAFORM = "terraform"
	OPENTOFU  = "opentofu"
	INFRACOST = "infracost"
	GIT       = "git"
	BICEP     = "bicep"
	OPA       = "opa"
	VAULT     = "vault"

	DESTROY         = "destroy"
	CostEstimate    = "cost_estimate"
	REPORT          = "report"
	CreateIssue     = "create_issue"
	PollIssueStatus = "poll_issue_status"
	CREATE          = "create"
	DELETE          = "delete"
	GETCREDS        = "getcreds"
)

type ExecutorConstructor func(config map[string]any) Executor

// Registry to store executor constructors and supported operations
var registry = make(map[string]ExecutorConstructor)
var supportedOperations = make(map[string][]string)

// RegisterExecutor registers an executor and its supported operations
func RegisterExecutor(name string, constructor ExecutorConstructor, operations []string) {
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("Executor %s is already registered", name))
	}
	registry[name] = constructor
	supportedOperations[name] = operations
}

func createBase(config map[string]any) *ExecutorBase {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{}) // Structured logging
	logger.SetLevel(logrus.InfoLevel)
	logger.Debugf("Initializing ExecutorBase with Logger")
	return &ExecutorBase{
		Customer:    config["customer"].(string),
		Workspace:   config["workspace"].(string),
		Provider:    config["provider"].(string),
		Resource:    config["resource"].(string),
		Provisioner: config["provisioner"].(string),
		Action:      config["action"].(string),
		Variables:   config["variables"].(map[string]any),

		Logger: logger,
	}
}

// GetExecutor retrieves an executor from the registry
func GetExecutor(name string, config map[string]any) (Executor, error) {

	constructor, exists := registry[name]
	if !exists {
		return nil, fmt.Errorf("executor %s not found", name)
	}
	return constructor(config), nil
}

// Init functions

func init() {
	// Register the Executors
	RegisterExecutor(TERRAFORM, func(config map[string]any) Executor {
		return &TerraformExecutor{ExecutorBase: createBase(config)}
	}, []string{CREATE, DELETE})
	RegisterExecutor(INFRACOST, func(config map[string]any) Executor {
		return &InfraCostExecutor{ExecutorBase: createBase(config), Submitter: config["submitter"].(string), Project: config["project"].(string)}
	}, []string{CostEstimate, REPORT})
	RegisterExecutor(GIT, func(config map[string]any) Executor {
		return &GitExecutor{ExecutorBase: createBase(config), Submitter: config["submitter"].(string), Project: config["project"].(string)}
	}, []string{CreateIssue, PollIssueStatus})
	RegisterExecutor(OPENTOFU, func(config map[string]any) Executor { return &OpenTFExecutor{ExecutorBase: createBase(config)} }, []string{CREATE, DELETE})
	RegisterExecutor(VAULT, func(config map[string]any) Executor {
		return &VaultExecutor{ExecutorBase: createBase(config)}
	}, []string{GETCREDS})
	RegisterExecutor(BICEP, func(config map[string]any) Executor {
		return &BicepExecutor{ExecutorBase: createBase(config), DeploymentName: config["deploymentName"].(string), File: config["file"].(string), ResourceGroup: config["resource_group"].(string)}
	}, []string{CREATE, DESTROY})

}
