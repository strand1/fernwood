package tools

import (
	"testing"
)

func TestIsBinary(t *testing.T) {
	// Test text content
	textData := []byte("Hello, world!\nThis is a text file.\n")
	if IsBinary(textData) {
		t.Error("Expected text data to not be identified as binary")
	}

	// Test binary content with null bytes
	binaryData := []byte("Hello\x00World")
	if !IsBinary(binaryData) {
		t.Error("Expected binary data with null bytes to be identified as binary")
	}

	// Test binary content with high control character ratio
	controlData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C}
	if !IsBinary(controlData) {
		t.Error("Expected data with high control character ratio to be identified as binary")
	}

	// Test valid UTF-8 text
	utf8Data := []byte("Hello, 世界! 🌍")
	if IsBinary(utf8Data) {
		t.Error("Expected valid UTF-8 text to not be identified as binary")
	}
}

func TestDetectBinaryType(t *testing.T) {
	// Test image file detection by extension
	imageTypes := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".webp", ".tiff", ".svg"}
	for _, ext := range imageTypes {
		fileType := DetectBinaryType([]byte("dummy"), "test"+ext)
		if fileType != "image" {
			t.Errorf("Expected %s to be detected as image, got %s", ext, fileType)
		}
	}

	// Test PDF detection
	pdfData := []byte("%PDF-1.4\nsome content")
	fileType := DetectBinaryType(pdfData, "document.pdf")
	if fileType != "pdf" {
		t.Errorf("Expected PDF data to be detected as pdf, got %s", fileType)
	}

	// Test archive detection by extension
	archiveTypes := []string{".zip", ".tar", ".gz", ".rar", ".7z"}
	for _, ext := range archiveTypes {
		fileType := DetectBinaryType([]byte("dummy"), "archive"+ext)
		if fileType != "archive" {
			t.Errorf("Expected %s to be detected as archive, got %s", ext, fileType)
		}
	}

	// Test text file
	textData := []byte("This is plain text content")
	fileType2 := DetectBinaryType(textData, "text.txt")
	if fileType2 != "text" {
		t.Errorf("Expected text file to be detected as text, got %s", fileType2)
	}
}

func TestIsImageFile(t *testing.T) {
	// Test image files
	imageFiles := []string{"photo.png", "image.jpg", "graphic.gif", "picture.jpeg"}
	for _, filename := range imageFiles {
		if !IsImageFile(filename) {
			t.Errorf("Expected %s to be identified as an image file", filename)
		}
	}

	// Test non-image files
	nonImageFiles := []string{"document.txt", "program.go", "data.json", "archive.zip"}
	for _, filename := range nonImageFiles {
		if IsImageFile(filename) {
			t.Errorf("Expected %s to not be identified as an image file", filename)
		}
	}
}

func TestHumanSize(t *testing.T) {
	testCases := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024*1024 - 1, "1024.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
	}

	for _, tc := range testCases {
		result := HumanSize(tc.bytes)
		if result != tc.expected {
			t.Errorf("HumanSize(%d) = %s; expected %s", tc.bytes, result, tc.expected)
		}
	}
}