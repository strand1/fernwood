package tools

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// FormatOutput applies the presentation layer formatting to command output
func FormatOutput(output string, exitCode int, duration time.Duration, stderr string) string {
	var result strings.Builder
	
	// Add the main output
	if output != "" {
		result.WriteString(output)
	}
	
	// Add stderr if present (fix for Story 2)
	if stderr != "" {
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString("[stderr]\n")
		result.WriteString(stderr)
	}
	
	// Add metadata footer (Story 4)
	metadata := fmt.Sprintf("[exit:%d | %s]", exitCode, formatDuration(duration))
	
	if result.Len() > 0 {
		result.WriteString("\n")
	}
	result.WriteString(metadata)
	
	// Check for overflow and handle appropriately
	formattedOutput := result.String()
	overflowResult := CheckOverflow(formattedOutput)
	
	if overflowResult.OverflowOccurred {
		return FormatOverflowMessage(overflowResult)
	}
	
	return formattedOutput
}

// FormatBinaryError creates a user-friendly error message for binary files
func FormatBinaryError(path string, size int64) string {
	fileType := DetectBinaryType(nil, path) // We only need extension check here
	humanSize := HumanSize(size)
	
	switch fileType {
	case "image":
		return fmt.Sprintf("[error] cat: binary image file (%s). Use: see %s", humanSize, filepath.Base(path))
	case "pdf":
		return fmt.Sprintf("[error] cat: binary pdf file (%s). Use: see %s", humanSize, filepath.Base(path))
	case "archive":
		return fmt.Sprintf("[error] cat: binary archive file (%s). Use: cat -b %s for base64", humanSize, filepath.Base(path))
	default:
		return fmt.Sprintf("[error] cat: binary file (%s). Use: cat -b %s for base64", humanSize, filepath.Base(path))
	}
}

// FormatCommandError creates a standardized error message with recovery guidance
func FormatCommandError(command string, errorMessage string) string {
	// Specific error handling for common cases
	if strings.Contains(errorMessage, "command not found") {
		return fmt.Sprintf("[error] %s\nAvailable: cat, ls, see, write, grep, memory, clip, ...", errorMessage)
	}
	
	// Generic error with recovery guidance
	return fmt.Sprintf("[error] %s", errorMessage)
}

// formatDuration converts a duration to a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	} else if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
}

// ApplyBinaryGuard checks if output is binary and formats an appropriate error if so
func ApplyBinaryGuard(data []byte, path string) (string, bool) {
	if IsBinary(data) {
		size := int64(len(data))
		errorMsg := FormatBinaryError(path, size)
		return errorMsg, true
	}
	return "", false
}