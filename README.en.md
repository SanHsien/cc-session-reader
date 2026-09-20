# cc-session-reader (Windows & Multi-Agent Handoff Fork)

[![CI](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml/badge.svg)](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20x64%20%7C%20arm64-blue.svg)](#)

[~AÔN-e](README.md) | [English](README.en.md)

**cc-session-reader** is a high-efficiency transcript extractor, static compressor, and cross-agent handoff tool tailored for Windows and multi-agent AI development workflows.
It parses local Claude Code and Claude Desktop session files (`.jsonl`), achieving **80% 88%** token reduction via a fast Go binary by stripping heavy harness frames and raw tool outputs while fully preserving essential user/assistant reasoning.

This fork is expanded for **Universal Multi-Agent Summarization & Handoff**: when Claude Desktop reaches its 5-hour rate limit or finishes high-level planning, subsequent agents such as **Codex, Cursor, Antigravity, and Hermes** can take over seamlessly without token bloat!

---

## Why Multi-Agent Handoff?

Modern AI developers leverage diverse agents for their unique strengths:
- **Claude Desktop / Claude Code**: Elite high-level architecture, deep planning, and comprehensive documentation (often throttled by 5-hour rate limits).
- **Codex Desktop (with opencodex / Antigravity quota)**: High-throughput code generation, refactoring, and local Quality Gates.
- **Cursor**: In-IDE inline coding, contextual diffing, and fast multi-file navigation.
- **Hermes & CLI Agents**: Autonomous background scripts and standalone tool flows.

Manually copying transcripts burns hundreds of thousands of tokens on raw JSON and noise. **cc-session-reader serves as the clean bridge between agents.**

---

## Cross-Agent Handoff Matrix

| Target Agent | Handoff Method | How It Works |
|---|---|---|
| **Codex Desktop** | In chat, prompt:<br>Ø=ÜI *"Take over the recent Claude session"*<br>Ø=ÜI *"Hand over Claude's latest discussion on [feature]"* | The built-in skill invokes `cc-session` in the background and injects clean context. |
| **Cursor** | Generates `.cursor/handoff.md`, then in Composer/Chat:<br>Ø=ÜI `@handoff.md Continue with next steps` | Cursor reads the refined goals, completed tasks, and actionable next steps cleanly. |
| **Antigravity** | In opencodex session / chat prompt:<br>Ø=ÜI *"Read session <id> and take over development"* | Inherits clean reasoning to maximize high-throughput execution with models like Gemini 3.8 Flash. |
| **Hermes / Other CLI** | Reads generated `HANDOFF.md` | Standard structured Markdown card compatible with any CLI agent. |
| **Claude Desktop / Code** | In a fresh session:<br>Ø=ÜI `/cc-session inherit <id>` | Paginates through prior session history safely. |

---

## Quick Installation (Windows PowerShell)

Run the following one-liner in PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/SanHsien/cc-session-reader/main/install.ps1 | iex"
```

This automatically:
1. Downloads the latest Windows binary `cc-session.exe` to `$env:LOCALAPPDATA\cc-session\`.
2. Sets up ready-to-use skills in `~/.codex/skills/claude-handoff` and `~/.claude/skills/cc-session`.

---

## Structured Handoff Artifact Format

When generating a session summary for handoff, the tool formats it as:

```markdown
# Agent Handoff Summary
- Source Session: 33b4ebf5 (2026-09-20 10:56)
- Workspace: C:\Users\...\my-project
- Goal: [Original task goal]
- Progress:
  - [x] Completed architecture and data models
  - [x] Passed local sanity checks
- Key Constraints: [Decisions to avoid circular work]
- Handoff Trigger: [e.g. 5h Rate Limit / Plan ready for execution]
- Actionable Next Steps:
  1. Modify src/components/NavBar.tsx
  2. Run dev_check.ps1 to verify
```

---

## CLI Reference

| Command | Description | Example |
|---|---|---|
| `list` | List recent sessions | `cc-session list -n 10` |
| `context` | Emit compact context format | `cc-session context <id>` |
| `inherit` | Paged session inheritance | `cc-session inherit <id>` |
| `read` | Full conversation with one-line tool summaries | `cc-session read <id>` |
| `stats` | Character and token distribution stats (80%+ reduction) | `cc-session stats <id>` |

---

## Governance & Verification

```powershell
# Run the canonical Windows verification gate
powershell -NoProfile -File tools/dev_check.ps1
```

## License & Attribution

Licensed under Apache License, Version 2.0. Original work Copyright 2026 Mapleÿ. See [NOTICE.md](NOTICE.md) for details.
