package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"log"
	"os"
)

func SaveStateToStorage(ctx context.Context, state map[string]map[string]any, project string) error {
	// Convert state to JSON
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
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	filePath := fmt.Sprintf("./resources/aws/terraform/%s.json", project)

	// Check if file exists
	if _, err := os.Stat(filePath); err == nil {
		// File exists, so remove it
		logger.Infof("Removing existing file: %s", filePath)
		err := os.Remove(filePath)
		if err != nil {
			return fmt.Errorf("failed to remove existing file: %w", err)
		}
		log.Printf("Existing file %s removed successfully\n", filePath)
	} else if !os.IsNotExist(err) {
		// Some other error occurred (e.g., permission issues)
		return fmt.Errorf("error checking file existence: %w", err)
	}
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write state to file: %w", err)
	}
	logger.Infof("State saved to %s", filePath)
	return nil

}
