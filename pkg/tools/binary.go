package tools

import (
	"bytes"
	"fmt"
	"path/filepath"
	"unicode/utf8"
)

// IsBinary determines if the given data is likely binary content.
// It checks for null bytes, valid UTF-8 encoding, and the ratio of control characters.
func IsBinary(data []byte) bool {
	// If data contains null bytes, it's very likely binary
	if bytes.Contains(data, []byte{0}) {
		return true
	}

	// If data is not valid UTF-8, it's likely binary
	if !utf8.Valid(data) {
		return true
	}

	// Count control characters (except common whitespace)
	controlCount := 0
	for _, b := range data {
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			controlCount++
		}
	}

	// If more than 10% of bytes are control characters, consider it binary
	return len(data) > 0 && float64(controlCount)/float64(len(data)) > 0.1
}

// DetectBinaryType attempts to determine the specific type of binary file.
// Returns "image", "pdf", "archive", or "binary" based on file content and extension.
func DetectBinaryType(data []byte, path string) string {
	// Check file extension first for quick identification
	ext := filepath.Ext(path)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".tiff", ".svg":
		return "image"
	case ".pdf":
		return "pdf"
	case ".zip", ".tar", ".gz", ".rar", ".7z":
		return "archive"
	}

	// Check magic bytes for more accurate detection
	if len(data) >= 4 {
		// PDF magic number
		if bytes.HasPrefix(data, []byte("%PDF")) {
			return "pdf"
		}
		
		// PNG magic number
		if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
			return "image"
		}
		
		// JPEG magic numbers
		if len(data) >= 10 && 
		   ((data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF) ||
		    (bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF, 0xE0}) && 
		     bytes.Equal(data[6:10], []byte("JFIF")))) {
			return "image"
		}
		
		// ZIP magic number
		if data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
			return "archive"
		}
	}

	// If we determined it's binary but couldn't classify further
	if IsBinary(data) {
		return "binary"
	}

	return "text"
}

// IsImageFile quickly checks if a file is likely an image based on its extension.
func IsImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".tiff", ".svg":
		return true
	default:
		return false
	}
}

// HumanSize converts a byte count to a human-readable string.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}