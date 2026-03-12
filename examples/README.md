# Morrohsu Examples

This directory contains examples demonstrating the Morrohsu Unix-Style CLI Agent Architecture implementation.

## Examples

### morrohsu_demo.go

Demonstrates the core features of the Morrohsu architecture:

1. **Binary Detection** - Shows how text and binary content are identified
2. **File Type Classification** - Demonstrates detection of images, PDFs, etc.
3. **Human-readable Sizes** - Shows file size formatting
4. **Command Execution** - Demonstrates the unified `run` tool
5. **Command Chaining Parsing** - Shows how command chains are parsed

To run:
```bash
go run morrohsu_demo.go
```

## Key Features Demonstrated

- **Binary Guard**: Prevents binary content from flooding the context
- **stderr Visibility**: Ensures error messages are always visible
- **Overflow Protection**: Handles large outputs gracefully
- **Metadata Footer**: Provides exit codes and timing information
- **Command Chaining**: Parses complex command sequences
- **Progressive Help**: Offers guidance through the help system
- **Unified Tool Registry**: Single `run` tool replaces multiple discrete tools

## Architecture Layers

### Layer 1: Unix Execution Layer
- Raw command semantics
- Proper exit codes
- Binary detection
- Pipe-compatible output

### Layer 2: LLM Presentation Layer
- Cognitive constraints applied post-execution
- Binary guard
- Overflow truncation with spillover
- Metadata footer
- stderr attachment on failure

## Related Files

The implementation spans multiple files in `pkg/tools/`:

- `binary.go` - Binary detection utilities
- `chain.go` - Command chaining parser
- `output.go` - Presentation layer logic
- `overflow.go` - Output truncation and spillover
- `commands.go` - Command handler implementations
- `registry.go` - Unified tool registry

See `spec/morrohsu_spec.md` for the full specification and `spec/morrohsu_implementation_summary.md` for implementation details.