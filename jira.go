package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	jira "github.com/ctreminiom/go-atlassian/v2/jira/v2"
	model "github.com/ctreminiom/go-atlassian/v2/pkg/infra/models"
)

type getJiraIssueParams struct {
	IssueKey string   `json:"issue_key"`
	Fields   []string `json:"fields"`
	Expand   []string `json:"expand"`
}

type searchJiraIssuesParams struct {
	JQL           string   `json:"jql"`
	Fields        []string `json:"fields"`
	Expand        []string `json:"expand"`
	MaxResults    int      `json:"max_results"`
	NextPageToken string   `json:"next_page_token"`
}

type listJiraIssueCommentsParams struct {
	IssueKey   string   `json:"issue_key"`
	OrderBy    string   `json:"order_by"`
	Expand     []string `json:"expand"`
	StartAt    int      `json:"start_at"`
	MaxResults int      `json:"max_results"`
}

func newJiraTools(client *jira.Client) ([]anthropic.BetaTool, error) {
	getJiraIssueTool, err := toolrunner.NewBetaToolFromJSONSchema("getJiraIssue", "Retrieves the details of a single Jira issue by its key (e.g. \"PROJ-123\"), including its summary, description, status, assignee, labels, and more. Pass fields to fetch only specific fields, or leave it empty to fetch all of them", func(ctx context.Context, params getJiraIssueParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, resp, err := client.Issue.Get(ctx, params.IssueKey, params.Fields, params.Expand)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, jiraError(err, resp)
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	searchJiraIssuesTool, err := toolrunner.NewBetaToolFromJSONSchema("searchJiraIssues", "Searches Jira issues using a JQL query, e.g. \"project = PROJ AND text ~ \\\"login bug\\\"\" or \"project = PROJ AND status = 'In Progress' AND assignee = currentUser()\". Supports fields/expand filters and max_results/next_page_token pagination (50 results per page by default)", func(ctx context.Context, params searchJiraIssuesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		maxResults := params.MaxResults
		if maxResults <= 0 {
			maxResults = 50
		}

		result, resp, err := client.Issue.Search.SearchJQL(ctx, params.JQL, params.Fields, params.Expand, maxResults, params.NextPageToken)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, jiraError(err, resp)
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listJiraIssueCommentsTool, err := toolrunner.NewBetaToolFromJSONSchema("listJiraIssueComments", "Lists the comments on a Jira issue by its key. Use order_by to sort by creation or update time, and start_at/max_results for pagination (20 comments per page by default)", func(ctx context.Context, params listJiraIssueCommentsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		maxResults := params.MaxResults
		if maxResults <= 0 {
			maxResults = 20
		}

		result, resp, err := client.Issue.Comment.Gets(ctx, params.IssueKey, params.OrderBy, params.Expand, params.StartAt, maxResults)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, jiraError(err, resp)
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	return []anthropic.BetaTool{
		getJiraIssueTool,
		searchJiraIssuesTool,
		listJiraIssueCommentsTool,
	}, nil
}

// jiraError wraps the Jira client's sentinel error with the HTTP status code and the
// response body, which carries Jira's actual error details (e.g. errorMessages).
// Without this the model would only see a bare message like "no atlassian resource found".
func jiraError(err error, resp *model.ResponseScheme) error {
	if resp == nil {
		return err
	}

	body := strings.TrimSpace(resp.Bytes.String())
	if body == "" {
		return fmt.Errorf("%w (status %d)", err, resp.Code)
	}

	return fmt.Errorf("%w (status %d): %s", err, resp.Code, body)
}
