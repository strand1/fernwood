# AGENTS.md

🚨 **WARNING FOR AGENTS**: NEVER ADD UNTRACKED FILES TO GIT WITHOUT EXPLICIT PERMISSION! PAY ATTENTION TO .gitignore PATTERNS! 🚨

This file documents issues, tasks, and important information for agents working on the Fernwood codebase.

## Project Overview

**Fernwood** is an AI-powered coding assistant CLI tool built in Go. It provides a terminal-based interface for coding assistance with support for multiple AI providers and various development tools.

- **Main package**: `github.com/strand1/fernwood`
- **Binary**: `fernwood`
- **Build system**: GNU Make with Go
- **Primary language**: Go 1.26+

## ⚠️ CRITICAL: MULCH WORKFLOW - READ THIS BEFORE DOING ANYTHING ⚠️

This project uses **Mulch** (https://github.com/jayminwest/mulch) for codebase indexing, semantic search, and context management.

### 🔴 STOP - DO NOT EDIT FILES MANUALLY

DO NOT directly edit files in `.mulch/expertise/`. Doing so will break the schema validation and cause errors.

DO NOT randomly add files to git. Respect `.gitignore` patterns.

### ✅ CORRECT MULCH WORKFLOW (MANDATORY)

1. **Load context first**: Run `mulch prime` to understand current project state
2. **Search before implementing**: Run `mulch search "topic"` to avoid duplicating work
3. **Record learnings properly**: Use `mulch record <domain> --type <type> --description "..."` 
4. **Validate your changes**: Run `mulch validate` to check for schema errors
5. **Sync to repository**: Run `mulch sync` to commit expertise changes

### ⚠️ FAILURE MODES TO AVOID

❌ DO NOT edit `.mulch/expertise/*.jsonl` files directly
❌ DO NOT use random IDs like "mx-morrohsu-X" - let Mulch generate proper IDs
❌ DO NOT skip `mulch validate` before committing
❌ DO NOT add `.mulch/` files to git (they're intentionally gitignored)

### Installation

Fernwood includes built-in skills for Mulch. The main Mulch integration is managed through:

- `.mulch/` directory: Mulch configuration and data (GITIGNORED)
- Skills: Workspace skills for Mulch operations

### Using Mulch

When working on Fernwood, ALWAYS follow this process:

1. **Mulch data** is stored in `.mulch/` (gitignored FOR GOOD REASON)
2. Indexing is automatic on first run but can be triggered manually
3. The `mulch-context` skill provides Mulch integration capabilities
4. ALWAYS run `mulch prime` first to load project context
5. ALWAYS run `mulch validate` after making changes
6. ALWAYS run `mulch sync` to properly commit expertise

### Critical Commands (USE THESE OR FACE CONSEQUENCES)

- `mulch --help` - Understand the tool first
- `mulch prime` - Load all project context (RUN THIS FIRST)
- `mulch search "query"` - Find existing knowledge
- `mulch record <domain> --type <type> --description "content"` - Add new knowledge
- `mulch validate` - Check for errors (RUN THIS BEFORE SYNC)
- `mulch sync` - Commit expertise changes properly

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
7. **READ DOCUMENTATION FIRST**: Always read AGENTS.md and related documentation before making changes.
8. **USE MULCH PROPERLY**: Never edit Mulch files directly. Always use `mulch record` and `mulch sync`.
9. **CONTEXT LOADING**: Always run `mulch prime` before starting work to understand project context.
10. **VALIDATION**: Always run `mulch validate` after making changes to catch errors early.
11. **GIT DISCIPLINE**: NEVER RUN `git add .` - ALWAYS CHECK `git status` FIRST - ONLY ADD FILES THAT SHOULD BE TRACKED!

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
