// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"testing"
)

func TestCommandRegistry_RegisterAndGet(t *testing.T) {
	r := NewCommandRegistry()

	// Register a simple command
	called := false
	r.Register("test", "Test command", func(args []string, stdin string) (string, error) {
		called = true
		return "test output", nil
	})

	// Get handler
	handler, ok := r.GetHandler("test")
	if !ok {
		t.Fatal("Expected to find 'test' command")
	}

	// Execute handler
	output, err := handler([]string{"arg1"}, "stdin")
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	if output != "test output" {
		t.Errorf("Expected 'test output', got %q", output)
	}
	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestCommandRegistry_Alias(t *testing.T) {
	r := NewCommandRegistry()

	// Register command
	r.Register("ls", "List files", func(args []string, stdin string) (string, error) {
		return "file1\nfile2", nil
	})

	// Create alias
	r.RegisterAlias("fs.ls", "ls")

	// Get handler via alias
	handler, ok := r.GetHandler("fs.ls")
	if !ok {
		t.Fatal("Expected to find 'fs.ls' alias")
	}

	output, err := handler(nil, "")
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}
	if output != "file1\nfile2" {
		t.Errorf("Expected 'file1\\nfile2', got %q", output)
	}
}

func TestCommandRegistry_HelpText(t *testing.T) {
	r := NewCommandRegistry()

	help := r.HelpText()
	if help == "" {
		t.Error("Expected non-empty help text")
	}

	// Check for built-in commands
	expected := []string{"echo", "time", "help"}
	for _, cmd := range expected {
		if !contains(help, cmd) {
			t.Errorf("Expected help text to contain %q", cmd)
		}
	}
}

func TestCommandRegistry_List(t *testing.T) {
	r := NewCommandRegistry()

	list := r.List()
	if len(list) == 0 {
		t.Error("Expected non-empty command list")
	}

	// Check for built-in commands
	expected := []string{"echo", "time", "help"}
	for _, cmd := range expected {
		if !containsString(list, cmd) {
			t.Errorf("Expected list to contain %q", cmd)
		}
	}
}

func TestCommandRegistry_Exec_Simple(t *testing.T) {
	r := NewCommandRegistry()

	r.Register("test", "Test", func(args []string, stdin string) (string, error) {
		return "output", nil
	})

	output := r.Exec("test", "")
	if output != "output" {
		t.Errorf("Expected 'output', got %q", output)
	}
}

func TestCommandRegistry_Exec_Empty(t *testing.T) {
	r := NewCommandRegistry()

	output := r.Exec("", "")
	if output != "[error] empty command" {
		t.Errorf("Expected error for empty command, got %q", output)
	}
}

func TestCommandRegistry_Exec_Unknown(t *testing.T) {
	r := NewCommandRegistry()

	output := r.Exec("unknowncommand", "")
	if !contains(output, "[error]") {
		t.Errorf("Expected error for unknown command, got %q", output)
	}
	if !contains(output, "unknown command") {
		t.Errorf("Expected 'unknown command' in error, got %q", output)
	}
}

func TestExecChain_Single(t *testing.T) {
	r := NewCommandRegistry()
	r.Register("cmd", "Test", func(args []string, stdin string) (string, error) {
		return "result", nil
	})

	segments := []Segment{
		{Command: "cmd", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	if output != "result" {
		t.Errorf("Expected 'result', got %q", output)
	}
}

func TestExecChain_Pipe(t *testing.T) {
	r := NewCommandRegistry()

	// First command outputs "hello"
	r.Register("cmd1", "Test 1", func(args []string, stdin string) (string, error) {
		return "hello", nil
	})

	// Second command receives stdin
	r.Register("cmd2", "Test 2", func(args []string, stdin string) (string, error) {
		return "received: " + stdin, nil
	})

	segments := []Segment{
		{Command: "cmd1", Operator: OpPipe},
		{Command: "cmd2", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	if output != "received: hello" {
		t.Errorf("Expected 'received: hello', got %q", output)
	}
}

func TestExecChain_And(t *testing.T) {
	r := NewCommandRegistry()

	r.Register("cmd1", "Test 1", func(args []string, stdin string) (string, error) {
		return "first", nil
	})

	r.Register("cmd2", "Test 2", func(args []string, stdin string) (string, error) {
		return "second", nil
	})

	segments := []Segment{
		{Command: "cmd1", Operator: OpAnd},
		{Command: "cmd2", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	// Both should execute and outputs collected
	if output != "first\nsecond" {
		t.Errorf("Expected 'first\\nsecond', got %q", output)
	}
}

func TestExecChain_And_SkipOnFailure(t *testing.T) {
	r := NewCommandRegistry()

	r.Register("cmd1", "Test 1", func(args []string, stdin string) (string, error) {
		return "", &testError{"failed"}
	})

	r.Register("cmd2", "Test 2", func(args []string, stdin string) (string, error) {
		return "should not reach", nil
	})

	segments := []Segment{
		{Command: "cmd1", Operator: OpAnd},
		{Command: "cmd2", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	if contains(output, "should not reach") {
		t.Errorf("Expected cmd2 to be skipped, got %q", output)
	}
}

func TestExecChain_Or_ExecuteOnFailure(t *testing.T) {
	r := NewCommandRegistry()

	r.Register("cmd1", "Test 1", func(args []string, stdin string) (string, error) {
		return "", &testError{"failed"}
	})

	r.Register("cmd2", "Test 2", func(args []string, stdin string) (string, error) {
		return "fallback", nil
	})

	segments := []Segment{
		{Command: "cmd1", Operator: OpOr},
		{Command: "cmd2", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	// cmd1 fails, so cmd2 executes. Error output from cmd1 is collected, then cmd2 output
	if !contains(output, "fallback") {
		t.Errorf("Expected output to contain 'fallback', got %q", output)
	}
}

func TestExecChain_Seq(t *testing.T) {
	r := NewCommandRegistry()

	r.Register("cmd1", "Test 1", func(args []string, stdin string) (string, error) {
		return "first", nil
	})

	r.Register("cmd2", "Test 2", func(args []string, stdin string) (string, error) {
		return "second", nil
	})

	segments := []Segment{
		{Command: "cmd1", Operator: OpSeq},
		{Command: "cmd2", Operator: OpNone},
	}

	output := execChain(r, segments, "")
	if output != "first\nsecond" {
		t.Errorf("Expected 'first\\nsecond', got %q", output)
	}
}

func TestTokenizeCommand_Simple(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"ls", []string{"ls"}},
		{"ls -la", []string{"ls", "-la"}},
		{"cat file.txt", []string{"cat", "file.txt"}},
		{"echo hello world", []string{"echo", "hello", "world"}},
	}

	for _, tt := range tests {
		result := tokenizeCommand(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenizeCommand(%q): expected %v, got %v", tt.input, tt.expected, result)
			continue
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("tokenizeCommand(%q): expected %v, got %v", tt.input, tt.expected, result)
				break
			}
		}
	}
}

func TestTokenizeCommand_Quotes(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`echo "hello world"`, []string{"echo", `"hello world"`}},
		{`echo 'hello world'`, []string{"echo", `'hello world'`}},
		{`cat "file with spaces.txt"`, []string{"cat", `"file with spaces.txt"`}},
	}

	for _, tt := range tests {
		result := tokenizeCommand(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenizeCommand(%q): expected %d tokens, got %d", tt.input, len(tt.expected), len(result))
			continue
		}
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// testError is a simple error implementation for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
