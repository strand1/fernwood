# AGENTS.md

This file documents issues, tasks, and important information for agents working on the Fernwood codebase.

## Project Overview

**Fernwood** is an AI-powered coding assistant CLI tool built in Go. It provides a terminal-based interface for coding assistance with support for multiple AI providers and various development tools.

- **Main package**: `github.com/strand1/fernwood`
- **Binary**: `fernwood`
- **Build system**: GNU Make with Go
- **Primary language**: Go 1.26+

## Mulch Integration

This project uses **Mulch** (https://github.com/jayminwest/mulch) for codebase indexing, semantic search, and context management.

### Installation

Fernwood includes built-in skills for Mulch. The main Mulch integration is managed through:

- `.mulch/` directory: Mulch configuration and data
- Skills: Workspace skills for Mulch operations

### Using Mulch

When working on Fernwood, be aware that:

1. **Mulch data** is stored in `.mulch/` (gitignored)
2. Indexing is automatic on first run but can be triggered manually
3. The `mulch-context` skill provides Mulch integration capabilities

### Mulch Versions

- **Mulch Prime**: The latest stable version with all features
- Use `mulch` or other variants as needed for specific use cases

Refer to the Mulch documentation for advanced usage: https://github.com/jayminwest/mulch

## Current Issues & Tasks

### Known Problems

1. **Binary Size**: Binaries are large (~16-17MB) due to embedded assets and dependencies. Consider:
   - UPX compression
   - Stripping symbols more aggressively
   - Asset optimization

2. **Test Failures**: See `TEST_FAILURES.md` for details on failing tests.

3. **MIPS Compatibility**: Requires e_flag patching for NaN2008 kernels (see Makefile).

### Active Tasks

1. **Feature Development**:
   - Improve MCP (Model Context Protocol) integration
   - Enhance Docker support and isolation
   - Add more skill templates

2. **Performance**:
   - Reduce memory footprint
   - Speed up startup time
   - Optimize Mulch indexing

3. **Testing**:
   - Fix unit tests in various packages
   - Add integration tests for Mulch workflows
   - Cross-platform testing (especially ARM)

### Architecture Notes

- **Skills system**: Extensible via `~/.fernwood/workspace/skills/`
- **Workspace**: Default location `~/.fernwood/workspace`
- **Configuration**: JSON-based config files
- **Provider abstraction**: Supports multiple AI providers through interfaces

## Build Instructions

```bash
# Build for current platform
make build

# Build for specific architectures
make build-linux-arm64   # ARM64 (Raspberry Pi 4/5, Apple Silicon)
# For amd64: builds to build/fernwood-linux-amd64 by default on x86_64

# Build for all platforms
make build-all

# Install
make install

# Test
make test
```

## Development Guidelines

1. **Code Generation**: Run `make generate` before committing to ensure generated code is up-to-date.
2. **Go Version**: Use Go 1.26+ (project uses modern Go features).
3. **Linting**: Run `make lint` and fix issues before PRs.
4. **Git Hooks**: Pre-commit hooks may run linters and tests.
5. **File Management**: Do not randomly add files to the repository. Pay attention to `.gitignore` patterns and respect intentionally untracked files/directories.
6. **Mulch Data**: Files in `.mulch/` are gitignored for a reason - do not commit them unless explicitly required for sharing expertise contexts.

## Environment Variables

Key environment variables:
- `FERNWOOD_HOME`: Override default `~/.fernwood` location
- `WORKSPACE_DIR`: Override workspace directory
- `INSTALL_PREFIX`: Installation prefix (default `~/.local`)
- `VERSION`: Override build version string

## Important Files & Directories

```
.
├── cmd/fernwood/        # Main entry point
├── pkg/                 # Core packages
│   ├── agent/          # Agent logic
│   ├── config/         # Configuration management
│   ├── mulch/          # Mulch integration
│   └── ...
├── skills/              # Built-in skills
├── .mulch/              # Mulch data (gitignored)
├── build/               # Build outputs
├── Makefile             # Build orchestration
├── .golangci.yaml       # Linter config
├── go.mod              # Go dependencies
└── AGENTS.md           # This file
```

## Contact & Resources

- **Repo**: https://github.com/strand1/fernwood
- **Issues**: Use GitHub Issues for bugs and feature requests
- **Mulch**: https://github.com/jayminwest/mulch

---

**Last Updated**: 2026-03-12
**Agent Version**: 0.1.0
