package executors

import (
	"context"
	"errors"
	"fmt"

	"encoding/json"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/surajsub/temporal-rest-dsl/models"

	"github.com/google/go-github/github"
)

type GitExecutor struct {
	*ExecutorBase
	Submitter string
	Project   string
}

// Constructor for GitExecutor
func NewGitExecutor(customer, workspace, provider, resource, provisioner, action, operation, submitter, project string) *GitExecutor {
	return &GitExecutor{
		ExecutorBase: NewExecutorBase(customer, workspace, provider, resource, provisioner, action, operation),
		Submitter:    submitter,
		Project:      project,
	}
}

func (g *GitExecutor) Execute(step models.Step, executor string, payload map[string]any) (map[string]any, error) {
	g.Logger.Infof("Executing GitExecutor with action %s and operation %s", g.Action, step.Operation)

	if step.Operation == "create_issue" && g.Action == "create" {

		issueURL,issueNumber, err := g.CreateGitHubIssue(payload)
		if err != nil {
			return nil, err
		}
		g.Logger.Infof("Issue URL is %s\n", issueURL)

		return map[string]any{"issue_url": issueURL,"issue_id":issueNumber}, nil
	}

	if step.Operation == "poll_issue_status" && g.Action == "create" {
		err := g.PollGitHubIssueStatus(payload, payload["token"].(string))
		if err != nil {
			return nil, err
		}
		return map[string]any{"status": "approved"}, nil
	}

	if g.Action == "delete" {
		g.Logger.Infof("Executing GitHub delete operation for customer %s\n - DO NOTHING", g.Customer)

		return nil, nil
	}
	return nil, errors.New("unsupported operation for GitExecutor")
}

func (g *GitExecutor) CheckOperation(step models.Step) error {
	// Implement any necessary checks for the Git operation

	if step.Operation == "create_issue" && step.Action == "create" {
		g.Logger.Infof("Checking create_issue operation for GitExecutor")
		return nil
	}
	if step.Operation == "poll_issue_status" && step.Action == "create" {
		g.Logger.Infof("Checking poll_issue_status operation for GitExecutor")
		return nil
	}

	return nil
}

// ValidateOperation ensures only supported Git operations are allowed
func (g *GitExecutor) ValidateOperation(step models.Step) error {

	if step.Operation == "create_issue" {
		g.Logger.Infof("Validating operation %s for GitExecutor", step.Operation)
		return nil
	}
	if step.Operation == "progress_issue_status" {
		g.Logger.Infof("Validating operation %s for GitExecutor", step.Operation)
		return nil
	}
	return nil
}

func (g *GitExecutor) CreateGitHubIssue(payload map[string]any) (string,string, error) {
	ctx := context.Background()
	client := CreateGitHubClient(payload["token"].(string))
	g.Logger.Infof("Creating GitHub issue...for Project %s", g.Project)

	var gitHubIssueTitle = fmt.Sprintf("Issue Created for Account %s for Project %s\n . The requester is %s", g.Customer, g.Project, g.Submitter)
	repoOwner := payload["repo_owner"].(string)
	repoName := payload["repo_name"].(string)

	title := gitHubIssueTitle
	body := payload["body"].(string)

	issueRequest := &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}

	issue, _, err := client.Issues.Create(ctx, repoOwner, repoName, issueRequest)
	if err != nil {
		g.Logger.Error("Failed to create GitHub issue:", err)
		return "", "",err
	}

	log.Printf("How to make this work ------ %s", issue.GetNumber())

	apiIssue, _, err := client.Issues.Get(ctx, repoOwner, repoName, int(issue.GetNumber()))
	if err != nil {
		g.Logger.Error("Failed to fetch GitHub issue:", err)
		return "", "", err
	}

	// Check here
	log.Printf("the issue url for the api is %s", apiIssue)
	return issue.GetHTMLURL(),strconv.Itoa(issue.GetNumber()), nil
}


func (g *GitExecutor) PollGitHubIssueStatus(payload map[string]any, token string) error {
	ctx := context.Background()
	client := CreateGitHubClient(payload["token"].(string))
	g.Logger.Infof("Polling GitHub issue...for Project %s", g.Project)

	// Parse issue URL
	//
	repoOwner := payload["repo_owner"].(string)
	g.Logger.Infof("Printing the data repoowners is %s", repoOwner)
	repoName := payload["repo_name"].(string)

	g.Logger.Infof("and the reponame is %s", repoName)
	issueNumber := payload["issue_id"]
	g.Logger.Infof("Printing the issue number that's part of the payload %s",issueNumber);
	if (issueNumber == nil){
		g.Logger.Error("Issue number is not provided...")
	}
	var convertedIssue int
	fmt.Printf("issueNumber type: %T, value: %#v\n", issueNumber, issueNumber)
		
	switch v := issueNumber.(type) {
case string:
	k,err := strconv.Atoi(v)
	if err == nil{
		convertedIssue = int(k)
		fmt.Println("Conversion successful from string:", convertedIssue)
	}
	
case float64:
    convertedIssue = int(v)
    fmt.Println("Conversion successful from float64:", convertedIssue)
case int:
    convertedIssue = v
    fmt.Println("Conversion successful from int:", convertedIssue)
case int64:
    convertedIssue = int(v)
    fmt.Println("Conversion successful from int64:", convertedIssue)
case json.Number:
    i, err := v.Int64()
    if err == nil {
        convertedIssue = int(i)
        fmt.Println("Conversion successful from json.Number:", convertedIssue)
    } else {
        fmt.Println("Failed to convert json.Number:", err)
    }
default:
    fmt.Println("Unsupported type:", reflect.TypeOf(issueNumber))
}

g.Logger.Infof("Printing the data issue number is %d", convertedIssue)
	
	if v, ok := issueNumber.(float64); ok {
		convertedIssue = int(v)
            fmt.Println("Conversion successful:", convertedIssue)
        } else {
            fmt.Println("Conversion failed: value is not an int")
        }

	g.Logger.Infof("Printing the data issue number  is %d", convertedIssue)


	// Poll the issue status
	for {
		issue, _, err := client.Issues.Get(ctx, repoOwner, repoName, convertedIssue)
		if err != nil {
			g.Logger.Error("Failed to fetch GitHub issue:", err)
			return err
		}

		if issue.GetState() == "closed" {
			comments, _, err := client.Issues.ListComments(ctx, repoOwner, repoName, convertedIssue, nil)
			if err != nil {
				log.Println("Failed to fetch GitHub comments:", err)
				return err
			}

			// Check for approval
			for _, comment := range comments {
				if strings.Contains(strings.ToLower(comment.GetBody()), "approved") {
					g.Logger.Infof("GitHub issue approved")
					return nil
				}
			}
			log.Println("GitHub issue closed but no approval found")
			return errors.New("approval not found")
		}

		g.Logger.Infof("GitHub issue not yet closed, polling again...")
		time.Sleep(2 * time.Minute)
	}
}
