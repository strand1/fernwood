package tools

import (
	"strings"
	"testing"
)

func TestCheckOverflow(t *testing.T) {
	// Test case where no overflow occurs
	shortOutput := strings.Repeat("line\n", 10) // 10 lines, ~50 bytes
	result := CheckOverflow(shortOutput)
	
	if result.OverflowOccurred {
		t.Error("Expected no overflow for short output")
	}
	
	if result.TruncatedOutput != shortOutput {
		t.Error("Expected truncated output to equal original for short output")
	}
	
	// Test case where line limit is exceeded
	longOutputByLines := strings.Repeat("short line\n", 300) // 300 lines
	result = CheckOverflow(longOutputByLines)
	
	if !result.OverflowOccurred {
		t.Error("Expected overflow when line limit is exceeded")
	}
	
	lines := strings.Split(result.TruncatedOutput, "\n")
	if len(lines) > MaxOutputLines {
		t.Errorf("Expected truncated output to have at most %d lines, got %d", MaxOutputLines, len(lines))
	}
	
	if result.FullOutputPath == "" {
		t.Error("Expected full output path to be set when overflow occurs")
	}
	
	// Test case where byte limit is exceeded
	longOutputByBytes := strings.Repeat("this is a longer line with more content to exceed byte limits ", 1000)
	result = CheckOverflow(longOutputByBytes)
	
	if !result.OverflowOccurred {
		t.Error("Expected overflow when byte limit is exceeded")
	}
	
	if len(result.TruncatedOutput) > MaxOutputBytes {
		t.Errorf("Expected truncated output to have at most %d bytes, got %d", MaxOutputBytes, len(result.TruncatedOutput))
	}
	
	if result.FullOutputPath == "" {
		t.Error("Expected full output path to be set when overflow occurs")
	}
}

func TestFormatOverflowMessage(t *testing.T) {
	// Test formatting when no overflow occurred
	noOverflow := OverflowResult{
		TruncatedOutput:   "simple output",
		OverflowOccurred:  false,
		FullOutputPath:    "",
		OriginalLineCount: 1,
		OriginalByteCount: 13,
	}
	
	result := FormatOverflowMessage(noOverflow)
	if result != "simple output" {
		t.Error("Expected simple output when no overflow occurred")
	}
	
	// Test formatting when overflow occurred
	withOverflow := OverflowResult{
		TruncatedOutput:   "first part of output",
		OverflowOccurred:  true,
		FullOutputPath:    "/tmp/fernwood-output/test.txt",
		OriginalLineCount: 1000,
		OriginalByteCount: 50000,
	}
	
	result = FormatOverflowMessage(withOverflow)
	if !strings.Contains(result, "output truncated") {
		t.Error("Expected overflow message to contain truncation notice")
	}
	
	if !strings.Contains(result, "/tmp/fernwood-output/test.txt") {
		t.Error("Expected overflow message to contain full output path")
	}
	
	if !strings.Contains(result, "Explore:") {
		t.Error("Expected overflow message to contain exploration hints")
	}
}

func TestGetTempDir(t *testing.T) {
	tempDir := getTempDir()
	
	if tempDir == "" {
		t.Error("Expected temp directory path to be non-empty")
	}
	
	if !strings.Contains(tempDir, "fernwood-output") {
		t.Error("Expected temp directory to contain 'fernwood-output'")
	}
}