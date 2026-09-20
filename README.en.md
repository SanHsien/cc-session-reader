# cc-session-reader (Windows & Codex Desktop Fork)

[![CI](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml/badge.svg)](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20x64%20%7C%20arm64-blue.svg)](#)

[~AÔN-e](README.md) | [English](README.en.md)

**cc-session-reader** is a high-efficiency transcript extractor and session handoff tool tailored for Windows and AI-assisted development.
It statically parses local Claude Code and Claude Desktop session files (`.jsonl`), achieving **80% 88%** token reduction by stripping heavy harness frames and raw tool outputs while fully preserving user and assistant reasoning.

This fork provides first-class support for **Codex Desktop**, enabling seamless natural-language session handoffs without opening a command prompt!

---

## Highlights

1. **Zero-LLM Static Extraction**:
   - Compresses 300K+ token transcripts down to 30K 50K tokens with a fast Go binary.
2. **No Terminal Required (GUI Friendly)**:
   - Bundles a ready-to-use Codex skill (`skills/claude-handoff`).
   - Simply prompt in Codex Desktop: *"Hand over from Claude's recent session"*, and Codex takes care of the rest.
3. **Windows-First**:
   - Streamlined specifically for Windows 11 + PowerShell environments.

---

## Quick Installation (Windows PowerShell)

Run the following one-liner in PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/SanHsien/cc-session-reader/main/install.ps1 | iex"
```

---

## Desktop Natural Language Handoff

When Claude Desktop hits the 5-hour rate limit:
1. Switch to **Codex Desktop**.
2. Type naturally in the prompt box:
   - *"Take over the recent Claude session"*
   - *"Hand over Claude's latest discussion on [feature/project]"*
   - *"Read Claude session <session-id>"*
3. Codex automatically extracts the clean context in the background and resumes coding immediately!

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

Licensed under the Apache License, Version 2.0. Original work Copyright 2026 Mapleÿ. See [NOTICE.md](NOTICE.md) for details.
