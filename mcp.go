package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
	anthropicmcp "github.com/anthropics/anthropic-sdk-go/mcp"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func createGitLabMcp(ctx context.Context, gitlabUrl string, gitlabToken string) ([]anthropic.BetaTool, error) {
	transport := &mcp.StreamableClientTransport{
		Endpoint: fmt.Sprintf("%s/api/v4/mcp", gitlabUrl),
		HTTPClient: &http.Client{
			Transport: &authRoundTripper{token: gitlabToken, next: http.DefaultTransport},
		},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "annoyer", Version: "v1.0.0"}, nil)
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}

	mcpToolsList, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	return anthropicmcp.NewBetaTools(mcpToolsList.Tools, session)
}
