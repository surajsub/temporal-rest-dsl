package db
// TODO - Remove this 
const (
	UpdateSubmissionStepStatus = `
		UPDATE submission_steps
		SET status = $1, updated_at = $2
		WHERE provider = $3 AND submission_id = $4 AND step_id = $5
	`

	InsertSubmission = `
		INSERT INTO submissions (account, submitter, project, action, deployment_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	GetSubmissionStepsByStatus = `
		SELECT * FROM submission_steps
		WHERE submission_id = $1 AND status = $2
	`
)
