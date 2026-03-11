package commands

import "context"

func clearCommand() Definition {
	return Definition{
		Name:        "clear",
		Description: "Clear the chat history",
		Usage:       "/clear",
		Handler: func(_ context.Context, req Request, rt *Runtime) error {
			if rt == nil || rt.ClearHistory == nil {
				return req.Reply(unavailableMsg)
			}
			if err := rt.ClearHistory(); err != nil {
				return req.Reply("Failed to clear context: " + err.Error())
			}
			return req.Reply("Context cleared.")
		},
	}
}
