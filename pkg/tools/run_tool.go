// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RunTool provides unified command execution via a single tool interface.
// Supports Unix-style command chaining (|, &&, ||, ;) and semantic commands.
type RunTool struct {
	registry          *CommandRegistry
	workspace         string
	sessionStorage    string
	restrictToWorkspace bool
}

// NewRunTool creates a new run tool with the given workspace and restrictions.
func NewRunTool(workspace, sessionStorage string, restrict bool) *RunTool {
	return &RunTool{
		registry:          NewCommandRegistryFull(workspace, sessionStorage, restrict),
		workspace:         workspace,
		sessionStorage:    sessionStorage,
		restrictToWorkspace: restrict,
	}
}

// SetRegistry allows overriding the default registry (useful for testing or custom commands).
func (t *RunTool) SetRegistry(r *CommandRegistry) {
	t.registry = r
}

// Name returns the tool name for LLM tool definitions.
func (t *RunTool) Name() string {
	return "run"
}

// Description returns the tool description for LLM tool definitions.
func (t *RunTool) Description() string {
	return `Execute Unix-style commands. Supports chaining: cmd1 | cmd2 && cmd3.

Available commands:
  File I/O: ls, cat, write, stat, rm, cp, mv, mkdir
  Text: grep, head, tail, wc
  Memory: memory store/facts/forget/search, memory record/query
  Topics: topic list/info/runs/run/rename/search
  Utils: echo, time, help

Use 'help' for full command list. Use 'run(command="help")' from LLM.`
}

// Parameters returns the JSON schema for LLM tool definitions.
func (t *RunTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": "Command to execute (e.g., 'ls -la | grep .go')",
			},
			"stdin": map[string]any{
				"type":        "string",
				"description": "Standard input for the command (optional)",
			},
		},
		"required": []string{"command"},
	}
}

// Execute runs the command and returns the result.
func (t *RunTool) Execute(ctx context.Context, args map[string]any) *ToolResult {
	start := time.Now()

	// Extract arguments
	command, ok := args["command"].(string)
	if !ok || strings.TrimSpace(command) == "" {
		return &ToolResult{
			ForLLM:  "command is required and must be a non-empty string",
			IsError: true,
		}
	}

	stdin, _ := args["stdin"].(string)

	// Execute command via registry
	output := t.registry.Exec(command, stdin)
	duration := time.Since(start)

	// Apply overflow protection
	output = applyOverflowProtection(output, command)

	// Add metadata footer (using formatDuration from output.go)
	footer := fmt.Sprintf("\n[exit:0 | %s]", formatDuration(duration))
	if strings.Contains(output, "[error]") {
		footer = fmt.Sprintf("\n[exit:1 | %s]", formatDuration(duration))
	}

	return &ToolResult{
		ForLLM:  output + footer,
		IsError: strings.Contains(output, "[error]"),
	}
}

// applyOverflowProtection truncates output if it exceeds limits and writes to temp file.
// This integrates with Morrohsu overflow protection.
func applyOverflowProtection(output, command string) string {
	result := CheckOverflow(output)
	if result.OverflowOccurred {
		return FormatOverflowMessage(result)
	}
	return result.TruncatedOutput
}
