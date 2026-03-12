package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	
	"github.com/strand1/fernwood/pkg/tools"
)

func main() {
	fmt.Println("Morrohsu Unix-Style CLI Agent Architecture Demo")
	fmt.Println("===============================================")
	
	// Initialize the command registry
	registry := tools.InitializeRegistry()
	
	// Create the unified run tool
	runTool := tools.NewRunTool(registry)
	
	// Demonstrate binary detection
	demoBinaryDetection()
	
	// Demonstrate command execution
	demoCommandExecution(runTool)
}

func demoBinaryDetection() {
	fmt.Println("\n1. Binary Detection Demo:")
	fmt.Println("------------------------")
	
	// Text content
	textData := []byte("Hello, world!\nThis is a text file.\n")
	fmt.Printf("Text data is binary: %v\n", tools.IsBinary(textData))
	
	// Binary content with null bytes
	binaryData := []byte("Hello\x00World")
	fmt.Printf("Binary data (null bytes) is binary: %v\n", tools.IsBinary(binaryData))
	
	// Image file detection
	imagePath := "photo.png"
	fileType := tools.DetectBinaryType([]byte("dummy"), imagePath)
	fmt.Printf("File '%s' detected as: %s\n", imagePath, fileType)
	
	// PDF detection
	pdfData := []byte("%PDF-1.4\nsome content")
	pdfType := tools.DetectBinaryType(pdfData, "document.pdf")
	fmt.Printf("PDF data detected as: %s\n", pdfType)
	
	// Human readable sizes
	sizes := []int64{512, 1024, 1024 * 1024, 1024 * 1024 * 1024}
	for _, size := range sizes {
		fmt.Printf("Size %d bytes = %s\n", size, tools.HumanSize(size))
	}
}

func demoCommandExecution(runTool *tools.RunTool) {
	fmt.Println("\n2. Command Execution Demo:")
	fmt.Println("--------------------------")
	
	ctx := context.Background()
	
	// Create a test file
	testContent := "This is a test file for demonstration.\nIt has multiple lines.\nIncluding a line with the word ERROR.\n"
	err := os.WriteFile("demo_test.txt", []byte(testContent), 0644)
	if err != nil {
		fmt.Printf("Warning: Could not create test file: %v\n", err)
		return
	}
	defer os.Remove("demo_test.txt")
	
	// Demonstrate cat command
	fmt.Println("Executing: run command=\"cat demo_test.txt\"")
	args := map[string]any{
		"command": "cat demo_test.txt",
	}
	result := runTool.Execute(ctx, args)
	fmt.Printf("Result:\n%s\n", formatResult(result))
	
	// Demonstrate grep command
	fmt.Println("Executing: run command=\"grep ERROR demo_test.txt\"")
	args = map[string]any{
		"command": "grep ERROR demo_test.txt",
	}
	result = runTool.Execute(ctx, args)
	fmt.Printf("Result:\n%s\n", formatResult(result))
	
	// Demonstrate ls command
	fmt.Println("Executing: run command=\"ls .\"")
	args = map[string]any{
		"command": "ls .",
	}
	result = runTool.Execute(ctx, args)
	fmt.Printf("Result:\n%s\n", formatResult(result))
	
	// Demonstrate help system
	fmt.Println("Executing: run command=\"help\"")
	args = map[string]any{
		"command": "help",
	}
	result = runTool.Execute(ctx, args)
	fmt.Printf("Result:\n%s\n", formatResult(result))
	
	// Demonstrate command chaining parsing
	fmt.Println("Command Chaining Parsing Demo:")
	fmt.Println("-----------------------------")
	chain := "cat demo_test.txt | grep ERROR | wc -l"
	segments := tools.ParseChain(chain)
	fmt.Printf("Parsing chain: %s\n", chain)
	for i, segment := range segments {
		fmt.Printf("  Segment %d: Command='%s', Operator='%s'\n", 
			i, segment.Command, segment.Operator.String())
	}
}

func formatResult(result *tools.ToolResult) string {
	var builder strings.Builder
	
	if result.IsError {
		builder.WriteString("[ERROR] ")
	}
	
	builder.WriteString(result.ForLLM)
	
	if result.Media != nil && len(result.Media) > 0 {
		builder.WriteString(fmt.Sprintf("\n[Media: %v]", result.Media))
	}
	
	if result.Async {
		builder.WriteString("\n[Async operation started]")
	}
	
	return builder.String()
}