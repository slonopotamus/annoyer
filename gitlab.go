package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"gitlab.com/gitlab-org/api/client-go/v3"
)

type createMergeRequestDiscussionParams struct {
	MergeRequest int64                                      `json:"merge_request_id"`
	Options      gitlab.CreateMergeRequestDiscussionOptions `json:"options"`
}

type createMergeRequestNoteParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.CreateMergeRequestNoteOptions `json:"options"`
}

type getFileParams struct {
	Path    string                   `json:"path"`
	Options gitlab.GetRawFileOptions `json:"options"`
}

type getMergeRequestChangesParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestChangesOptions `json:"options"`
}

type getMergeRequestCommitsParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestCommitsOptions `json:"options"`
}

type getTraceFileParams struct {
	JobID int64 `json:"job_id"`
}

type listMergeRequestDiffsParams struct {
	MergeRequest int64                               `json:"merge_request_id"`
	Options      gitlab.ListMergeRequestDiffsOptions `json:"options"`
}

type listMergeRequestDiscussionsParams struct {
	MergeRequest int64                                     `json:"merge_request_id"`
	Options      gitlab.ListMergeRequestDiscussionsOptions `json:"options"`
}

type listMergeRequestNotesParams struct {
	MergeRequest int64                               `json:"merge_request_id"`
	Options      gitlab.ListMergeRequestNotesOptions `json:"options"`
}

type listMergeRequestPipelinesParams struct {
	MergeRequest int64 `json:"merge_request_id"`
}

type listPipelineJobsParams struct {
	PipelineID int64                  `json:"pipeline_id"`
	Options    gitlab.ListJobsOptions `json:"options"`
}

type listTreeParams struct {
	Options gitlab.ListTreeOptions `json:"options"`
}

type searchParams struct {
	Query   string               `json:"query"`
	Options gitlab.SearchOptions `json:"options"`
}

func marshalToBetaToolResult(v any) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	result, err := json.Marshal(v)
	if err != nil {
		return anthropic.BetaToolResultBlockParamContentUnion{}, err
	}

	return anthropic.BetaToolResultBlockParamContentUnion{OfText: &anthropic.BetaTextBlockParam{Text: string(result)}}, nil
}

func newGitLabTools(client *gitlab.Client, pid any) ([]anthropic.BetaTool, error) {
	createDiscussionTool, err := toolrunner.NewBetaToolFromJSONSchema("createDiscussion", "Creates a merge request discussion", func(ctx context.Context, params createMergeRequestDiscussionParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		_, _, err := client.Discussions.CreateMergeRequestDiscussion(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		return anthropic.BetaToolResultBlockParamContentUnion{}, err
	})
	if err != nil {
		return nil, err
	}

	createMergeRequestNoteTool, err := toolrunner.NewBetaToolFromJSONSchema("createMergeRequestNote", "Posts a top-level comment on the merge request, not tied to a diff line", func(ctx context.Context, params createMergeRequestNoteParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		_, _, err := client.Notes.CreateMergeRequestNote(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		return anthropic.BetaToolResultBlockParamContentUnion{}, err
	})
	if err != nil {
		return nil, err
	}

	getFileTool, err := toolrunner.NewBetaToolFromJSONSchema("getFile", "Gets file contents by path", func(ctx context.Context, params getFileParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.RepositoryFiles.GetRawFile(pid, params.Path, &params.Options, gitlab.WithContext(ctx))
		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(result)},
		}, err
	})
	if err != nil {
		return nil, err
	}

	getMergeRequestChangesTool, err := toolrunner.NewBetaToolFromJSONSchema("getMergeRequestChanges", "Gets details of a merge request", func(ctx context.Context, params getMergeRequestChangesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.GetMergeRequestChanges(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getMergeRequestCommitsTool, err := toolrunner.NewBetaToolFromJSONSchema("getMergeRequestCommits", "gets a list of merge request commits", func(ctx context.Context, params getMergeRequestCommitsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.GetMergeRequestCommits(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getTraceFileTool, err := toolrunner.NewBetaToolFromJSONSchema("getTraceFile", "Gets the log output of a CI job", func(ctx context.Context, params getTraceFileParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Jobs.GetTraceFile(pid, params.JobID, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}
		trace, err := io.ReadAll(result)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(trace)},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestDiffsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestDiffs", "List merge request changed files", func(ctx context.Context, params listMergeRequestDiffsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.ListMergeRequestDiffs(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestDiscussionsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestDiscussions", "Lists existing merge request discussions, including inline review comments", func(ctx context.Context, params listMergeRequestDiscussionsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Discussions.ListMergeRequestDiscussions(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestNotesTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestNotes", "Lists all notes (comments) on the merge request", func(ctx context.Context, params listMergeRequestNotesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Notes.ListMergeRequestNotes(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestPipelinesTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestPipelines", "Lists the CI pipelines for the merge request", func(ctx context.Context, params listMergeRequestPipelinesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.ListMergeRequestPipelines(pid, params.MergeRequest, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listPipelineJobsTool, err := toolrunner.NewBetaToolFromJSONSchema("listPipelineJobs", "Lists the jobs of a CI pipeline", func(ctx context.Context, params listPipelineJobsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Jobs.ListPipelineJobs(pid, params.PipelineID, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listTreeTool, err := toolrunner.NewBetaToolFromJSONSchema("listTree", "Lists the repository file tree", func(ctx context.Context, params listTreeParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Repositories.ListTree(pid, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	searchBlobsTool, err := toolrunner.NewBetaToolFromJSONSchema("searchBlobs", "Searches files matching query", func(ctx context.Context, params searchParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Search.BlobsByProject(pid, params.Query, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	return []anthropic.BetaTool{
		createDiscussionTool,
		createMergeRequestNoteTool,
		getFileTool,
		getMergeRequestChangesTool,
		getMergeRequestCommitsTool,
		getTraceFileTool,
		listMergeRequestDiffsTool,
		listMergeRequestDiscussionsTool,
		listMergeRequestNotesTool,
		listMergeRequestPipelinesTool,
		listPipelineJobsTool,
		listTreeTool,
		searchBlobsTool,
	}, nil
}
