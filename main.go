package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/tools/agenttoolset"
	"github.com/spf13/cobra"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

const defaultPrompt = `
Produce a concise summary (2-5 sentences) note of the entire merge request if none exists yet: what it does and which areas it touches.
Only report issues you are confident about.
Compilation issues will be caught by CI, no need to report them.
Create messages in MR with your findings.
Use discussions for messages that demand reaction, use notes for messages that don't.
Add "🤖 AI-generated review" header in italic to your messages.
Use markers for issue severity:
* 🔴 [critical]
* 🟠 [high]
* 🟡 [medium]
* 🟢 [low]
Ignore CI status.
Resolve pre-existing AI reviews if they are no longer applicable.
If you find no issues blocking the merge, approve the merge request silently.`

func doMain(ctx context.Context, allowedRoots []string) error {
	gitlabToken := os.Getenv("GITLAB_TOKEN")
	if gitlabToken == "" {
		return fmt.Errorf("GITLAB_TOKEN env variable is not set")
	}

	gitlabUrl := os.Getenv("CI_SERVER_URL")
	if gitlabUrl == "" {
		return fmt.Errorf("CI_SERVER_URL env variable is not set")
	}

	gitlabProjectId := os.Getenv("CI_PROJECT_ID")
	if gitlabProjectId == "" {
		return fmt.Errorf("CI_PROJECT_ID env variable is not set")
	}

	git, err := gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl))
	if err != nil {
		return err
	}

	mergeRequestId, err := strconv.ParseInt(os.Getenv("CI_MERGE_REQUEST_IID"), 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse CI_MERGE_REQUEST_IID env variable: %w", err)
	}

	localTools, err := newGitLabTools(git, gitlabProjectId)
	if err != nil {
		return err
	}

	llmModel := os.Getenv("ANNOYER_MODEL")
	if llmModel == "" {
		llmModel = anthropic.ModelClaudeOpus5
	}

	var tools []anthropic.BetaTool

	/*
		gitlabMcpTools, err := createGitLabMcp(ctx, gitlabUrl, gitlabToken)
		if err != nil {
			return err
		}
		tools = append(tools, gitlabMcpTools...)
	*/

	tools = append(tools, localTools...)

	if len(allowedRoots) > 0 {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		agentEnv := agenttoolset.AgentToolContext{Workdir: wd, AllowedRoots: allowedRoots}
		tools = append(tools,
			agenttoolset.BetaReadTool(&agentEnv),
			agenttoolset.BetaGlobTool(&agentEnv),
			agenttoolset.BetaGrepTool(&agentEnv),
		)
	}

	gitRef := fmt.Sprintf("refs/merge-requests/%d/head", mergeRequestId)
	promptFile, resp, err := git.RepositoryFiles.GetRawFile(gitlabProjectId, "REVIEW.md", &gitlab.GetRawFileOptions{Ref: &gitRef})

	prompt := defaultPrompt
	if err != nil && (resp == nil || resp.StatusCode != http.StatusNotFound) {
		return err
	} else if len(promptFile) > 0 {
		prompt = string(promptFile)
	}

	return runLLM(ctx, llmModel, []string{fmt.Sprintf("You are reviewing merge request %d", mergeRequestId), prompt}, tools)
}

func newAnnoyerCommand() *cobra.Command {
	var allowedRoots []string

	cmd := &cobra.Command{
		Use:           "annoyer",
		Short:         "Review a GitLab merge request with an LLM agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doMain(cmd.Context(), allowedRoots)
		},
	}
	cmd.Flags().StringSliceVar(&allowedRoots, "allowed-roots", nil, "additional directories the file tools may access")
	return cmd
}

func main() {
	if err := newAnnoyerCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
