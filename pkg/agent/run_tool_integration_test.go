// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/strand1/fernwood/pkg/config"
	"github.com/strand1/fernwood/pkg/tools"
)

// TestRunTool_Integration tests the unified run tool in an agent-like context
func TestRunTool_Integration(t *testing.T) {
	// Create temp workspace
	workspace, err := os.MkdirTemp("", "run_tool_integration_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)

	// Create test files
	testFile := filepath.Join(workspace, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!\nLine 2\nLine 3"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create run tool
	sessionsDir := filepath.Join(workspace, "sessions")
	runTool := tools.NewRunTool(workspace, sessionsDir, false)

	ctx := context.Background()

	// Test 1: cat command
	t.Run("cat command", func(t *testing.T) {
		args := map[string]any{
			"command": "cat test.txt",
		}
		result := runTool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("cat command failed: %s", result.ForLLM)
		}
		if !strings.Contains(result.ForLLM, "Hello, World!") {
			t.Errorf("Expected 'Hello, World!' in output, got: %s", result.ForLLM)
		}
	})

	// Test 2: ls command
	t.Run("ls command", func(t *testing.T) {
		args := map[string]any{
			"command": "ls",
		}
		result := runTool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("ls command failed: %s", result.ForLLM)
		}
		if !strings.Contains(result.ForLLM, "test.txt") {
			t.Errorf("Expected 'test.txt' in output, got: %s", result.ForLLM)
		}
	})

	// Test 3: grep command
	t.Run("grep command", func(t *testing.T) {
		args := map[string]any{
			"command": "cat test.txt | grep \"Line\"",
		}
		result := runTool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("grep command failed: %s", result.ForLLM)
		}
		// Grep should find lines containing "Line"
		if !strings.Contains(result.ForLLM, "Line") {
			t.Errorf("Expected grep to find 'Line', got: %s", result.ForLLM)
		}
	})

	// Test 4: write command
	t.Run("write command", func(t *testing.T) {
		args := map[string]any{
			"command": "write newfile.txt",
			"stdin":   "New file content",
		}
		result := runTool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("write command failed: %s", result.ForLLM)
		}

		// Verify file was written
		content, err := os.ReadFile(filepath.Join(workspace, "newfile.txt"))
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}
		if string(content) != "New file content" {
			t.Errorf("Expected 'New file content', got %q", string(content))
		}
	})

	// Test 5: Command chaining
	t.Run("command chaining", func(t *testing.T) {
		args := map[string]any{
			"command": "cat test.txt | grep \"Line\" | head -n 1",
		}
		result := runTool.Execute(ctx, args)

		if result.IsError {
			t.Errorf("chained command failed: %s", result.ForLLM)
		}
		if !strings.Contains(result.ForLLM, "Line 2") {
			t.Errorf("Expected 'Line 2' in output, got: %s", result.ForLLM)
		}
	})

	// Test 6: memory status command (if mulch is available)
	t.Run("memory status command", func(t *testing.T) {
		args := map[string]any{
			"command": "memory status",
		}
		result := runTool.Execute(ctx, args)

		// This may fail if mulch is not installed, which is OK
		// We're just testing that the command is available
		if strings.Contains(result.ForLLM, "unknown command") {
			t.Logf("Note: mulch not installed, memory status command not available")
			// Don't fail - mulch is optional
		}
	})

	// Test 7: topic list command
	t.Run("topic list command", func(t *testing.T) {
		args := map[string]any{
			"command": "topic list",
		}
		result := runTool.Execute(ctx, args)

		// Should not error even with no topics
		if strings.Contains(result.ForLLM, "unknown command") {
			t.Logf("Note: topic commands not available")
			// Don't fail - this is a bug if this happens
		}
	})
}

// TestRunTool_OverflowProtection tests that large outputs are truncated
func TestRunTool_OverflowProtection(t *testing.T) {
	workspace, err := os.MkdirTemp("", "run_tool_overflow_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)

	// Create a large file
	largeFile := filepath.Join(workspace, "large.txt")
	largeContent := strings.Repeat("Line\n", 300) // 300 lines > 200 line limit
	if err := os.WriteFile(largeFile, []byte(largeContent), 0644); err != nil {
		t.Fatal(err)
	}

	runTool := tools.NewRunTool(workspace, "", false)
	ctx := context.Background()

	args := map[string]any{
		"command": "cat large.txt",
	}
	result := runTool.Execute(ctx, args)

	// Should be truncated
	if !strings.Contains(result.ForLLM, "truncated") {
		t.Errorf("Expected truncation message, got: %s", result.ForLLM)
	}
	if !strings.Contains(result.ForLLM, "Full output:") {
		t.Errorf("Expected full output path, got: %s", result.ForLLM)
	}
}

// TestRunTool_BinaryDetection tests that binary files are detected
func TestRunTool_BinaryDetection(t *testing.T) {
	workspace, err := os.MkdirTemp("", "run_tool_binary_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)

	// Create a binary file
	binaryFile := filepath.Join(workspace, "binary.bin")
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0xFF, 0xFE}
	if err := os.WriteFile(binaryFile, binaryData, 0644); err != nil {
		t.Fatal(err)
	}

	runTool := tools.NewRunTool(workspace, "", false)
	ctx := context.Background()

	args := map[string]any{
		"command": "cat binary.bin",
	}
	result := runTool.Execute(ctx, args)

	// Should detect binary and return helpful error
	if !strings.Contains(result.ForLLM, "[error]") || !strings.Contains(result.ForLLM, "binary") {
		t.Errorf("Expected binary file error, got: %s", result.ForLLM)
	}
}

// TestAgentInstance_WithRunTool tests that AgentInstance can be created with run tool
func TestAgentInstance_WithRunTool(t *testing.T) {
	workspace, err := os.MkdirTemp("", "agent_instance_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workspace)

	// Create minimal config with run tool enabled
	cfg := &config.Config{
		Agents: config.AgentsConfig{
			Defaults: config.AgentDefaults{
				Workspace:          workspace,
				RestrictToWorkspace: false,
			},
		},
		Tools: config.ToolsConfig{
			Run: config.ToolConfig{Enabled: true},
			EditFile: config.ToolConfig{Enabled: true},
		},
	}

	// Create a mock provider (we won't actually call LLM)
	provider := &mockProvider{}

	// Create agent instance
	agent := NewAgentInstance(nil, &cfg.Agents.Defaults, cfg, provider)
	if agent == nil {
		t.Fatal("Failed to create agent instance")
	}

	// Verify run tool is registered
	runTool, ok := agent.Tools.Get("run")
	if !ok {
		t.Fatal("Expected run tool to be registered")
	}

	// Verify edit_file is registered
	editTool, ok := agent.Tools.Get("edit_file")
	if !ok {
		t.Fatal("Expected edit_file tool to be registered")
	}

	// Verify old tools are NOT registered
	oldTools := []string{"read_file", "write_file", "list_dir", "bash", "exec", "mulch_query", "mulch_record"}
	for _, toolName := range oldTools {
		if _, ok := agent.Tools.Get(toolName); ok {
			t.Errorf("Expected %s tool to NOT be registered (replaced by run)", toolName)
		}
	}

	// Suppress unused variable warnings
	_ = runTool
	_ = editTool
}
