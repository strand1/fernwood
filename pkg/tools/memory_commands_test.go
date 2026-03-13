// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/strand1/fernwood/pkg/providers"
	"github.com/strand1/fernwood/pkg/session"
)

func TestRegisterMemoryCommands(t *testing.T) {
	registry := NewCommandRegistry()
	RegisterMemoryCommands(registry, "/tmp/test")

	// Test that memory commands are registered
	memoryCommands := []string{
		"memory store",
		"memory record",
		"memory facts",
		"memory search",
		"memory query",
		"memory forget",
		"memory status",
	}

	for _, cmd := range memoryCommands {
		_, ok := registry.GetHandler(cmd)
		if !ok {
			t.Errorf("Expected command '%s' to be registered", cmd)
		}
	}

	// Test that aliases are registered
	aliases := []string{
		"mem.store",
		"mem.record",
		"mem.facts",
		"mem.search",
		"mem.query",
		"mem.forget",
		"mem.status",
	}

	for _, alias := range aliases {
		_, ok := registry.GetHandler(alias)
		if !ok {
			t.Errorf("Expected alias '%s' to be registered", alias)
		}
	}
}

func TestRegisterTopicCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewCommandRegistry()
	RegisterTopicCommands(registry, tmpDir)

	// Test that topic commands are registered
	topicCommands := []string{
		"topic list",
		"topic info",
		"topic runs",
		"topic run",
		"topic rename",
		"topic search",
		"topic current",
	}

	for _, cmd := range topicCommands {
		_, ok := registry.GetHandler(cmd)
		if !ok {
			t.Errorf("Expected command '%s' to be registered", cmd)
		}
	}

	// Test that aliases are registered
	aliases := []string{
		"topics",
		"topic.info",
		"topic.runs",
	}

	for _, alias := range aliases {
		_, ok := registry.GetHandler(alias)
		if !ok {
			t.Errorf("Expected alias '%s' to be registered", alias)
		}
	}
}

func TestCmdTopicList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_list_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test sessions
	createTestSession(t, tmpDir, "test1", "Test session 1")
	createTestSession(t, tmpDir, "test2", "Test session 2")

	// Test topic list
	output, err := cmdTopicList(tmpDir, 10)
	if err != nil {
		t.Fatalf("cmdTopicList failed: %v", err)
	}

	if !strings.Contains(output, "Recent topics") {
		t.Errorf("Expected 'Recent topics' in output, got: %s", output)
	}
	if !strings.Contains(output, "test1") {
		t.Errorf("Expected 'test1' in output, got: %s", output)
	}
	if !strings.Contains(output, "test2") {
		t.Errorf("Expected 'test2' in output, got: %s", output)
	}

	// Test with limit
	output, err = cmdTopicList(tmpDir, 1)
	if err != nil {
		t.Fatalf("cmdTopicList with limit failed: %v", err)
	}
	// Should only show 1 topic
	count := strings.Count(output, "test")
	if count > 1 {
		t.Errorf("Expected limit of 1, got %d topics", count)
	}
}

func TestCmdTopicList_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	output, err := cmdTopicList(tmpDir, 10)
	if err != nil {
		t.Fatalf("cmdTopicList failed: %v", err)
	}

	if output != "No topics found" {
		t.Errorf("Expected 'No topics found', got: %s", output)
	}
}

func TestCmdTopicInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_info_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	createTestSession(t, tmpDir, "test-info", "Test session for info")

	output, err := cmdTopicInfo(tmpDir, "test-info")
	if err != nil {
		t.Fatalf("cmdTopicInfo failed: %v", err)
	}

	if !strings.Contains(output, "Topic: test-info") {
		t.Errorf("Expected topic key in output, got: %s", output)
	}
	if !strings.Contains(output, "Created:") {
		t.Errorf("Expected 'Created:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Messages:") {
		t.Errorf("Expected 'Messages:' in output, got: %s", output)
	}

	// Test with non-existent topic
	_, err = cmdTopicInfo(tmpDir, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent topic")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

func TestCmdTopicRuns(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_runs_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create session with multiple messages
	sess := &session.Session{
		Key:     "test-runs",
		Created: time.Now(),
		Updated: time.Now(),
		Messages: []providers.Message{
			{Role: "user", Content: "Message 1"},
			{Role: "assistant", Content: "Response 1"},
			{Role: "user", Content: "Message 2"},
			{Role: "assistant", Content: "Response 2"},
			{Role: "user", Content: "Message 3"},
		},
	}
	saveSession(tmpDir, sess)

	output, err := cmdTopicRuns(tmpDir, "test-runs", 10)
	if err != nil {
		t.Fatalf("cmdTopicRuns failed: %v", err)
	}

	if !strings.Contains(output, "Runs in topic test-runs") {
		t.Errorf("Expected topic key in output, got: %s", output)
	}

	// Test with limit
	output, err = cmdTopicRuns(tmpDir, "test-runs", 2)
	if err != nil {
		t.Fatalf("cmdTopicRuns with limit failed: %v", err)
	}
	// Should only show 2 runs
	if strings.Count(output, "Run") > 3 {
		t.Errorf("Expected limit of 2, got more runs")
	}
}

func TestCmdTopicSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_search_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create session with searchable content
	sess := &session.Session{
		Key:     "test-search",
		Created: time.Now(),
		Updated: time.Now(),
		Messages: []providers.Message{
			{Role: "user", Content: "Hello World"},
			{Role: "assistant", Content: "Goodbye World"},
		},
	}
	saveSession(tmpDir, sess)

	// Test search
	output, err := cmdTopicSearch(tmpDir, "test-search", "World")
	if err != nil {
		t.Fatalf("cmdTopicSearch failed: %v", err)
	}

	if !strings.Contains(output, "Found 2 matches") {
		t.Errorf("Expected 2 matches, got: %s", output)
	}

	// Test search with no matches
	output, err = cmdTopicSearch(tmpDir, "test-search", "NonExistent")
	if err != nil {
		t.Fatalf("cmdTopicSearch failed: %v", err)
	}

	if !strings.Contains(output, "No matches") {
		t.Errorf("Expected 'No matches', got: %s", output)
	}
}

func TestCmdTopicCurrent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_topic_current_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with no sessions
	output, err := cmdTopicCurrent(tmpDir)
	if err != nil {
		t.Fatalf("cmdTopicCurrent failed: %v", err)
	}

	if output != "No topics found" {
		t.Errorf("Expected 'No topics found', got: %s", output)
	}

	// Create a session
	createTestSession(t, tmpDir, "current-test", "Test")

	output, err = cmdTopicCurrent(tmpDir)
	if err != nil {
		t.Fatalf("cmdTopicCurrent failed: %v", err)
	}

	if !strings.Contains(output, "Current topic: current-test") {
		t.Errorf("Expected current topic, got: %s", output)
	}
}

func TestNewCommandRegistryFull(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_registry_full_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewCommandRegistryFull(tmpDir, tmpDir, false)

	// Verify FS commands are registered
	_, ok := registry.GetHandler("ls")
	if !ok {
		t.Error("Expected 'ls' command to be registered")
	}

	// Verify memory commands are registered
	_, ok = registry.GetHandler("memory store")
	if !ok {
		t.Error("Expected 'memory store' command to be registered")
	}

	// Verify topic commands are registered
	_, ok = registry.GetHandler("topic list")
	if !ok {
		t.Error("Expected 'topic list' command to be registered")
	}

	// Verify built-in commands are still available
	_, ok = registry.GetHandler("echo")
	if !ok {
		t.Error("Expected 'echo' command to be registered")
	}
}

// Helper function to create test sessions
func createTestSession(t *testing.T, storage, key, summary string) {
	t.Helper()

	sess := &session.Session{
		Key:      key,
		Summary:  summary,
		Created:  time.Now(),
		Updated:  time.Now(),
		Messages: []providers.Message{
			{Role: "user", Content: "Test message"},
		},
	}

	if err := saveSession(storage, sess); err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}
}

// TestParseDecisionJSON tests parsing decision records from JSON
func TestParseDecisionJSON(t *testing.T) {
	tests := []struct {
		name          string
		content       string
		wantTitle     string
		wantRationale string
		wantOk        bool
	}{
		{
			name:          "valid JSON with title and rationale",
			content:       `{"title":"Test Decision","rationale":"This is the rationale"}`,
			wantTitle:     "Test Decision",
			wantRationale: "This is the rationale",
			wantOk:        true,
		},
		{
			name:          "invalid JSON",
			content:       `{invalid json}`,
			wantTitle:     "",
			wantRationale: "",
			wantOk:        false,
		},
		{
			name:          "missing title",
			content:       `{"rationale":"This is the rationale"}`,
			wantTitle:     "",
			wantRationale: "",
			wantOk:        false,
		},
		{
			name:          "missing rationale",
			content:       `{"title":"Test Decision"}`,
			wantTitle:     "",
			wantRationale: "",
			wantOk:        false,
		},
		{
			name:          "not JSON",
			content:       "Just plain text",
			wantTitle:     "",
			wantRationale: "",
			wantOk:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, rationale, ok := parseDecisionJSON(tt.content)
			if ok != tt.wantOk {
				t.Errorf("parseDecisionJSON() ok = %v, want %v", ok, tt.wantOk)
			}
			if title != tt.wantTitle {
				t.Errorf("parseDecisionJSON() title = %q, want %q", title, tt.wantTitle)
			}
			if rationale != tt.wantRationale {
				t.Errorf("parseDecisionJSON() rationale = %q, want %q", rationale, tt.wantRationale)
			}
		})
	}
}

// TestParseFailureJSON tests parsing failure records from JSON
func TestParseFailureJSON(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantDesc       string
		wantResolution string
		wantOk         bool
	}{
		{
			name:           "valid JSON with description and resolution",
			content:        `{"description":"Something failed","resolution":"Fixed it"}`,
			wantDesc:       "Something failed",
			wantResolution: "Fixed it",
			wantOk:         true,
		},
		{
			name:           "missing description",
			content:        `{"resolution":"Fixed it"}`,
			wantDesc:       "",
			wantResolution: "",
			wantOk:         false,
		},
		{
			name:           "missing resolution",
			content:        `{"description":"Something failed"}`,
			wantDesc:       "",
			wantResolution: "",
			wantOk:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc, resolution, ok := parseFailureJSON(tt.content)
			if ok != tt.wantOk {
				t.Errorf("parseFailureJSON() ok = %v, want %v", ok, tt.wantOk)
			}
			if desc != tt.wantDesc {
				t.Errorf("parseFailureJSON() description = %q, want %q", desc, tt.wantDesc)
			}
			if resolution != tt.wantResolution {
				t.Errorf("parseFailureJSON() resolution = %q, want %q", resolution, tt.wantResolution)
			}
		})
	}
}

// TestParseNamedRecordJSON tests parsing named records from JSON
func TestParseNamedRecordJSON(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantDesc    string
		wantOk      bool
	}{
		{
			name:     "valid JSON with name and description",
			content:  `{"name":"Test Pattern","description":"This is a pattern"}`,
			wantName: "Test Pattern",
			wantDesc: "This is a pattern",
			wantOk:   true,
		},
		{
			name:     "missing name",
			content:  `{"description":"This is a pattern"}`,
			wantName: "",
			wantDesc: "",
			wantOk:   false,
		},
		{
			name:     "missing description",
			content:  `{"name":"Test Pattern"}`,
			wantName: "",
			wantDesc: "",
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, desc, ok := parseNamedRecordJSON(tt.content)
			if ok != tt.wantOk {
				t.Errorf("parseNamedRecordJSON() ok = %v, want %v", ok, tt.wantOk)
			}
			if name != tt.wantName {
				t.Errorf("parseNamedRecordJSON() name = %q, want %q", name, tt.wantName)
			}
			if desc != tt.wantDesc {
				t.Errorf("parseNamedRecordJSON() description = %q, want %q", desc, tt.wantDesc)
			}
		})
	}
}
