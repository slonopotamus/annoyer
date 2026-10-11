package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
	"gitlab.com/gitlab-org/api/client-go/v3"
)

type getFileParams struct {
	Path    string                   `json:"path"`
	Options gitlab.GetRawFileOptions `json:"options"`
}

func getGitLabFile(ctx context.Context, git *gitlab.Client, pid any, params getFileParams) ([]byte, error) {
	result, _, err := git.RepositoryFiles.GetRawFile(pid, params.Path, &params.Options, gitlab.WithContext(ctx))
	return result, err
}

type searchParams struct {
	Query   string               `json:"query"`
	Options gitlab.SearchOptions `json:"options"`
}

func searchBlobs(ctx context.Context, git *gitlab.Client, pid any, params searchParams) ([]*gitlab.Blob, error) {
	result, _, err := git.Search.BlobsByProject(pid, params.Query, &params.Options, gitlab.WithContext(ctx))
	return result, err
}

type getMergeRequestChangesParams struct {
	MergeRequest int64                                `json:"merge_request_id"`
	Options      gitlab.GetMergeRequestChangesOptions `json:"options"`
}

func getMergeRequestChanges(ctx context.Context, git *gitlab.Client, pid any, params getMergeRequestChangesParams) (*gitlab.MergeRequest, error) {
	result, _, err := git.MergeRequests.GetMergeRequestChanges(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
	return result, err
}

type listMergeRequestDiffsParams struct {
	MergeRequest int64                               `json:"merge_request_id"`
	Options      gitlab.ListMergeRequestDiffsOptions `json:"options"`
}

func listMergeRequestDiffs(ctx context.Context, git *gitlab.Client, pid any, params listMergeRequestDiffsParams) ([]*gitlab.MergeRequestDiff, error) {
	result, _, err := git.MergeRequests.ListMergeRequestDiffs(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
	return result, err
}

type createMergeRequestDiscussionParams struct {
	MergeRequest int64                                      `json:"merge_request_id"`
	Options      gitlab.CreateMergeRequestDiscussionOptions `json:"options"`
}

func createDiscussion(ctx context.Context, git *gitlab.Client, pid any, params createMergeRequestDiscussionParams) error {
	_, _, err := git.Discussions.CreateMergeRequestDiscussion(pid, params.MergeRequest, &params.Options, gitlab.WithContext(ctx))
	return err
}

func newGitLabTools(client *gitlab.Client, pid any) ([]anthropic.BetaTool, error) {
	createDiscussionTool, err := toolrunner.NewBetaToolFromJSONSchema("createDiscussion", "Creates a merge request discussion", func(ctx context.Context, params createMergeRequestDiscussionParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		err := createDiscussion(ctx, client, pid, params)
		return anthropic.BetaToolResultBlockParamContentUnion{}, err
	})
	if err != nil {
		return nil, err
	}

	getFileTool, err := toolrunner.NewBetaToolFromJSONSchema("getFile", "Gets file contents by path", func(ctx context.Context, params getFileParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, err := getGitLabFile(ctx, client, pid, params)
		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(result)},
		}, err
	})
	if err != nil {
		return nil, err
	}

	getMergeRequestChangesTool, err := toolrunner.NewBetaToolFromJSONSchema("getMergeRequestChanges", "Gets details of a merge request", func(ctx context.Context, params getMergeRequestChangesParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		result, err := getMergeRequestChanges(ctx, client, pid, params)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		encoded, err := json.Marshal(result)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: string(encoded)},
		}, err
	})
	if err != nil {
		return nil, err
	}

	listMergeRequestDiffsTool, err := toolrunner.NewBetaToolFromJSONSchema("listMergeRequestDiffs", "List merge request changed files", func(ctx context.Context, params listMergeRequestDiffsParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		diffs, err := listMergeRequestDiffs(ctx, client, pid, params)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		var sb strings.Builder
		for _, diff := range diffs {
			encoded, err := json.Marshal(diff)
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			sb.WriteString("\n")
			sb.WriteString(string(encoded))
		}

		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: sb.String()},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	searchBlobsTool, err := toolrunner.NewBetaToolFromJSONSchema("searchBlobs", "Searches files matching query", func(ctx context.Context, params searchParams) (anthropic.BetaToolResultBlockParamContentUnion, error) {
		blobs, err := searchBlobs(ctx, client, pid, params)
		if err != nil {
			return anthropic.BetaToolResultBlockParamContentUnion{}, err
		}

		var sb strings.Builder
		for _, blob := range blobs {
			encoded, err := json.Marshal(blob)
			if err != nil {
				return anthropic.BetaToolResultBlockParamContentUnion{}, err
			}
			sb.WriteString("\n")
			sb.WriteString(string(encoded))
		}

		return anthropic.BetaToolResultBlockParamContentUnion{
			OfText: &anthropic.BetaTextBlockParam{Text: sb.String()},
		}, nil
	})
	if err != nil {
		return nil, err
	}

	return []anthropic.BetaTool{createDiscussionTool, getFileTool, getMergeRequestChangesTool, listMergeRequestDiffsTool, searchBlobsTool}, nil
}
