package commands

import "context"

func clearCommand() Definition {
	return Definition{
		Name:        "clear",
		Description: "Clear the chat history and refresh mulch summaries",
		Usage:       "/clear",
		Handler: func(_ context.Context, req Request, rt *Runtime) error {
			if rt == nil || rt.ClearHistory == nil {
				return req.Reply(unavailableMsg)
			}
			// ClearHistory handles all messaging itself (sends progress updates).
			// Just invoke it and return nil on success; errors are already messaged.
			return rt.ClearHistory()
		},
	}
}
