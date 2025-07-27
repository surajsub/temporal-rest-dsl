package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/surajsub/temporal-rest-dsl/db"
	"github.com/surajsub/temporal-rest-dsl/models"
	"log"
	"time"
)

// DBActivity  /*
func DBActivity(ctx context.Context, step models.Step, submission string, activity string, status string, result map[string]any) (map[string]any, error) {
	logger := GetDSLActivityLogger(ctx)

	logger.Infof("[********* Running activity: %s ********* with result %v]", activity, result)

	logger.Infof("Executing  %s for %s in workspace %s with variables: %v", step.Operation, step.Resource, step.Workspace, step.Variables)

	logger.Infof("[******* Running the DB Activity %s ********]", step.Activity)

	logger.Infof("[******Update SUBMISSIONS set status= %s where SUBMISSION_ID= %s and step_id=%s ]", status, submission, "COMPLETE")

	// Call the deployResource function and get the output
	//return nil, nil

	if result == nil {
		logger.Debugf("Result is nil so it's an INSERT ************")
	}
	ctx = context.Background()
	conn := db.NewPostgresManager() // Update

	var outputResult string
	if result != nil {
		log.Printf("result: %v", result)
	}
	// Convert map to json string
	jsonStr, err := json.Marshal(result)
	if err != nil {
		fmt.Println(err)
	}
	outputResult = string(jsonStr)
	err = conn.Update(ctx, "submission_steps",
		map[string]any{ // SET clause
			"status":          "SUCCESS",
			"last_updated_at": time.Now(),

			"step_result": outputResult,
		},
		map[string]any{ // WHERE clause
			"provider":      step.Provider,
			"submission_id": submission,
			"step_id":       step.ID,
		})

	if err != nil {
		logger.Errorf("Failed to perform the update %v", err)
	}

	logger.Infof("the connection is %s", conn)
	return nil, nil
}
