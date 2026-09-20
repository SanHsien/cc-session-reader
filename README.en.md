# cc-session-reader (Windows Multi-Agent Handoff Fork)

[![CI](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml/badge.svg)](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20x64%20%7C%20arm64-blue.svg)](#)

[繁體中文](README.md) | [English](README.en.md)

This is a Windows-first maintained fork of [`Mapleeeeeeeeeee/cc-session-reader`](https://github.com/Mapleeeeeeeeeee/cc-session-reader). It preserves the upstream Go parser for filtering Claude Code JSONL transcripts and importing historical sessions, and adds an `agent-handoff` Skill so Claude, Codex, Cursor, Antigravity, Hermes, and other agents can continue from the same `HANDOFF.md`.

```text
Any source agent
  │  Stage 1: summarize the current state
  ▼
HANDOFF.md in the project root
  │  Stage 2: verify and take over
  ▼
Claude / Codex / Cursor / Antigravity / Hermes / other agents
```

## Unified workflow

Every agent uses the same two prompts. There are no agent-specific commands to remember.

### Stage 1: summarize

> Please summarize the current work and create HANDOFF.md in the project root for the next agent.

The source agent records the objective, completed work, branch/HEAD, verification evidence, decisions, risks, and actionable next steps from the current conversation and workspace. A normal summary does not require a session ID or CLI command.

Use `cc-session list` and `cc-session inherit <id>` only when importing a historical Claude Code session that is not already in the current context. The agent then updates `HANDOFF.md` from the filtered history.

### Stage 2: take over

> Please read HANDOFF.md in the project root, verify the current state, take over, and continue with the next step.

The target agent reads project rules first and verifies the branch, HEAD, dirty files, and test status. If `HANDOFF.md` is stale, the current workspace and newer user instructions take precedence. In interfaces that require explicit file attachment, attach or mention `@HANDOFF.md` without changing the prompt.

See [Two-Stage Handoff Workflow](docs/HANDOFF_WORKFLOW.en.md) for the full specification.

## Skill responsibilities

| Skill | Purpose | When to use it |
|---|---|---|
| `agent-handoff` | Create, verify, and continue from `HANDOFF.md` | Routine handoffs between any source and target agents |
| `cc-session` | Read and compress historical Claude Code JSONL | Importing an old session that is outside the current context |

`agent-handoff` is not tied to a specific model or harness. It requires the receiving agent to verify the real workspace before trusting a potentially stale summary.

## Quick installation (Windows PowerShell)

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/SanHsien/cc-session-reader/main/install.ps1 | iex"
```

The installer:

1. Downloads the Windows `cc-session.exe` from this fork's latest release to `$env:LOCALAPPDATA\cc-session\`.
2. Installs the fork's `cc-session` and `agent-handoff` Skills for Claude Code.
3. Installs the fork's Codex Skill at `~/.codex/skills/agent-handoff`.
4. Creates `$env:LOCALAPPDATA\cc-session\agent-handoff-claude.zip`.

Claude Desktop custom Skills require a manual account-level upload at **Customize > Skills > + Create skill > Upload a skill**. A local script cannot perform that upload. See [Claude's official instructions](https://support.claude.com/en/articles/12512180-use-skills-in-claude).

## CLI reference (advanced)

| Command | Description | Example |
|---|---|---|
| `list` | List recent Claude Code sessions | `cc-session list -n 10` |
| `inherit` | Import the complete compressed history in pages | `cc-session inherit <id>` |
| `context` | Print compact context with metadata | `cc-session context <id>` |
| `read` | Print the conversation and tool summaries | `cc-session read <id>` |
| `expand` | Expand a specific tool call | `cc-session expand <id> <tool-id>` |
| `stats` | Show character and token distribution | `cc-session stats <id>` |

## Development and verification

```powershell
pwsh -NoProfile -File tools/dev_check.ps1
```

This fork maintains Windows amd64/arm64 only. Public upstream documentation has Traditional Chinese and English mirrors; fork-local governance documents may remain Traditional Chinese only.

## License and provenance

Licensed under the Apache License 2.0. See [NOTICE.md](NOTICE.md), [FORK.md](FORK.md), and [LICENSE](LICENSE) for upstream provenance and fork details.
