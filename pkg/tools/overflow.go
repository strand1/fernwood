package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	// MaxOutputLines is the maximum number of lines to return in a tool result
	MaxOutputLines = 200
	
	// MaxOutputBytes is the maximum number of bytes to return in a tool result
	MaxOutputBytes = 50 * 1024 // 50KB
)

// OverflowResult represents the result of checking for output overflow
type OverflowResult struct {
	// TruncatedOutput is the truncated output that fits within limits
	TruncatedOutput string
	
	// OverflowOccurred indicates if the output was truncated
	OverflowOccurred bool
	
	// FullOutputPath is the path to the temporary file containing full output
	FullOutputPath string
	
	// OriginalLineCount is the number of lines in the original output
	OriginalLineCount int
	
	// OriginalByteCount is the number of bytes in the original output
	OriginalByteCount int
}

// CheckOverflow determines if output exceeds limits and handles accordingly
func CheckOverflow(output string) OverflowResult {
	result := OverflowResult{
		TruncatedOutput:   output,
		OverflowOccurred:  false,
		FullOutputPath:    "",
		OriginalLineCount: 0,
		OriginalByteCount: len(output),
	}
	
	// Count lines
	lines := strings.Split(output, "\n")
	result.OriginalLineCount = len(lines)
	
	// Check if we exceed either limit
	byteLimitExceeded := len(output) > MaxOutputBytes
	lineLimitExceeded := len(lines) > MaxOutputLines
	
	if byteLimitExceeded || lineLimitExceeded {
		result.OverflowOccurred = true
		
		// Write full output to temporary file
		tempDir := getTempDir()
		filename := fmt.Sprintf("cmd-%s.txt", time.Now().Format("20060102-150405"))
		fullPath := filepath.Join(tempDir, filename)
		
		if err := os.WriteFile(fullPath, []byte(output), 0644); err == nil {
			result.FullOutputPath = fullPath
		}
		
		// Truncate output based on which limit was hit first
		truncated := ""
		if lineLimitExceeded {
			// Truncate to MaxOutputLines
			if MaxOutputLines < len(lines) {
				truncated = strings.Join(lines[:MaxOutputLines], "\n")
			} else {
				truncated = strings.Join(lines, "\n")
			}
		} else {
			// Truncate to MaxOutputBytes
			if MaxOutputBytes < len(output) {
				truncated = output[:MaxOutputBytes]
			} else {
				truncated = output
			}
		}
		
		result.TruncatedOutput = truncated
	}
	
	return result
}

// FormatOverflowMessage creates a user-friendly message when overflow occurs
func FormatOverflowMessage(result OverflowResult) string {
	if !result.OverflowOccurred {
		return result.TruncatedOutput
	}
	
	// Create exploration hints
	humanSize := HumanSize(int64(result.OriginalByteCount))
	message := fmt.Sprintf("%s\n\n--- output truncated (%d lines, %s) ---\nFull output: %s\nExplore: cat %s | grep <pattern>\n         cat %s | tail 100",
		result.TruncatedOutput,
		result.OriginalLineCount,
		humanSize,
		result.FullOutputPath,
		result.FullOutputPath,
		result.FullOutputPath)
	
	return message
}

// getTempDir returns the directory for storing overflow files
func getTempDir() string {
	tempDir := filepath.Join(os.TempDir(), "fernwood-output")
	
	// Create directory if it doesn't exist
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		os.MkdirAll(tempDir, 0755)
	}
	
	return tempDir
}