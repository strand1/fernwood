// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"strings"
)

// execChain executes a parsed command chain with proper operator semantics.
// segments: parsed command segments from ParseChain
// initialStdin: initial standard input (for first command in chain)
// Returns: final output string
func execChain(registry *CommandRegistry, segments []Segment, initialStdin string) string {
	if len(segments) == 0 {
		return "[error] empty command chain"
	}

	var collected []string // accumulated outputs for &&, ||, and ;
	var lastOutput string
	var lastErr bool
	pipeInput := initialStdin

	for i, seg := range segments {
		if i > 0 {
			prevOp := segments[i-1].Operator
			// && semantics: skip if previous failed
			if prevOp == OpAnd && lastErr {
				continue
			}
			// || semantics: skip if previous succeeded
			if prevOp == OpOr && !lastErr {
				continue
			}
		}

		// Determine stdin for this segment
		segStdin := ""
		if i == 0 {
			segStdin = pipeInput
		} else if segments[i-1].Operator == OpPipe {
			segStdin = lastOutput
		}

		// Execute single command
		lastOutput, lastErr = execSingle(registry, seg.Command, segStdin)

		// Pipe: output flows to next command's stdin, don't collect yet
		// &&, ||, or ;: collect output (like shell concatenates stdout)
		if i < len(segments)-1 && seg.Operator == OpPipe {
			// Piping — lastOutput will be next command's stdin
			continue
		}

		// Collect output for non-pipe operators or last segment
		if lastOutput != "" {
			collected = append(collected, lastOutput)
		}
	}

	return strings.Join(collected, "\n")
}

// execSingle executes a single command (no chaining operators).
// command: the command string (may include arguments)
// stdin: standard input for the command
// Returns: output string and error flag (true if command failed)
func execSingle(registry *CommandRegistry, command, stdin string) (string, bool) {
	// Tokenize the command first
	parts := tokenizeCommand(command)
	if len(parts) == 0 {
		return "[error] empty command", true
	}

	// Try progressively longer prefixes to find multi-word commands
	// e.g., "memory search \"2080ti\"" -> try "memory", then "memory search"
	var handler CommandHandler
	var ok bool
	var cmdName string
	var args []string

	for i := len(parts); i >= 1; i-- {
		cmdName = strings.Join(parts[:i], " ")
		args = parts[i:]
		handler, ok = registry.GetHandler(cmdName)
		if ok {
			break
		}
	}

	if !ok {
		// No match found - auto-fallback to shell execution
		// This allows the agent to use any shell command naturally
		// Reconstruct the full command string from parts
		fullCmd := strings.Join(parts, " ")
		handler, ok := registry.GetHandler("sh")
		if ok {
			// Execute via shell
			out, err := handler([]string{fullCmd}, stdin)
			if err != nil {
				return formatCommandError(fullCmd, err), true
			}
			return out, false
		}
		// No shell handler available - return error
		available := registry.List()
		return formatUnknownCommandError(parts[0], available), true
	}

	// Execute the command
	out, err := handler(args, stdin)
	if err != nil {
		return formatCommandError(cmdName, err), true
	}

	return out, false
}

// tokenizeCommand splits a command string into parts, respecting quotes.
// Similar to shell tokenization but simpler (no variable expansion, etc.)
func tokenizeCommand(input string) []string {
	var tokens []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	escaped := false

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		switch r {
		case '\\':
			if inSingleQuote {
				current.WriteRune(r)
			} else {
				escaped = true
			}
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
				current.WriteRune(r)
			} else {
				current.WriteRune(r)
			}
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
				current.WriteRune(r)
			} else {
				current.WriteRune(r)
			}
		case ' ', '\t':
			if !inSingleQuote && !inDoubleQuote {
				if current.Len() > 0 {
					tokens = append(tokens, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	// Add the last token
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// formatUnknownCommandError creates a helpful error message for unknown commands.
func formatUnknownCommandError(name string, available []string) string {
	if len(available) == 0 {
		return "[error] unknown command: " + name
	}

	// Show up to 10 available commands
	maxShow := 10
	if len(available) > maxShow {
		available = available[:maxShow]
	}

	return "[error] unknown command: " + name + "\nAvailable: " + strings.Join(available, ", ")
}

// formatCommandError formats a command execution error.
func formatCommandError(name string, err error) string {
	return "[error] " + name + ": " + err.Error()
}
