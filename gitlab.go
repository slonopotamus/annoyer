package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"gitlab.com/gitlab-org/api/client-go/v3"
)

type approveMergeRequestParams struct {
	MergeRequest int64                             `json:"merge_request_id"`
	Options      gitlab.ApproveMergeRequestOptions `json:"options"`
}

type compareParams struct {
	Options gitlab.CompareOptions `json:"options"`
}

type createMergeRequestDiscussionParams struct {
	MergeRequest int64                                      `json:"merge_request_id"`
	Options      gitlab.CreateMergeRequestDiscussionOptions `json:"options"`
}

type createMergeRequestNoteParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.CreateMergeRequestNoteOptions `json:"options"`
}

type getBranchParams struct {
	Branch string `json:"branch"`
}

type getCommitParams struct {
	SHA     string                  `json:"sha"`
	Options gitlab.GetCommitOptions `json:"options"`
}

type getCommitDiffParams struct {
	SHA     string                      `json:"sha"`
	Options gitlab.GetCommitDiffOptions `json:"options"`
}

type getFileParams struct {
	Path    string                   `json:"path"`
	Options gitlab.GetRawFileOptions `json:"options"`
}

type getFileBlameParams struct {
	Path    string                     `json:"path"`
	Options gitlab.GetFileBlameOptions `json:"options"`
}

type getJobParams struct {
	JobID int64 `json:"job_id"`
}

type getJobLogParams struct {
	JobID int64 `json:"job_id"`
}

type getMergeRequestParams struct {
	MergeRequest int64                          `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestsOptions `json:"options"`
}

type getMergeRequestCommitsParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestCommitsOptions `json:"options"`
}

type getPipelineTestReportParams struct {
	PipelineID int64 `json:"pipeline_id"`
}

type listBranchesParams struct {
	Options gitlab.ListBranchesOptions `json:"options"`
}

type listCommitsParams struct {
	Options gitlab.ListCommitsOptions `json:"options"`
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

type listMergeRequestsParams struct {
	Options gitlab.ListProjectMergeRequestsOptions `json:"options"`
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

const maxJobLogBytes = 100 * 1024

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

	compareTool, err := toolrunner.NewBetaToolFromJSONSchema("compare", "Compares two refs (branch, tag, or commit SHA) and returns the diff between them; set from and to in options. Use to diff the MR head against the current target branch head, or any two SHAs", func(ctx context.Context, params compareParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Repositories.Compare(pid, &params.Options, gitlab.WithContext(ctx))
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

	getBranchTool, err := toolrunner.NewBetaToolFromJSONSchema("getBranch", "Retrieves a single branch by name, including the SHA of its HEAD commit. Use to find the latest commit of a branch, e.g. the MR's source branch", func(ctx context.Context, params getBranchParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Branches.GetBranch(pid, params.Branch, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getCommitTool, err := toolrunner.NewBetaToolFromJSONSchema("getCommit", "Retrieves a single commit by SHA (full or short), including its message, author, and parent commits", func(ctx context.Context, params getCommitParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Commits.GetCommit(pid, params.SHA, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getCommitDiffTool, err := toolrunner.NewBetaToolFromJSONSchema("getCommitDiff", "Retrieves the file diff of a single commit by SHA. Use to review a multi-commit MR commit by commit", func(ctx context.Context, params getCommitDiffParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Commits.GetCommitDiff(pid, params.SHA, &params.Options, gitlab.WithContext(ctx))
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

	getFileBlameTool, err := toolrunner.NewBetaToolFromJSONSchema("getFileBlame", "Retrieves the blame of a file: for each line, the commit that last modified it, with author and date. Use to tell whether a problematic line was introduced by this MR or is pre-existing. ref selects a branch or tag; defaults to the project's default branch. Use range_start and range_end to limit to a line range", func(ctx context.Context, params getFileBlameParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.RepositoryFiles.GetFileBlame(pid, params.Path, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getJobTool, err := toolrunner.NewBetaToolFromJSONSchema("getJob", "Retrieves a single CI job by its job ID (from listPipelineJobs): status, stage, name, and duration", func(ctx context.Context, params getJobParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Jobs.GetJob(pid, params.JobID, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	getJobLogTool, err := toolrunner.NewBetaToolFromJSONSchema("getJobLog", "Retrieves the log (trace) of a CI job by its job ID (from listPipelineJobs). Use to diagnose why a job failed. Logs can be large, so the result is truncated to its tail when long — the failure summary is at the end of the log", func(ctx context.Context, params getJobLogParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		reader, _, err := client.Jobs.GetTraceFile(pid, params.JobID, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		jobLog, err := io.ReadAll(reader)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		if len(jobLog) > maxJobLogBytes {
			jobLog = append([]byte(fmt.Sprintf("[log truncated, showing the last %d of %d bytes]\n", maxJobLogBytes, len(jobLog))), jobLog[len(jobLog)-maxJobLogBytes:]...)
		}

		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(jobLog)},
		}, nil
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

	getPipelineTestReportTool, err := toolrunner.NewBetaToolFromJSONSchema("getPipelineTestReport", "Retrieves the unit test results of a CI pipeline (pipeline ID from listMergeRequestPipelines): pass/fail/skip totals and the list of failed and skipped tests with durations", func(ctx context.Context, params getPipelineTestReportParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Pipelines.GetPipelineTestReport(pid, params.PipelineID, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listBranchesTool, err := toolrunner.NewBetaToolFromJSONSchema("listBranches", "Lists the repository's branches. Use name_regex to filter by pattern. 20 branches per page by default", func(ctx context.Context, params listBranchesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Branches.ListBranches(pid, &params.Options, gitlab.WithContext(ctx))
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return marshalToBetaToolResult(result)
	})
	if err != nil {
		return nil, err
	}

	listCommitsTool, err := toolrunner.NewBetaToolFromJSONSchema("listCommits", "Lists commits of the repository. ref_name selects a branch or tag; defaults to the project's default branch. Use path to narrow to commits touching a specific file, and since/until to bound the time range. 20 commits per page by default", func(ctx context.Context, params listCommitsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.Commits.ListCommits(pid, &params.Options, gitlab.WithContext(ctx))
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

	listMergeRequestsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequests", "Lists merge requests of the project: ID, title, state, source/target branches, and author. Use state to filter (opened, closed, merged, all; opened by default). Returns basic merge request objects; use getMergeRequest for full details", func(ctx context.Context, params listMergeRequestsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, _, err := client.MergeRequests.ListProjectMergeRequests(pid, &params.Options, gitlab.WithContext(ctx))
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
		compareTool,
		createDiscussionTool,
		createMergeRequestNoteTool,
		getBranchTool,
		getCommitTool,
		getCommitDiffTool,
		getFileTool,
		getFileBlameTool,
		getJobTool,
		getJobLogTool,
		getMergeRequestTool,
		getMergeRequestCommitsTool,
		getPipelineTestReportTool,
		listBranchesTool,
		listCommitsTool,
		listMergeRequestDiffsTool,
		listMergeRequestDiscussionsTool,
		listMergeRequestNotesTool,
		listMergeRequestPipelinesTool,
		listMergeRequestsTool,
		listPipelineJobsTool,
		listTreeTool,
		resolveDiscussionTool,
		searchBlobsTool,
		unapproveMergeRequestTool,
	}, nil
}
