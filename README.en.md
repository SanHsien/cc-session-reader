# cc-session-reader (Windows & Multi-Agent Handoff Fork)

[![CI](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml/badge.svg)](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20x64%20%7C%20arm64-blue.svg)](#)

[~AÔN-e](README.md) | [English](README.en.md)

**cc-session-reader** is a high-efficiency transcript extractor, static compressor, and cross-agent handoff tool tailored for Windows and multi-agent AI development workflows.
It parses local Claude Code and Claude Desktop session files (`.jsonl`), achieving **80% 88%** token reduction via a fast Go binary by stripping heavy harness frames and raw tool outputs while fully preserving essential user/assistant reasoning.

---

## Unified Two-Stage Handoff Workflow

To eliminate confusion across different tools, cross-agent handoffs are standardized into two clear phases with uniform prompts:

```text
      [Stage 1: Source (Claude)]
       In Claude Desktop / Code prompt:
       Ø=ÜI "Please summarize the current session and generate HANDOFF.md for the next agent."
                          %
                          % (cc-session statically compresses and outputs HANDOFF.md)
                          %¼
                  [Project Root: HANDOFF.md]
                          %
                          % Zero token waste / Structured
                          %¼
      [Stage 2: Target (Codex / Cursor / Antigravity / Hermes)]
       In target agent prompt:
       Ø=ÜI Codex / Antigravity : "Please read HANDOFF.md and continue with the next steps."
       Ø=ÜI Cursor               : "@HANDOFF.md Please read this summary and implement the next steps."
       Ø=ÜI Hermes / CLI agents  : "Read HANDOFF.md and continue with the next action items."
```

---

## Standardized Two-Stage Prompt Reference

### Stage 1: Summarize & Export

When you finish planning in **Claude Desktop / Claude Code**, or hit the **5-hour rate limit**, enter:

> **"Please summarize the current session and generate HANDOFF.md for the next agent."**

#### Standard Handoff Card Format (`HANDOFF.md`):
```markdown
# Project Handoff Summary (HANDOFF.md)
- **Source Session**: [Session ID & Timestamp]
- **Core Mission Goal**: [One-sentence task objective]
- **Completed Progress**:
  - [x] [Completed architecture or code]
  - [x] [Passed tests]
- **Technical Constraints & Decisions**: [Decisions to avoid circular work]
- **Handoff Trigger**: [e.g. 5h Rate Limit / Plan ready for execution]
- **Actionable Next Steps**:
  1. [Specific files to modify or verification commands]
```

---

### Stage 2: Takeover & Execute

Switch to the target agent and paste the single unified prompt into its chat box without opening a terminal:

| Target Agent | Standardized Prompt | Action Taken |
|---|---|---|
| **Codex Desktop** | **"Please read HANDOFF.md and continue with the next steps."** | Ingests clean context, writes code, and runs local Quality Gates. |
| **Cursor** | **"@HANDOFF.md Please read this summary and implement the next steps."** | Cursor reads the refined summary and performs inline edits in the IDE. |
| **Antigravity** | **"Please read HANDOFF.md and continue with the next steps."** | Maximizes high-throughput execution with models like Gemini 3.8 Flash. |
| **Hermes / CLI** | **"Read HANDOFF.md and continue with the next action items."** | CLI agents autonomously execute remaining tasks. |
| **Claude (Return)**| **"Please read HANDOFF.md and review previous decisions."** | Resumes after quota reset with zero context bloat. |

---

## Quick Installation (Windows PowerShell)

Run the following one-liner in PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/SanHsien/cc-session-reader/main/install.ps1 | iex"
```

---

## CLI Reference

| Command | Description | Example |
|---|---|---|
| `list` | List recent sessions | `cc-session list -n 10` |
| `context` | Emit compact context format | `cc-session context <id>` |
| `inherit` | Paged session inheritance | `cc-session inherit <id>` |
| `read` | Full conversation with one-line tool summaries | `cc-session read <id>` |
| `stats` | Character and token distribution stats | `cc-session stats <id>` |

---

## Governance & Verification

```powershell
# Run the canonical Windows verification gate
powershell -NoProfile -File tools/dev_check.ps1
```

## License & Attribution

Licensed under Apache License, Version 2.0. Original work Copyright 2026 Mapleÿ. See [NOTICE.md](NOTICE.md) for details.
