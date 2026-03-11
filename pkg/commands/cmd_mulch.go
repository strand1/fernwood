package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/strand1/fernwood/pkg/memory"
)

func mulchCommand() Definition {
	return Definition{
		Name:        "mulch",
		Description: "Manage mulch expertise summaries",
		Usage:       "/mulch [subcommand]",
		SubCommands: []SubCommand{
			{
				Name:        "summarize",
				Description: "Summarize stale expertise domains",
				Handler: func(_ context.Context, req Request, rt *Runtime) error {
					if rt == nil {
						return req.Reply("runtime not available")
					}
					if rt.GetMulchManager == nil || rt.GetLLMProvider == nil {
						return req.Reply("mulch not available in this runtime")
					}
					mulchMgr := rt.GetMulchManager()
					llmProvider := rt.GetLLMProvider()
					if mulchMgr == nil || !mulchMgr.Enabled || llmProvider == nil {
						return req.Reply("mulch is not enabled")
					}

					// Get model from GetModelInfo
					model, _ := rt.GetModelInfo()
					if model == "" {
						model = "default"
					}
					summarizer := memory.NewProviderSummarizer(llmProvider, model)

					// Run summarization synchronously
					ctx, cancel := context.WithTimeout(context.Background(), defaultCmdTimeout)
					defer cancel()
					refreshed, skipped, err := mulchMgr.SummarizeDomains(ctx, summarizer)
					if err != nil {
						return req.Reply(fmt.Sprintf("Summarization failed: %v", err))
					}
					msg := fmt.Sprintf("Domain summarization complete: %d refreshed, %d skipped.", len(refreshed), len(skipped))
					return req.Reply(msg)
				},
			},
		},
	}
}

// defaultCmdTimeout is a reasonable timeout for /mulch commands.
const defaultCmdTimeout = 5 * time.Minute
