package main

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"gitlab.com/gitlab-org/api/client-go/v3"
)

type approveMergeRequestParams struct {
	MergeRequest int64                             `json:"merge_request_id"`
	Options      gitlab.ApproveMergeRequestOptions `json:"options"`
}

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

type getMergeRequestParams struct {
	MergeRequest int64                          `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestsOptions `json:"options"`
}

type getMergeRequestCommitsParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestCommitsOptions `json:"options"`
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

type resolveMergeRequestDiscussionParams struct {
	MergeRequest int64  `json:"merge_request_id"`
	DiscussionID string `json:"discussion_id"`
}

type searchParams struct {
	Query   string               `json:"query"`
	Options gitlab.SearchOptions `json:"options"`
}

type unapproveMergeRequestParams struct {
	MergeRequest int64 `json:"merge_request_id"`
}

func marshalToBetaToolResult(v any) (anthropic.BetaToolResultBlockParamContentUnion, error) {
	result, err := json.Marshal(v)
	if err != nil {
		return anthropic.BetaToolResultBlockParamContentUnion{}, err
	}

	return anthropic.BetaToolResultBlockParamContentUnion{OfText: &anthropic.BetaTextBlockParam{Text: string(result)}}, nil
}

func newGitLabTools(client *gitlab.Client, pid any) ([]anthropic.BetaTool, error) {
	approveMergeRequestTool, err := toolrunner.NewBetaToolFromJSONSchema("approveMergeRequest", "Approves a merge request. Pass sha (the MR's HEAD commit) if you want to bind the approval to a specific commit", func(ctx context.Context, params approveMergeRequestParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequestApprovals.ApproveMergeRequest(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	createDiscussionTool, err := toolrunner.NewBetaToolFromJSONSchema("createDiscussion", "Creates a new thread (discussion) on a merge request; replies can be added to it later. To place it inline on a diff, set position: old_path and new_path are both required; use new_line for added lines, old_line for removed lines, both for unchanged lines, and take base_sha/start_sha/head_sha from the diff_refs field returned by getMergeRequest. To reply to an existing thread, pass in_reply_to_discussion_id. Returns the created discussion with its ID, which can be used to resolve it or reply to it", func(ctx context.Context, params createMergeRequestDiscussionParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Discussions.CreateMergeRequestDiscussion(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	createMergeRequestNoteTool, err := toolrunner.NewBetaToolFromJSONSchema("createMergeRequestNote", "Creates a top-level note on a merge request. Notes are not attached to specific lines; use createDiscussion for inline diff comments. Returns the created note with its ID", func(ctx context.Context, params createMergeRequestNoteParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Notes.CreateMergeRequestNote(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getFileTool, err := toolrunner.NewBetaToolFromJSONSchema("getFile", "Retrieves the raw contents of a file at the given path. Pass ref (branch, tag, or commit) to read a specific version; defaults to the project HEAD. Use the MR's source branch to read files as changed by the MR", func(ctx context.Context, params getFileParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.RepositoryFiles.GetRawFile(pid, params.Path, &params.Options, gitlab.WithContext(ctx))
		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(result)},
		}, err
	})
	if err != nil {
		return nil, err
	}

	getMergeRequestTool, err := toolrunner.NewBetaToolFromJSONSchema("getMergeRequest", "Retrieves information about a single merge request, including its state, title, description, source/target branches, and the diff_refs SHAs needed for inline comments. Does not include the file diffs themselves; use listMergeRequestDiffs for those", func(ctx context.Context, params getMergeRequestParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.GetMergeRequest(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getMergeRequestCommitsTool, err := toolrunner.NewBetaToolFromJSONSchema("getMergeRequestCommits", "Retrieves commits for a merge request", func(ctx context.Context, params getMergeRequestCommitsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.GetMergeRequestCommits(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestDiffsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestDiffs", "Lists the diffs of files changed in a merge request. Subject to diff size limits: for very large files the diff may be excluded with too_large or collapsed set", func(ctx context.Context, params listMergeRequestDiffsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.ListMergeRequestDiffs(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestDiscussionsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestDiscussions", "Lists all discussion items (threads) for a merge request, including diff (inline) and overview threads", func(ctx context.Context, params listMergeRequestDiscussionsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Discussions.ListMergeRequestDiscussions(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestNotesTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestNotes", "Lists all notes (comments) for a merge request", func(ctx context.Context, params listMergeRequestNotesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Notes.ListMergeRequestNotes(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestPipelinesTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestPipelines", "Lists all CI pipelines for a merge request", func(ctx context.Context, params listMergeRequestPipelinesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.ListMergeRequestPipelines(pid, params.MergeRequest, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listPipelineJobsTool, err := toolrunner.NewBetaToolFromJSONSchema("listPipelineJobs", "Lists all jobs for a CI pipeline. Results are paginated (20 per page by default); retried jobs are not returned by default", func(ctx context.Context, params listPipelineJobsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Jobs.ListPipelineJobs(pid, params.PipelineID, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listTreeTool, err := toolrunner.NewBetaToolFromJSONSchema("listTree", "Lists all repository files and directories (like git ls-tree). Use path for a subdirectory and recursive=true for nested entries. ref selects a branch or tag; defaults to the project's default branch. 20 entries per page by default", func(ctx context.Context, params listTreeParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Repositories.ListTree(pid, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	resolveDiscussionTool, err := toolrunner.NewBetaToolFromJSONSchema("resolveDiscussion", "Resolves (or reopens) a discussion thread on a merge request by its ID. Use to mark a previously reported issue as fixed. Requires Developer or higher role, or being the author of the change", func(ctx context.Context, params resolveMergeRequestDiscussionParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		resolved := true
		result, _, err := client.Discussions.ResolveMergeRequestDiscussion(pid, params.MergeRequest, params.DiscussionID, &gitlab.ResolveMergeRequestDiscussionOptions{Resolved: &resolved}, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	searchBlobsTool, err := toolrunner.NewBetaToolFromJSONSchema("searchBlobs", "Searches code (file contents) in the project. By default searches the project's default branch — pass ref (branch or tag) to search another version, e.g. the MR's source branch. Supports filename:, path:, extension: filters and * wildcards", func(ctx context.Context, params searchParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Search.BlobsByProject(pid, params.Query, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	unapproveMergeRequestTool, err := toolrunner.NewBetaToolFromJSONSchema("unapproveMergeRequest", "Removes the approval of the merge request", func(ctx context.Context, params unapproveMergeRequestParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		_, err := client.MergeRequestApprovals.UnapproveMergeRequest(pid, params.MergeRequest, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(map[string]string{"status": "unapproved"})
	})
	if err != nil {
		return nil, err
	}

	return []anthropic.BetaTool{
		approveMergeRequestTool,
		createDiscussionTool,
		createMergeRequestNoteTool,
		getFileTool,
		getMergeRequestTool,
		getMergeRequestCommitsTool,
		listMergeRequestDiffsTool,
		listMergeRequestDiscussionsTool,
		listMergeRequestNotesTool,
		listMergeRequestPipelinesTool,
		listPipelineJobsTool,
		listTreeTool,
		resolveDiscussionTool,
		searchBlobsTool,
		unapproveMergeRequestTool,
	}, nil
}
