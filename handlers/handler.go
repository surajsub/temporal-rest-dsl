package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/surajsub/temporal-rest-dsl/db"
	"github.com/surajsub/temporal-rest-dsl/models"
	"github.com/surajsub/temporal-rest-dsl/workflows"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"

	"io"
	"log"
	"net/http"
	"reflect"

	"gopkg.in/yaml.v2"
	"gorm.io/datatypes"
)

type WorkflowLogger struct {
	logger *logrus.Logger
}

const SignalName = "step_control_signal"

func GetWorkflowActivityHistoryHandler(c echo.Context, temporalClient client.Client) error {
	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "workflow_id required"})
	}

	// Optional: Get run_id as query param
	runID := c.QueryParam("run_id")

	service := temporalClient.WorkflowService()
	ctx := context.Background()

	req := &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: "default",
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: workflowID,
			RunId:      runID,
		},
	}

	resp, err := service.GetWorkflowExecutionHistory(ctx, req)
	if err != nil {
		log.Printf("Failed to fetch history: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	activityEvents := []*history.HistoryEvent{}
	for _, event := range resp.History.Events {
		switch event.GetEventType() {
		case enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED,
			enums.EVENT_TYPE_ACTIVITY_TASK_STARTED,
			enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED,
			enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
			enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT,
			enums.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
			activityEvents = append(activityEvents, event)
		}
	}

	return c.JSON(http.StatusOK, activityEvents)
}

func GetSubmissionIDStatus(c echo.Context, temporalClient client.Client) error {
	log.Printf("GetSubmissionIDStatus")
	submission_id := c.Param("submission_id")
	// Parse UUID for validation
	parsedID, err := uuid.Parse(submission_id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid submission ID"})
	}

	var submission db.Submission
	if err := db.GormDB.Preload("Steps").First(&submission, "id = ?", parsedID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Submission not found"})
	}


	// New Query
	// Results are stored in the db as a string so we have to unmarshal them
	var rawresults []db.WorkflowResult
	if err := db.GormDB.Raw(`SELECT step_id,status,step_result from SUBMISSION_STEPS where SUBMISSION_ID=?`, parsedID).Scan(&rawresults).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch step results"})
	}

	var parsedResults []db.StepResult
	for _, r := range rawresults {
		var resultMap map[string]interface{}
		if err := json.Unmarshal([]byte(r.StepResult), &resultMap); err != nil {
			// fallback to nil or empty object
			resultMap = map[string]interface{}{"error": "invalid json"}
		}
		parsedResults = append(parsedResults, db.StepResult{
			StepID:     r.StepID,
			Status: r.Status,
			StepResult: resultMap,
		})
	}

	if temporalClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Provisioning Service is not available",
		})
	}
	// Create a child context
	ctx, cancel := context.WithTimeout(c.Request().Context(), 3* time.Second)
	defer cancel()
	resp, err := temporalClient.DescribeWorkflowExecution(
		ctx,
		//	c.Request().Context(),
		submission.WorkflowID,
		submission.RunID, // empty string "" means latest run
	)

	// if err != nil {
	// 	log.Printf("Error describing workflow [%s]: %v", submission.WorkflowID, err)
	// 	return c.JSON(http.StatusInternalServerError, map[string]interface{}{
	// 		"error":       "Unable to retrieve workflow status",
	// 		"workflow_id": submission,
	// 		"run_id":      submission.RunID,
	// 	})
	// }
	
	if err != nil {
			log.Printf("Error describing workflow [%s]: %v", submission.WorkflowID, err)
			// Fallback response: return DB results and indicate temporal unavailable
			return c.JSON(http.StatusOK, map[string]interface{}{
				"status":          "Unknown (Temporal Unavailable)",
				"temporal_online": false,
				"results":         parsedResults,
			})
		}

	info := resp.WorkflowExecutionInfo
	statusStr := enumspb.WorkflowExecutionStatus_name[int32(info.Status)]
	startTime := info.GetStartTime().AsTime()
	closeTime := info.GetCloseTime().AsTime()
	duration := closeTime.Sub(startTime)
	if info.Status == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		duration = time.Since(startTime)
	}

	// Build base response
		response := map[string]interface{}{
			"status":     statusStr,
			"start_time": startTime,
			"duration":   duration.String(),
			"close_time": closeTime.String(),
		}

		// Add results ONLY if status is COMPLETED
			if info.Status == enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED {
				results := []map[string]interface{}{}
				for _, step := range submission.Steps {
					results = append(results, map[string]interface{}{
						"step_id":     step.StepID,
						"step_result": step.StepResult,
						"status":      step.Status,
					})
				}
				response["results"] = results
			}

			return c.JSON(http.StatusOK, response)

}

func GetWorkflowStatusHandler(c echo.Context, temporalClient client.Client) error {
	workflowID := c.Param("workflow_id")
	if workflowID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing workflow_id",
		})
	}

	runID := c.QueryParam("run_id") // Optional

	resp, err := temporalClient.DescribeWorkflowExecution(
		c.Request().Context(),
		workflowID,
		runID, // empty string "" means latest run
	)

	if err != nil {
		log.Printf("Error describing workflow [%s]: %v", workflowID, err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":       "Unable to retrieve workflow status",
			"workflow_id": workflowID,
			"run_id":      runID,
		})
	}

	info := resp.WorkflowExecutionInfo
	statusStr := enumspb.WorkflowExecutionStatus_name[int32(info.Status)]
	startTime := info.GetStartTime().AsTime()
	closeTime := info.GetCloseTime().AsTime()
	duration := closeTime.Sub(startTime)
	if info.Status == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		duration = time.Since(startTime)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"workflow_id": workflowID,
		"run_id":      info.Execution.GetRunId(),
		"status":      statusStr,
		"start_time":  startTime,
		"duration":    duration.String(),
	})
}

func NewSendSignalHandler(c echo.Context, temporalClient client.Client) error {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	logger.Infof("NewSendSignalHandler")
	var payload models.RetrySignal
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid signal payload",
		})
	}

	submissionID := c.QueryParam("submission_id")
	if submissionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing submission_id",
		})
	}

	// The temporal client should be initialized here

	if temporalClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Temporal Client  is not available",
		})
	}

	log.Printf("NewSendSignalHandler and printing the payload for the retry %v ", payload.Inputs)

	// Based on the submission id , we should query the db and get the corresponding workflow id and runid and use that for our retry

	var submission db.Submission
	if err := db.GormDB.Preload("Steps").First(&submission, "id = ?", submissionID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Submission not found"})
	}

	err := temporalClient.SignalWorkflow(c.Request().Context(), submission.WorkflowID, submission.RunID, SignalName, payload)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to send signal",
		})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"status": "Retry Signal Submitted ",
		//"workflow_id":   workflowID,
		//"run_id":        runID,
		"submitted_by":  submission.Submitter,
		"submission_id": submissionID,
		"step":          payload.StepID,
		"action":        payload.Action,
	})

}

func SendSignalHandler(c echo.Context, temporalClient client.Client) error {
	var payload models.RetrySignal
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid signal payload",
		})
	}

	workflowID := c.QueryParam("workflow_id")
	runID := c.QueryParam("run_id") // Optional
	submissionID := c.QueryParam("submission_id")

	if submissionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing submission_id",
		})
	}

	if temporalClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Temporal client is not available",
		})
	}

	err := temporalClient.SignalWorkflow(c.Request().Context(), workflowID, runID, SignalName, payload)
	if err != nil {
		log.Printf("Failed to signal workflow: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to send signal",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":        "signal sent",
		"workflow_id":   workflowID,
		"run_id":        runID,
		"submission_id": submissionID,
		"step":          payload.StepID,
		"action":        payload.Action,
	})
}

/*
Entry point into the system. This is the wo



*/

func SubmitWorkflowHandler(c echo.Context, temporalClient client.Client) error {

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	contentType := c.Request().Header.Get("Content-Type")
	log.Printf("Content-Type: %v", contentType)
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot read body"})
	}

	var req workflows.WorkflowInput
	switch contentType {
	case "application/json":
		err = json.Unmarshal(body, &req)
	case "application/x-yaml", "text/yaml", "application/yaml":
		err = yaml.Unmarshal(body, &req)
	default:
		log.Printf("Unsupported Content-Type: %s", contentType)
		return c.JSON(http.StatusUnsupportedMediaType, map[string]string{"error": "unsupported content type"})
	}

	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
	}

	var input workflows.WorkflowInput
	err = yaml.Unmarshal(body, &input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid YAML"})
	}

	if input.Action == "" {
		log.Println("Warning: 'Action' field is missing or empty in YAML.")
	}
	requiredFields := []string{"Account", "DeploymentId", "Submitter", "Action", "Project", "WorkflowName"}
	missingFields := checkMissingFields(input, requiredFields)

	if len(missingFields) > 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": fmt.Sprintf("missing required fields: %v", missingFields)})
	}
	workflowOptions := client.StartWorkflowOptions{
		ID:        input.Account + "-" + uuid.NewString(),
		TaskQueue: "customer-task-queue-" + input.Account,
	}

	submissionID := uuid.New()
	input.SubmissionID = submissionID.String()
	workflowLogger := WorkflowLogger{logger: logger}

	roleid := os.Getenv("ROLE_ID")
	secretid := os.Getenv("SECRET_ID")

	input.SecretId = secretid
	input.RoleID = roleid

	we, err := temporalClient.ExecuteWorkflow(context.WithValue(context.Background(), "logger", &workflowLogger), workflowOptions, workflows.TemporalExecutorWorkflow, input)

	if err != nil {
		logger.Errorf("Failed to start workflow: %v", err)

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	desc, err := temporalClient.DescribeWorkflowExecution(context.Background(), we.GetID(), we.GetRunID())
	if err != nil {
		fmt.Println("Failed to describe workflow", err)
		return err
	}
	logger.Infof("Workflow Status %s \n\n", desc.WorkflowExecutionInfo.Status)

	logger.Infof("Workflow started successfully. WorkflowID: %s RunID: %s\n", we.GetID(), we.GetRunID())
	submission := db.Submission{
		ID:           submissionID,
		WorkflowName: input.WorkflowName,
		Account:      input.Account,
		Submitter:    input.Submitter,
		Project:      input.Project,
		Action:       input.Action,
		DeploymentID: input.DeploymentId,
		RunID:        we.GetRunID(),
		WorkflowID:   we.GetID(),
	}

	var steps []db.SubmissionStep
	for _, step := range input.Steps {
		jsonVars, _ := json.Marshal(step.Variables)
		steps = append(steps, db.SubmissionStep{
			ID:            uuid.New(),
			SubmissionID:  we.GetID(),
			StepID:        step.ID,
			Provider:      step.Provider,
			Executor:      step.Executor,
			Resource:      step.Resource,
			Workspace:     step.Workspace,
			Operation:     step.Operation,
			DependsOn:     step.DependsOn,
			Variables:     datatypes.JSON(jsonVars),
			Status:        "PENDING",
			LastUpdatedAt: time.Now(),
			StepResult:    datatypes.JSON(""),
		})
	}

	submission.Steps = steps

	// save to DB
	if err := db.GormDB.WithContext(c.Request().Context()).Create(&submission).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		//"workflow_id":   we.GetID(),
		//"run_id":        we.GetRunID(),
		"submitted_by":    submission.Submitter,
		"submission_id":   submissionID.String(),
		"submission_time": time.Now().Format(time.RFC3339),
	})

}

/*func RetrySubmission(c echo.Context) error {
	ctx := context.Background()
	submissionID := c.Param("submissionID")
	stepID := c.Param("stepID")

	// Parse UUID
	parsedID, err := uuid.Parse(submissionID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid submission ID"})
	}

	var retryReq db.SubmissionRetryInput
	if err := db.GormDB.Preload("Steps").First(&retryReq, "id = ?", parsedID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Submission not found"})
	}

	// Connect to Temporal
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to connect to Temporal"})
	}
	defer temporalClient.Close()

	// Send retry signal
	err = temporalClient.SignalWorkflow(ctx, retryReq.WorkflowID, retryReq.RunID, "retryStep", stepID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": fmt.Sprintf("failed to send retry signal: %v", err)})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Retry signal sent"})

}*/

/*func GetSubmissionStatus(c echo.Context) error {
	subID := c.Param("id")

	// Parse UUID
	parsedID, err := uuid.Parse(subID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid submission ID"})
	}

	var submission db.Submission
	if err := db.GormDB.Preload("Steps").First(&submission, "id = ?", parsedID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Submission not found"})
	}

	type StepStatus struct {
		StepID        string `json:"step_id"`
		Executor      string `json:"executor"`
		Resource      string `json:"resource"`
		Status        string `json:"status"`
		LastUpdatedAt string `json:"last_updated_at"`
		//Variables json.RawMessage `json:"variables"`
	}

	var stepStatuses []StepStatus
	for _, step := range submission.Steps {
		stepStatuses = append(stepStatuses, StepStatus{
			StepID:        step.StepID,
			Executor:      step.Executor,
			Resource:      step.Resource,
			Status:        step.Status,
			LastUpdatedAt: step.LastUpdatedAt.String(),
			//Variables: json.RawMessage(step.Variables),
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"submission_id": submission.ID,
		"workflow_name": submission.WorkflowName,
		"account":       submission.Account,
		"project":       submission.Project,
		"action":        submission.Action,
		"created_at":    submission.CreatedAt.String(),
		"steps":         stepStatuses,
	})
}*/

func checkMissingFields(input any, requiredFields []string) []string {
	var missing []string
	v := reflect.ValueOf(input)

	for _, field := range requiredFields {
		if val := v.FieldByName(field); val.IsValid() && val.Kind() == reflect.String && val.String() == "" {
			missing = append(missing, field)
		}
	}
	return missing
}

func checkValidActionValues(input string) {
	if input != "create" || input != "delete" {

	}
}
func CustomHTTPErrorHandler(err error, c echo.Context) {
	requestID := uuid.New().String()
	c.Set("requestID", requestID)

	// Log the full error with request ID
	c.Logger().Errorf("Request ID: %s | Internal error: %v", requestID, err)

	// Hide internal error message from user
	c.JSON(http.StatusInternalServerError, map[string]interface{}{
		"error":      "Internal server error. Please contact support with the request ID.",
		"request_id": requestID,
	})
}

func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := uuid.New().String()
		c.Set("requestID", requestID)
		c.Response().Header().Set("X-Request-ID", requestID)
		return next(c)
	}
}
