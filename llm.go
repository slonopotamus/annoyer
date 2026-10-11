package main

import (
	"context"
	"log"
	"math"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

func runLLM(ctx context.Context, model string, prompt string, tools []anthropic.BetaTool) error {
	aiClient := anthropic.NewClient()

	runner := aiClient.Beta.Messages.NewToolRunnerStreaming(tools, anthropic.BetaToolRunnerParams{
		MaxTokens: math.MaxInt64,
		Model:     model,
		Messages: []anthropic.BetaMessageParam{
			anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(prompt)),
		},
	})

	for events, err := range runner.AllStreaming(ctx) {
		if err != nil {
			return err
		}

		var sb strings.Builder
		for event, err := range events {
			if err != nil {
				return err
			}
			switch ev := event.AsAny().(type) {
			case anthropic.BetaRawContentBlockDeltaEvent:
				if text := ev.Delta.AsTextDelta(); text.Text != "" {
					sb.WriteString(text.Text)
				}
			}
		}

		if sb.Len() > 0 {
			log.Print(sb.String())
		}
	}

	return nil
}
