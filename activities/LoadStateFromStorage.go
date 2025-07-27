package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

// LoadStateFromStorage retrieves the workflow state from a storage location (e.g., S3).
func LoadStateFromStorage(ctx context.Context, project string) (map[string]map[string]any, error) {
	loggerInterface := ctx.Value("logger")
	var logger *logrus.Logger
	if loggerInterface != nil {
		logger = loggerInterface.(*WorkflowLogger).logger
	} else {
		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetLevel(logrus.InfoLevel)
		logger.Warn("Using fallback logger as no logger was passed")
	}
	if loggerInterface != nil {
		fmt.Printf("Logger is not nil")
	}

	// TODO - This path is hardcoded to the location of the terraform resource.
	// Since it's owned by the platform we can persist this file in our local db 
	// For now the implementation stores it on the local system
	// 
	filePath := fmt.Sprintf("./resources/aws/terraform/%s.json", project)
	logger.Infof("Loading state from %s", filePath)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state map[string]map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}
	logger.Info("State loaded successfully")
	return state, nil
}
