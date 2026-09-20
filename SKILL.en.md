---
name: cc-session
description: |
  Read historical Claude Code / Desktop sessions with the cc-session CLI instead of raw JSONL.
  Filters transcripts outside context, reducing 300K+ tokens to 30-50K by retaining only conversation and one-line tool summaries.
  Use when users want to review, reference, or analyze past sessions.
argument-hint: "[list | inherit <id> | read <id> | context <id> | expand <id> <tool-id> | stats <id> | audit <id> | usage]"
allowed-tools:
  - Bash
  - Read
---

[English](SKILL.en.md) | [繁體中文](SKILL.md)

## Routing

Decide what to execute based on `$ARGUMENTS` (the input following `/cc-session`):

| `$ARGUMENTS` | Action |
|------|------|
| Empty | Run `cc-session list`, present list to user, and ask which session to inspect |
| `list` (with optional `-p`, `-n`) | Run `cc-session $ARGUMENTS`, display session list |
| Bare session id (e.g. `16d06326`) | Read that session → proceed with the inherit pagination flow |
| Known subcommand + args (`inherit`/`read`/`context`/`expand`/`stats`/`audit`/`usage`) | Run `cc-session $ARGUMENTS`; if inherit, paginate |

## Reading Session Transcripts

- Use `inherit` to paginate full sessions into context cleanly (each page ≤28K chars).
- Repeatedly invoking `cc-session inherit <id>` automatically advances to the next page until `[inherit complete]`.
- Use `read` for quick scans or offset jumps.
