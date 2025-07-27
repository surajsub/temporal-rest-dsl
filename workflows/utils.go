package workflows

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/models"
	"go.temporal.io/sdk/workflow"
	"io/ioutil"
	"regexp"
	"strings"
)

type WorkflowLogger struct {
	logger *logrus.Logger
}

func ProcessStepVariables(step models.Step, results map[string]any) map[string]any {
	processedVariables := make(map[string]any)

	for key, value := range step.Variables {
		switch v := value.(type) {
		case string:
			// Perform variable substitution if needed

			//processedVariables[key] = interpolateVariables(v, results)
			processedVariables[key] = replaceVariables(v, results)
		case []any:
			// Handle arrays: add quotes to each element
			var processedArray []string
			for _, item := range v {
				processedArray = append(processedArray, fmt.Sprintf("\"%s\"", item))
			}
			// Join and wrap the array in Terraform format
			processedVariables[key] = fmt.Sprintf("[%s]", strings.Join(processedArray, ", "))

		default:
			strValue, isString := value.(string)
			if isString && strings.Contains(strValue, "${") {
				processedVariables[key] = formatArrayForTerraform(replaceVariables(strValue, results))
			} else {
				processedVariables[key] = value
			}

		}
	}
	//fmt.Printf("Processed variables: %v\n", processedVariables)
	return processedVariables
}

func formatArrayForTerraform(value any) string {
	switch v := value.(type) {
	case []any:
		// Convert array elements to a comma-separated string
		var elements []string
		for _, item := range v {
			elements = append(elements, fmt.Sprintf("\"%v\"", item))
		}
		return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
	case string:
		// Return as-is for simple strings
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

func replaceVariables(input string, results map[string]any) string {
	// Use a regex to find placeholders in the format ${key.subkey}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	//fmt.Printf("The results from the call in the replaceVariables function are: %v \n", results)
	// Replace all placeholders in the input string
	replaced := re.ReplaceAllStringFunc(input, func(placeholder string) string {
		// Extract the key (e.g., "create_vpc.vpc_id")
		key := re.FindStringSubmatch(placeholder)[1]

		// Split the key into parts (e.g., "create_vpc" and "vpc_id")
		parts := strings.Split(key, ".")
		if len(parts) < 2 {
			// Invalid format, return the original placeholder
			return placeholder
		}

		// Look up the step and output in the results map
		stepID := parts[0]
		outputKey := parts[1]

		if stepResult, exists := results[stepID]; exists {
			if outputMap, ok := stepResult.(map[string]any); ok {
				if value, found := outputMap[outputKey]; found {
					switch v := value.(type) {
					case []any:
						// Format array values for Terraform
						var formattedArray []string
						for _, item := range v {
							formattedArray = append(formattedArray, fmt.Sprintf("\"%v\"", item))
						}
						return fmt.Sprintf("[%s]", strings.Join(formattedArray, ", "))
					default:
						// Return the value as a string
						return fmt.Sprintf("%v", v)
					}
				}
			}
		}

		// If no replacement found, return the original placeholder
		return placeholder
	})

	return replaced
}

func SaveStateToStorage(state map[string]map[string]any, project string) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	filePath := fmt.Sprintf("./storage/%s.json", project)
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state to file: %w", err)
	}

	return nil
}

func LoadStateFromStorage(project string) (map[string]map[string]any, error) {
	filePath := fmt.Sprintf("./storage/%s.json", project)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state map[string]map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state, nil
}

func PrepareStep(step models.Step, input WorkflowInput, results map[string]map[string]any) models.Step {
	step.Customer = input.Account
	step.Project = input.Project
	step.Submitter = input.Submitter
	step.Action = input.Action
	step.Variables = resolveVariables(step.Variables, results)
	return step
}

func GetDSLLogger(ctx workflow.Context) *logrus.Logger {
	if loggerInterface := ctx.Value("logger"); loggerInterface != nil {
		return loggerInterface.(*WorkflowLogger).logger
	}
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)
	logger.Warn("Using DSL Logger as no logger was passed")
	return logger
}

func extractDependencyID(variable string) string {
	// Extract dependency ID, e.g., "${create_vpc.vpc_id}" -> "create_vpc"
	parts := strings.Split(strings.Trim(variable, "${}"), ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func extractDependencyKey(variable string) string {
	// Extract dependency key, e.g., "${create_vpc.vpc_id}" -> "vpc_id"
	parts := strings.Split(strings.Trim(variable, "${}"), ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func WaitForDependencies(ctx workflow.Context, step models.Step, results map[string]map[string]any) error {
	for _, dep := range step.DependsOn {
		if _, exists := results[dep]; !exists {
			return workflow.NewContinueAsNewError(ctx, TemporalExecutorWorkflow, step)
		}
	}
	return nil
}

func resolveVariables(stepVars map[string]any, workflowVars map[string]map[string]any) map[string]any {
	resolvedVars := make(map[string]any)

	for key, value := range stepVars {
		if strVal, ok := value.(string); ok {

			// Check if the value is a template variable and replace it
			resolvedVal := variableRegex.ReplaceAllStringFunc(strVal, func(match string) string {
				depID := extractDependencyID(match)
				depKey := extractDependencyKey(match)

				if depOutput, exists := workflowVars[depID][depKey]; exists {
					// **Fix:** Preserve lists instead of converting them to strings
					switch v := depOutput.(type) {
					case []string:
						return fmt.Sprintf("%q", v) // Ensure list stays as a list
					case []any:
						strArray := make([]string, len(v))
						for i, item := range v {
							strArray[i] = fmt.Sprintf("%q", item)
						}
						return fmt.Sprintf("[%s]", strings.Join(strArray, ", "))
					default:
						return fmt.Sprintf("%v", depOutput) // Convert non-list values to string
					}
				}
				return match // Keep unresolved if value is not found
			})

			resolvedVars[key] = resolvedVal
		} else {
			// Preserve non-string values as they are
			resolvedVars[key] = value
		}
	}

	return resolvedVars
}
