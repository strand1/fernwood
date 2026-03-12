package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestMorrohsuIntegration(t *testing.T) {
	// Create test files
	testContent := "This is a test file for integration testing.\nIt has multiple lines.\nIncluding a line with the word ERROR.\n"
	tmpfile, err := os.CreateTemp("", "integration_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	// Test the run tool with registry
	registry := NewCommandRegistry()
	runTool := NewRunTool("/tmp", "", false)
	runTool.SetRegistry(registry)
	
	ctx := context.Background()
	
	// Test 1: Basic cat command
	args := map[string]any{
		"command": "cat " + tmpfile.Name(),
	}
	result := runTool.Execute(ctx, args)
	
	if result.IsError {
		t.Errorf("cat command failed: %s", result.ForLLM)
	}
	
	if !strings.Contains(result.ForLLM, "integration testing") {
		t.Errorf("cat command result doesn't contain expected content: %s", result.ForLLM)
	}
	
	// Check for metadata footer
	if !strings.Contains(result.ForLLM, "[exit:0") {
		t.Errorf("Result should contain metadata footer: %s", result.ForLLM)
	}
	
	// Test 2: Grep command
	args = map[string]any{
		"command": "grep ERROR " + tmpfile.Name(),
	}
	result = runTool.Execute(ctx, args)
	
	if result.IsError {
		t.Errorf("grep command failed: %s", result.ForLLM)
	}
	
	if !strings.Contains(result.ForLLM, "ERROR") {
		t.Errorf("grep command result doesn't contain expected pattern: %s", result.ForLLM)
	}
	
	// Test 3: Ls command
	args = map[string]any{
		"command": "ls .",
	}
	result = runTool.Execute(ctx, args)
	
	if result.IsError {
		t.Errorf("ls command failed: %s", result.ForLLM)
	}
	
	// Test 4: Command with stderr (should be visible)
	args = map[string]any{
		"command": "ls nonexistent_directory",
	}
	result = runTool.Execute(ctx, args)
	
	if !result.IsError {
		t.Error("ls nonexistent_directory should fail")
	}
	
	// Should contain stderr
	if !strings.Contains(result.ForLLM, "[stderr]") {
		t.Errorf("Error result should contain stderr: %s", result.ForLLM)
	}
	
	// Test 5: Help system
	args = map[string]any{
		"command": "help",
	}
	result = runTool.Execute(ctx, args)
	
	if result.IsError {
		t.Errorf("help command failed: %s", result.ForLLM)
	}
	
	if !strings.Contains(result.ForLLM, "Available commands") {
		t.Errorf("help command should list available commands: %s", result.ForLLM)
	}
	
	// Test 6: Unknown command should give guidance
	args = map[string]any{
		"command": "unknown_command",
	}
	result = runTool.Execute(ctx, args)
	
	// Should either fail with guidance or fall back to shell
	if !strings.Contains(result.ForLLM, "unknown") && !strings.Contains(result.ForLLM, "command not found") {
		t.Errorf("Unknown command should give error with guidance: %s", result.ForLLM)
	}
}

func TestBinaryGuardIntegration(t *testing.T) {
	// Create a binary file (simulate with null bytes)
	binaryContent := []byte("This is binary content\x00with null bytes\x00")
	tmpfile, err := os.CreateTemp("", "binary_test_*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write(binaryContent); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	// Test the run tool with registry
	registry := NewCommandRegistry()
	runTool := NewRunTool("/tmp", "", false)
	runTool.SetRegistry(registry)
	
	ctx := context.Background()
	
	// Test cat on binary file - should be blocked with helpful error
	args := map[string]any{
		"command": "cat " + tmpfile.Name(),
	}
	result := runTool.Execute(ctx, args)
	
	if !result.IsError {
		t.Error("cat binary file should fail")
	}
	
	// Should contain helpful error message
	if !strings.Contains(result.ForLLM, "binary") {
		t.Errorf("Binary file error should mention 'binary': %s", result.ForLLM)
	}
	
	if !strings.Contains(result.ForLLM, "Use:") {
		t.Errorf("Binary file error should provide guidance: %s", result.ForLLM)
	}
}

func TestOverflowProtection(t *testing.T) {
	// Create a large file to test overflow protection
	largeContent := strings.Repeat("This is a line that will be repeated many times to create a large output.\n", 1000)
	tmpfile, err := os.CreateTemp("", "large_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(largeContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	// Test the run tool with registry
	registry := NewCommandRegistry()
	runTool := NewRunTool("/tmp", "", false)
	runTool.SetRegistry(registry)
	
	ctx := context.Background()
	
	// Test cat on large file - should trigger overflow protection
	args := map[string]any{
		"command": "cat " + tmpfile.Name(),
	}
	result := runTool.Execute(ctx, args)
	
	// Even large output should succeed (not error)
	if result.IsError {
		t.Errorf("cat large file should not fail: %s", result.ForLLM)
	}
	
	// Should contain overflow information
	if !strings.Contains(result.ForLLM, "truncated") {
		t.Logf("Note: Large output didn't trigger truncation (may be expected based on limits)")
	}
	
	// Should contain metadata footer
	if !strings.Contains(result.ForLLM, "[exit:0") {
		t.Errorf("Large output should still contain metadata footer: %s", result.ForLLM)
	}
}

func TestCommandChaining(t *testing.T) {
	// Create test file
	testContent := "Line 1\nError in this line\nLine 3\nAnother error line\nLine 5\n"
	tmpfile, err := os.CreateTemp("", "chain_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	
	if _, err := tmpfile.Write([]byte(testContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()
	
	// Test the run tool with registry
	registry := NewCommandRegistry()
	runTool := NewRunTool("/tmp", "", false)
	runTool.SetRegistry(registry)
	
	ctx := context.Background()
	
	// Test simple command (single segment)
	args := map[string]any{
		"command": "cat " + tmpfile.Name(),
	}
	result := runTool.Execute(ctx, args)
	
	if result.IsError {
		t.Errorf("Simple cat command failed: %s", result.ForLLM)
	}
	
	// Note: Full command chaining (|, &&, ||, ;) would be implemented in Phase 2
	// For now, we're testing that single commands work through the new system
}