package executors

import (
	"github.com/surajsub/temporal-rest-dsl/models"
)

//Executor to execute an opa policy

type GLPIExecutor struct {
	*ExecutorBase
	TicketID     string
	TicketStatus string
}

// Constructor for GLPIExecutor
func NewGLPIExecutor(customer, workspace, provider, resource, provisioner, action, operation string) *GLPIExecutor {
	return &GLPIExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
	}
}

func (g *GLPIExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {

	return nil, nil
}

func (g *GLPIExecutor) ValidateOperation(step models.Step) error {

	// What does the GLPI Executor need to validate
	g.Logger.Info("Validating operation as provided in the input %s", step.Operation)
	return nil
}
