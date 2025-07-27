package executors

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/models"

	"log"
)

type Executor interface {
	Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error)
	ValidateOperation(step models.Step) error
}

type ExecutorBase struct {
	Customer            string
	Workspace           string
	Provider            string
	Resource            string
	Provisioner         string
	Variables           map[string]any
	SupportedOperations []string
	Action              string
	Operation           string
	Logger              *logrus.Logger
}

// Constructor for ExecutorBase
func NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation string) *ExecutorBase {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{}) // Structured logging
	logger.SetLevel(logrus.InfoLevel)

	return &ExecutorBase{
		Customer:            customer,
		Workspace:           workspace,
		Provider:            provider,
		Resource:            resource,
		Provisioner:         provisioner,
		Variables:           make(map[string]any),
		SupportedOperations: []string{},
		Action:              action,
		Operation:           operation,
		Logger:              logger,
	}
}

func (e *ExecutorBase) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {
	// Default implementation

	log.Printf("Executing operation %s for executor %s", e.Action, e.Provider)
	return nil, fmt.Errorf("execute not implemented for %s", executor)
}

func (e *ExecutorBase) ValidateOperation(operation string) error {
	// Default validation logic
	return nil
}
