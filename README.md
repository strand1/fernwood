<p align="center">
  <img src="assets/fernwood_logo.png" alt="Fernwood" width="200"/>
</p>

<h1 align="center">Fernwood</h1>

<p align="center">
  <strong>Terminal-based coding agent.</strong><br/>
  Forked from [PicoClaw](https://github.com/sipeed/picoclaw), with persistent memory.
</p>

<p align="center">
  <a href="https://golang.org/dl/"><img src="https://img.shields.io/badge/Go-1.26.1-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green" alt="MIT License"/></a>
  <img src="https://img.shields.io/badge/status-development-yellow" alt="Development"/>
</p>

<p align="center">
  <em>"The creativity is in the constraint design, not the output generation."</em>
</p>

---

## What is Fernwood?

Fernwood is a lightweight AI coding harness that works directly in your codebase. Built on PicoClaw's agent loop, it adds Unix-style command execution and Mulch-powered persistent knowledge.

**Core Philosophy:**
- Coding-focused — Built specifically for software development workflows
- Local-first — Runs on your machine, respects your workspace
- Always learning — Proactively records knowledge via Mulch
- Lightweight — Single binary, fast startup, low memory footprint
- Unix-style — Command chaining, pipes, and shell auto-fallback

---

## Quick Start

### 1. Initialize

```bash
./fernwood onboard
```
This creates your workspace at `~/.fernwood/workspace` and generates a default config.

### 2. Configure

Edit `~/.fernwood/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.fernwood/workspace",
      "model_name": "claude-sonnet-4-5",
      "max_tokens": 8192,
      "temperature": 0.5,
      "max_tool_iterations": 30
    }
  },
  "model_list": [
    {
      "model_name": "claude-sonnet-4-5",
      "model": "anthropic/claude-sonnet-4-5",
      "api_key": "YOUR_ANTHROPIC_API_KEY"
    }
  ]
}
```

### 3. Run

```bash
# Interactive session
./fernwood agent

# Single task
./fernwood agent -m "refactor the auth module to use interfaces"
```

---

## Mulch Setup (Optional)

Fernwood integrates with [Mulch](https://github.com/jayminwest/mulch) for persistent memory across sessions. To enable:

**1. Install Mulch CLI**

```bash
bun install -g @os-eco/mulch-cli
```

**2. Initialize Mulch in your workspace**

```bash
cd ~/.fernwood/workspace
mulch init
```

**3. Enable in config**

Edit `~/.fernwood/config.json`:

```json
{
  "mulch": {
    "enabled": true,
    "bin": "mulch"
  }
}
```

Once enabled, Fernwood will automatically record knowledge during conversations and reflect before session clears.

---

## Tool Architecture

Fernwood provides **3 core tools** plus optional integrations:

### Core Tools

| Tool | Purpose |
|------|---------|
| **`run`** | Unified command execution (primary tool) |
| **`edit_file`** | Surgical file edits (preferred for code changes) |
| **`append_file`** | Append content to files |

### The `run` Tool

The `run` tool is your primary interface, supporting Unix-style command chaining (`|`, `&&`, `||`, `;`) and automatic shell fallback.

**File operations:** `cat`, `ls`, `write`, `grep`, `head`, `tail`, `wc`, `stat`, `rm`, `cp`, `mv`, `mkdir`

**Memory (Mulch):** `memory record`, `memory facts`, `memory search`, `memory query`, `memory forget`, `memory status`

**Topic management:** `topic list`, `topic info`, `topic runs`, `topic run`, `topic rename`, `topic search`

**Skills:** `skill search`, `skill install`, `skill list`, `skill info`, `skill update`, `skill uninstall`

**Shell auto-fallback:** Any unknown command executes via shell (`git`, `python3`, `sed`, etc.)

*The `run` tool is based on [agent-clip](https://github.com/epiral/agent-clip).*

### Optional Tools

Enable via `~/.fernwood/config.json`:

- **Web search** — Multiple search backends
- **Web fetch** — Fetch and parse web pages
- **Message** — Send messages to channels (Discord, Matrix)
- **Send file** — Share files via channels
- **Skills** — Discover and install skills from registries
- **Subagent** — Spawn specialized agents for complex tasks
- **MCP** — Model Context Protocol server integrations

---

## Mulch Memory — Always Be Learning

Fernwood integrates with [Mulch](https://github.com/jayminwest/mulch) for persistent inter-session expertise. Knowledge is organized by domain and grows over time.

### How It Works

1. **Proactive Recording** — Agent will record knowledge during conversations

2. **Session-End Reflection** — Before `/clear`, Fernwood automatically reflects and records valuable learnings.

3. **On-Demand Retrieval** — Domain summaries load at startup; full content retrieved via `memory query <domain>`, so when discussing a topic Fernwood will remember key points.

*Powered by [Mulch](https://github.com/jayminwest/mulch) — persistent expertise management.* 


## Current Versions

| Binary | Architecture | Size | Use Case |
|--------|--------------|------|----------|
| `fernwood-linux-amd64` | x86_64 | ~17MB | Desktop, server, cloud VMs |
| `fernwood-linux-arm64` | ARM64 | ~16MB | Raspberry Pi, Apple Silicon, AWS Graviton |

---

## Credits

**PicoClaw** — [sipeed/picoclaw](https://github.com/sipeed/picoclaw)  
Fernwood is a fork of PicoClaw. The agent loop, base tool infrastructure, provider routing, and channel integrations are substantially their work.

**Mulch** — [jayminwest/mulch](https://github.com/jayminwest/mulch)  
Persistent expertise management and domain-based knowledge storage.

**agent-clip** — [epiral/agent-clip](https://github.com/epiral/agent-clip)  
The `run` tool's Unix-style command execution and chaining semantics.

---

## License

MIT — see [LICENSE](LICENSE).

---

**Last Updated**: 2026-03-13  
**Version**: Development (vdev)
