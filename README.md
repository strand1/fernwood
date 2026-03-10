<p align="center">
  <img src="assets/fernwood_logo.png" alt="Fernwood" width="200"/>
</p>

<h1 align="center">Fernwood 🌲</h1>

<p align="center">
  <strong>A focused coding agent for your terminal.</strong><br/>
  Local-first. Single binary. Persistent memory via <a href="https://github.com/jayminwest/mulch">Mulch</a>.
</p>

<p align="center">
  <a href="https://golang.org/dl/"><img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"/></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green" alt="MIT License"/></a>
  <img src="https://img.shields.io/badge/status-alpha-orange" alt="Alpha"/>
</p>

---

Fernwood is a lightweight agentic coding harness forked from [PicoClaw](https://github.com/sipeed/picoclaw). It strips the hardware bot and chat app integrations down to what matters for software development: a fast, local coding agent with four sharp tools and memory that persists across sessions.

**What's different from PicoClaw:**
- Stripped to a coding-focused toolset (`read_file`, `write_file`, `edit_file`, `bash`)
- Agent identity and system prompt rewritten for software development
- [Mulch](https://github.com/jayminwest/mulch) integration — expertise accumulates across sessions
- No Telegram, no WeChat, no hardware channels (Discord optional)
- Dependencies halved (~53 vs ~106 direct deps)

**What's inherited from PicoClaw:**
- Single self-contained Go binary
- Fast startup, low memory footprint
- Solid tool loop and agent orchestration
- Discord channel (still works if you want it)

---

## Quick Start

**1. Build**

```bash
git clone https://github.com/strand1/fernwood.git
cd fernwood
make build
```

**2. Initialize**

```bash
./build/fernwood onboard
```

This creates `~/.fernwood/config.json` and the workspace at `~/.fernwood/workspace`.

**3. Configure**

Edit `~/.fernwood/config.json` and set your API key:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.fernwood/workspace",
      "model_name": "claude-sonnet-4-6",
      "max_tokens": 8192,
      "temperature": 0.5,
      "max_tool_iterations": 30
    }
  },
  "model_list": [
    {
      "model_name": "claude-sonnet-4-6",
      "model": "anthropic/claude-sonnet-4-6",
      "api_key": "YOUR_ANTHROPIC_API_KEY"
    }
  ]
}
```

Any OpenAI-compatible provider works. See `config/config.example.json` for the full schema.

**4. Run**

```bash
# Single task
./build/fernwood agent -m "refactor the auth module to use interfaces"

# Interactive
./build/fernwood agent
```

---

## Tools

Fernwood gives the agent four tools:

| Tool | Description |
|------|-------------|
| `read_file` | Read a file or list the project tree. Call with no args to see the tree. |
| `write_file` | Create a new file. Use only for files that don't exist yet. |
| `edit_file` | Surgically replace a string in an existing file. Fails loudly on ambiguous matches. |
| `bash` | Run shell commands. Used for tests, builds, git operations. |

The agent is instructed to read before writing, make the smallest change that works, and run tests after any code change.

---

## Mulch Memory

Fernwood integrates with [Mulch](https://github.com/jayminwest/mulch) for persistent inter-session expertise. If `mulch` is installed and in your PATH, Fernwood will:

- Inject relevant expertise into the system prompt at session start (`mulch prime`)
- Record new learnings at session end (`mulch record`)

Configure in `~/.fernwood/config.json`:

```json
{
  "mulch": {
    "enabled": true,
    "bin": "mulch",
    "auto_record": true,
    "domains": ["code", "errors", "decisions"]
  }
}
```

With Mulch disabled, Fernwood works fine — sessions just don't accumulate long-term memory.

---

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `FERNWOOD_HOME` | Root directory for Fernwood data | `~/.fernwood` |
| `FERNWOOD_CONFIG` | Path to config file | `$FERNWOOD_HOME/config.json` |

**Workspace layout:**

```
~/.fernwood/workspace/
├── memory/
│   └── MEMORY.md       # Long-term memory
├── sessions/           # Conversation history
├── skills/             # Custom skill files
└── AGENTS.md           # Agent behavior guide
```

---

## Discord (Optional)

Fernwood keeps PicoClaw's Discord channel. To enable:

```json
{
  "channels": {
    "discord": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    }
  }
}
```

Then run `./build/fernwood gateway`.

---

## CLI Reference

```
fernwood onboard          Initialize config and workspace
fernwood agent            Interactive session
fernwood agent -m "..."   Single-task mode
fernwood gateway          Start Discord gateway
fernwood status           Show current status
```

---

## Roadmap

- [ ] Bubble Tea TUI — three-panel layout (agent log / tool calls / input)
- [ ] Session persistence and `--resume`
- [ ] Mulch auto-record at session end
- [ ] `FERNWOOD_*` env var rename (currently inherits some `PICOCLAW_*` names)
- [ ] Remove Antigravity OAuth code
- [ ] `make release` with Linux + macOS binaries

---

## Credits

Fernwood is a fork of [PicoClaw](https://github.com/sipeed/picoclaw) by [Sipeed](https://github.com/sipeed). The agent loop, tool infrastructure, provider routing, and Discord channel are substantially their work. PicoClaw is an impressive piece of engineering — Fernwood just points it at a different problem.

Mulch is by [Jaymin West](https://github.com/jayminwest).

---

## License

MIT — see [LICENSE](LICENSE).
