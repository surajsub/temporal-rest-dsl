package executors

import "context"

//Executor to execute an opa policy

type OPAExecutor struct {
	*ExecutorBase
	PolicyName string
	PolicyPath string
}

func EvaluateWithOPA(ctx context.Context, policyPath string, input map[string]any) (bool, error) {
	return true, nil
}
