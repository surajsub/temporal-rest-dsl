package db

import (
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"time"
)

type SubmissionRetryInput struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;"`
	Account    string
	WorkflowID string
	RunID      string
	StepID     string
	Status     string
	RetryCount int
	LastError  string
}

type Submission struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	WorkflowName string
	Account      string
	Submitter    string
	Project      string
	Action       string
	DeploymentID string
	RunID        string
	WorkflowID   string
	CreatedAt    time.Time
	Steps        []SubmissionStep `gorm:"foreignKey:SubmissionID"`
}

type SubmissionStep struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	SubmissionID  string
	StepID        string
	Provider      string
	Executor      string
	Resource      string
	Workspace     string
	Operation     string
	Variables     datatypes.JSON
	DependsOn     pq.StringArray `gorm:"type:text[]"`
	LastUpdatedAt time.Time
	Status        string // PENDING, SUCCESS, FAILED
	StepResult    datatypes.JSON
}


//  
 type WorkflowResult struct {
	StepID     string `json:"step_id"`
	Status string `json:"status"`
	StepResult string `json:"step_result"`
}
// To return as json object
type StepResult struct {
	StepID     string                 `json:"step_id"`
	Status string				 `json:"step_status"`
	StepResult map[string]interface{} `json:"step_result"`
}