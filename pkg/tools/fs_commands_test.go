// Fernwood - A lightweight agentic coding harness forked from PicoClaw
// License: MIT
//
// Copyright (c) 2026 Fernwood contributors

package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdLs(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "test_ls_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files and directories
	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)

	// Test ls
	output, err := cmdLs(tmpDir, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdLs failed: %v", err)
	}

	// Check output contains expected entries
	if !strings.Contains(output, "FILE: file1.txt") {
		t.Errorf("Expected output to contain 'FILE: file1.txt', got: %s", output)
	}
	if !strings.Contains(output, "FILE: file2.go") {
		t.Errorf("Expected output to contain 'FILE: file2.go', got: %s", output)
	}
	if !strings.Contains(output, "DIR:  subdir/") {
		t.Errorf("Expected output to contain 'DIR:  subdir/', got: %s", output)
	}

	// Directories should come before files
	dirIdx := strings.Index(output, "DIR:")
	fileIdx := strings.Index(output, "FILE:")
	if dirIdx > fileIdx {
		t.Error("Expected directories to be listed before files")
	}
}

func TestCmdCat(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_cat_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Hello, World!\nThis is a test file."
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Test cat
	output, err := cmdCat(tmpFile.Name(), filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdCat failed: %v", err)
	}

	if output != content {
		t.Errorf("Expected %q, got %q", content, output)
	}

	// Test cat on non-existent file
	tmpDir := filepath.Dir(tmpFile.Name())
	_, err = cmdCat("/nonexistent/file.txt", tmpDir, false)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if !strings.Contains(err.Error(), "No such file") {
		t.Errorf("Expected 'No such file' error, got: %v", err)
	}
}

func TestCmdCat_Binary(t *testing.T) {
	// Create temp binary file
	tmpFile, err := os.CreateTemp("", "test_cat_*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write binary content
	binaryData := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	os.WriteFile(tmpFile.Name(), binaryData, 0644)

	// Test cat on binary file
	output, err := cmdCat(tmpFile.Name(), filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdCat failed: %v", err)
	}

	// Should return binary error message, not actual content
	if !strings.Contains(output, "[error]") || !strings.Contains(output, "binary") {
		t.Errorf("Expected binary file error, got: %s", output)
	}
}

func TestCmdWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_write_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "test.txt")
	content := "Test content\nLine 2"

	// Test write
	output, err := cmdWrite(testPath, content, filepath.Dir(testPath), false)
	if err != nil {
		t.Fatalf("cmdWrite failed: %v", err)
	}

	if !strings.Contains(output, "Written") {
		t.Errorf("Expected 'Written' in output, got: %s", output)
	}

	// Verify file was written
	readContent, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read written file: %v", err)
	}
	if string(readContent) != content {
		t.Errorf("Expected %q, got %q", content, readContent)
	}

	// Test write with directory creation
	nestedPath := filepath.Join(tmpDir, "subdir", "nested", "file.txt")
	output, err = cmdWrite(nestedPath, "nested content", tmpDir, false)
	if err != nil {
		t.Fatalf("cmdWrite with nested path failed: %v", err)
	}
	if !strings.Contains(output, "Written") {
		t.Errorf("Expected 'Written' in output, got: %s", output)
	}
}

func TestCmdStat(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_stat_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Test content"
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Test stat
	output, err := cmdStat(tmpFile.Name(), filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdStat failed: %v", err)
	}

	// Check output contains expected fields
	if !strings.Contains(output, "File:") {
		t.Errorf("Expected 'File:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Size:") {
		t.Errorf("Expected 'Size:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Type:") {
		t.Errorf("Expected 'Type:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Modified:") {
		t.Errorf("Expected 'Modified:' in output, got: %s", output)
	}
	if !strings.Contains(output, "Mode:") {
		t.Errorf("Expected 'Mode:' in output, got: %s", output)
	}

	// Test stat on non-existent file
	_, err = cmdStat("/nonexistent/file.txt", "", false)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCmdRm(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_rm_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	// Verify file exists
	if _, err := os.Stat(tmpPath); os.IsNotExist(err) {
		t.Fatal("Test file doesn't exist")
	}

	// Test rm
	output, err := cmdRm(tmpPath, filepath.Dir(tmpPath), false)
	if err != nil {
		t.Fatalf("cmdRm failed: %v", err)
	}

	if !strings.Contains(output, "Removed") {
		t.Errorf("Expected 'Removed' in output, got: %s", output)
	}

	// Verify file was removed
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("File still exists after rm")
	}

	// Test rm on non-existent file
	_, err = cmdRm("/nonexistent/file.txt", "", false)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCmdCp(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_cp_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")
	content := "Source content"
	os.WriteFile(srcPath, []byte(content), 0644)

	// Test cp
	output, err := cmdCp(srcPath, dstPath, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdCp failed: %v", err)
	}

	if !strings.Contains(output, "Copied") {
		t.Errorf("Expected 'Copied' in output, got: %s", output)
	}

	// Verify file was copied
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Expected %q, got %q", content, dstContent)
	}

	// Test cp on non-existent source
	_, err = cmdCp("/nonexistent/source.txt", dstPath, "", false)
	if err == nil {
		t.Error("Expected error for non-existent source")
	}
}

func TestCmdMv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mv_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "dest.txt")
	content := "Source content"
	os.WriteFile(srcPath, []byte(content), 0644)

	// Test mv
	output, err := cmdMv(srcPath, dstPath, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdMv failed: %v", err)
	}

	if !strings.Contains(output, "Moved") {
		t.Errorf("Expected 'Moved' in output, got: %s", output)
	}

	// Verify file was moved (dest exists, source doesn't)
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("Destination file doesn't exist after mv")
	}
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("Source file still exists after mv")
	}

	// Verify content
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Expected %q, got %q", content, dstContent)
	}
}

func TestCmdMkdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_mkdir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "newdir")

	// Test mkdir
	output, err := cmdMkdir(testPath, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdMkdir failed: %v", err)
	}

	if !strings.Contains(output, "Created directory") {
		t.Errorf("Expected 'Created directory' in output, got: %s", output)
	}

	// Verify directory was created
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Failed to stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}

	// Test mkdir with nested path
	nestedPath := filepath.Join(tmpDir, "a", "b", "c")
	output, err = cmdMkdir(nestedPath, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdMkdir with nested path failed: %v", err)
	}
	if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
		t.Error("Nested directory was not created")
	}
}

func TestCmdGrep(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_grep_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "Hello World\nFoo Bar\nHello Again\nGoodbye"
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	// Test grep basic
	output, err := cmdGrep([]string{"Hello", tmpFile.Name()}, "", filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdGrep failed: %v", err)
	}

	if !strings.Contains(output, "Hello World") {
		t.Errorf("Expected 'Hello World' in output, got: %s", output)
	}
	if !strings.Contains(output, "Hello Again") {
		t.Errorf("Expected 'Hello Again' in output, got: %s", output)
	}
	if strings.Contains(output, "Foo Bar") {
		t.Errorf("Unexpected 'Foo Bar' in output, got: %s", output)
	}

	// Test grep -i (case insensitive)
	output, err = cmdGrep([]string{"-i", "hello", tmpFile.Name()}, "", filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdGrep -i failed: %v", err)
	}
	if !strings.Contains(output, "Hello World") {
		t.Errorf("Expected 'Hello World' in case-insensitive output, got: %s", output)
	}

	// Test grep -c (count)
	output, err = cmdGrep([]string{"-c", "Hello", tmpFile.Name()}, "", filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdGrep -c failed: %v", err)
	}
	if output != "2" {
		t.Errorf("Expected count '2', got: %s", output)
	}

	// Test grep -v (invert)
	output, err = cmdGrep([]string{"-v", "Hello", tmpFile.Name()}, "", filepath.Dir(tmpFile.Name()), false)
	if err != nil {
		t.Fatalf("cmdGrep -v failed: %v", err)
	}
	if strings.Contains(output, "Hello") {
		t.Errorf("Unexpected 'Hello' in inverted output, got: %s", output)
	}

	// Test grep from stdin
	output, err = cmdGrep([]string{"Hello"}, content, "", false)
	if err != nil {
		t.Fatalf("cmdGrep from stdin failed: %v", err)
	}
	if !strings.Contains(output, "Hello World") {
		t.Errorf("Expected 'Hello World' in stdin output, got: %s", output)
	}
}

func TestCmdHead(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tmpDir := t.TempDir()

	// Test head default (10 lines)
	output, err := cmdHead([]string{}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdHead failed: %v", err)
	}
	if output != content {
		t.Errorf("Expected full content, got: %s", output)
	}

	// Test head -n 3
	output, err = cmdHead([]string{"-n", "3"}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdHead -n failed: %v", err)
	}
	expected := "Line 1\nLine 2\nLine 3"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	// Test head from file
	tmpFile, err := os.CreateTemp("", "test_head_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	output, err = cmdHead([]string{"-n", "2", tmpFile.Name()}, "", tmpDir, false)
	if err != nil {
		t.Fatalf("cmdHead from file failed: %v", err)
	}
	expected = "Line 1\nLine 2"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestCmdTail(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tmpDir := t.TempDir()

	// Test tail default (10 lines)
	output, err := cmdTail([]string{}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdTail failed: %v", err)
	}
	if output != content {
		t.Errorf("Expected full content, got: %s", output)
	}

	// Test tail -n 3
	output, err = cmdTail([]string{"-n", "3"}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdTail -n failed: %v", err)
	}
	expected := "Line 3\nLine 4\nLine 5"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}

	// Test tail from file
	tmpFile, err := os.CreateTemp("", "test_tail_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	output, err = cmdTail([]string{"-n", "2", tmpFile.Name()}, "", tmpDir, false)
	if err != nil {
		t.Fatalf("cmdTail from file failed: %v", err)
	}
	expected = "Line 4\nLine 5"
	if output != expected {
		t.Errorf("Expected %q, got %q", expected, output)
	}
}

func TestCmdWc(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3\n"
	tmpDir := t.TempDir()

	// Test wc default (all counts)
	output, err := cmdWc([]string{}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdWc failed: %v", err)
	}
	// Should have: lines words chars (3 lines, 6 words, 21 chars)
	if !strings.Contains(output, "3") {
		t.Errorf("Expected line count in output, got: %s", output)
	}

	// Test wc -l (lines only)
	output, err = cmdWc([]string{"-l"}, content, tmpDir, false)
	if err != nil {
		t.Fatalf("cmdWc -l failed: %v", err)
	}
	if !strings.Contains(output, "3") {
		t.Errorf("Expected '3' in output, got: %s", output)
	}

	// Test wc from file
	tmpFile, err := os.CreateTemp("", "test_wc_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	os.WriteFile(tmpFile.Name(), []byte(content), 0644)

	output, err = cmdWc([]string{"-l", tmpFile.Name()}, "", tmpDir, false)
	if err != nil {
		t.Fatalf("cmdWc from file failed: %v", err)
	}
	if !strings.Contains(output, "3") || !strings.Contains(output, tmpFile.Name()) {
		t.Errorf("Expected line count and filename in output, got: %s", output)
	}
}

func TestRegisterFSCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_registry_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewCommandRegistry()
	RegisterFSCommands(registry, tmpDir, false)

	// Test that FS commands are registered
	fsCommands := []string{"ls", "cat", "write", "stat", "rm", "cp", "mv", "mkdir", "grep", "head", "tail", "wc"}
	for _, cmd := range fsCommands {
		_, ok := registry.GetHandler(cmd)
		if !ok {
			t.Errorf("Expected command '%s' to be registered", cmd)
		}
	}

	// Test that aliases are registered
	aliases := []string{"fs.ls", "fs.cat", "fs.write", "fs.stat", "fs.rm", "fs.cp", "fs.mv", "fs.mkdir"}
	for _, alias := range aliases {
		_, ok := registry.GetHandler(alias)
		if !ok {
			t.Errorf("Expected alias '%s' to be registered", alias)
		}
	}

	// Test executing a command through registry
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	output := registry.Exec("cat "+testFile, "")
	if !strings.Contains(output, "test content") {
		t.Errorf("Expected 'test content' in output, got: %s", output)
	}
}

func TestNewCommandRegistryWithFS(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test_registry_fs_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry := NewCommandRegistryWithFS(tmpDir, false)

	// Verify FS commands are registered
	_, ok := registry.GetHandler("ls")
	if !ok {
		t.Error("Expected 'ls' command to be registered")
	}

	// Verify built-in commands are still available
	_, ok = registry.GetHandler("echo")
	if !ok {
		t.Error("Expected 'echo' command to be registered")
	}
	_, ok = registry.GetHandler("help")
	if !ok {
		t.Error("Expected 'help' command to be registered")
	}
}

func TestCmdLsWithFlags(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create test files
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("hello"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	
	// Test ls -la (should use shell ls)
	result, err := cmdShellLs([]string{"-la"}, tmpDir, false)
	if err != nil {
		t.Fatalf("ls -la failed: %v", err)
	}
	
	// Should contain detailed output
	if !strings.Contains(result, "test.txt") {
		t.Errorf("Expected test.txt in output, got: %s", result)
	}
	if !strings.Contains(result, "subdir") {
		t.Errorf("Expected subdir in output, got: %s", result)
	}
	// -l flag should show permissions/ownership
	if !strings.Contains(result, "drwx") && !strings.Contains(result, "-rw") {
		t.Logf("Note: ls -la output doesn't show permissions: %s", result)
	}
}
