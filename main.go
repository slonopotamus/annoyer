package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/anthropics/anthropic-sdk-go"
	gitlab "gitlab.com/gitlab-org/api/client-go/v3"
)

func doMain(ctx context.Context) error {
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
		llmModel = "unsloth/Qwen3.8-27B-GGUF"
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

	prompt := fmt.Sprintf(`Review MR %d for issues.
Only report issues you are confident about.
Compilation issues will be caught by CI.
Create discussions in MR with your findings.
Mark discussions as AI-generated.
Add severity markers with icons (🔴 critical/🟠 high/🟡 medium/🟢 low).`, mergeRequestId)

	return runLLM(ctx, llmModel, prompt, tools)
}

func main() {
	if err := doMain(context.TODO()); err != nil {
		log.Fatal(err)
	}
}
