// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// CommandHandler is a function that executes a command.
// args: command-line arguments (not including the command name itself)
// stdin: standard input (from pipe or explicit stdin argument)
// Returns: output string and error
type CommandHandler func(args []string, stdin string) (string, error)

// CommandRegistry maps command names to handlers with help text and alias support.
type CommandRegistry struct {
	handlers map[string]CommandHandler
	help     map[string]string
	aliases  map[string]string // "fs.ls" -> "ls"
	mu       sync.RWMutex
}

// NewCommandRegistry creates a new command registry with built-in commands.
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{
		handlers: make(map[string]CommandHandler),
		help:     make(map[string]string),
		aliases:  make(map[string]string),
	}
	r.registerBuiltins()
	return r
}

// NewCommandRegistryWithFS creates a new command registry with file system commands.
// workspace: base directory for relative paths
// restrict: if true, restrict all operations to workspace
func NewCommandRegistryWithFS(workspace string, restrict bool) *CommandRegistry {
	r := NewCommandRegistry()
	RegisterFSCommands(r, workspace, restrict)
	return r
}

// NewCommandRegistryFull creates a new command registry with all commands (FS, memory, topic, skills).
// workspace: base directory for relative paths
// restrict: if true, restrict all operations to workspace
// sessionStorage: path to session storage directory
func NewCommandRegistryFull(workspace, sessionStorage string, restrict bool) *CommandRegistry {
	r := NewCommandRegistryWithFS(workspace, restrict)
	RegisterMemoryCommands(r, workspace)
	RegisterTopicCommands(r, sessionStorage)
	RegisterSkillCommands(r, workspace)
	return r
}

// Register adds a command to the registry.
// name: command name (e.g., "ls", "cat", "memory")
// description: help text shown in "help" output
// handler: function to execute the command
func (r *CommandRegistry) Register(name, description string, handler CommandHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[name] = handler
	r.help[name] = description
}

// RegisterAlias creates an alias for an existing command.
// alias: the new name (e.g., "fs.ls")
// target: the existing command name (e.g., "ls")
func (r *CommandRegistry) RegisterAlias(alias, target string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.aliases[alias] = target
}

// GetHandler returns the handler for a command name, resolving aliases.
// Returns nil if command not found.
func (r *CommandRegistry) GetHandler(name string) (CommandHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check for direct match first
	if handler, ok := r.handlers[name]; ok {
		return handler, true
	}

	// Check for alias
	if target, ok := r.aliases[name]; ok {
		if handler, ok := r.handlers[target]; ok {
			return handler, true
		}
	}

	return nil, false
}

// Exec executes a command string with the given stdin.
// Supports command chaining via ParseChain and ExecChain.
// Returns the final output string (may include error messages).
func (r *CommandRegistry) Exec(command, stdin string) string {
	if strings.TrimSpace(command) == "" {
		return "[error] empty command"
	}

	segments := ParseChain(command)
	if len(segments) == 0 {
		return "[error] failed to parse command"
	}

	return execChain(r, segments, stdin)
}

// Help returns the help text map for all commands.
func (r *CommandRegistry) Help() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]string, len(r.help))
	for k, v := range r.help {
		result[k] = v
	}
	return result
}

// HelpText returns formatted help text suitable for LLM consumption.
func (r *CommandRegistry) HelpText() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Get sorted command names for deterministic output
	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Available commands:\n")
	for _, name := range names {
		desc := r.help[name]
		// Truncate long descriptions
		if len(desc) > 200 {
			desc = desc[:197] + "..."
		}
		fmt.Fprintf(&b, "  %s — %s\n", name, desc)
	}

	// Include aliases
	if len(r.aliases) > 0 {
		b.WriteString("\nAliases:\n")
		for alias, target := range r.aliases {
			fmt.Fprintf(&b, "  %s → %s\n", alias, target)
		}
	}

	// Note about shell auto-fallback
	b.WriteString("\nNote: Unknown commands are automatically executed via shell (sh -c).\n")
	b.WriteString("      You can use any shell command: sed, awk, find, git, python3, etc.\n")
	b.WriteString("      Pipes (|), redirects (>), &&, || all work naturally.\n")

	return b.String()
}

// List returns a sorted list of all command names.
func (r *CommandRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.handlers))
	for name := range r.handlers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// registerBuiltins registers built-in utility commands.
func (r *CommandRegistry) registerBuiltins() {
	// echo — Echo back input or arguments
	r.Register("echo", "Echo back the input or arguments", func(args []string, stdin string) (string, error) {
		if stdin != "" {
			return strings.TrimSuffix(stdin, "\n"), nil
		}
		return strings.Join(args, " "), nil
	})

	// time — Return current timestamp
	r.Register("time", "Return the current timestamp", func(args []string, stdin string) (string, error) {
		return fmt.Sprintf("%d", time.Now().Unix()), nil
	})

	// help — List available commands
	r.Register("help", "List available commands or show help for a specific command", func(args []string, stdin string) (string, error) {
		if len(args) > 0 {
			// Show help for specific command
			cmdName := args[0]
			if handler, ok := r.GetHandler(cmdName); ok {
				_ = handler // Just checking if it exists
				if desc, ok := r.Help()[cmdName]; ok {
					return fmt.Sprintf("%s: %s", cmdName, desc), nil
				}
			}
			return fmt.Sprintf("Unknown command: %s", cmdName), nil
		}
		return r.HelpText(), nil
	})
}
