# Standard Two-Stage Multi-Agent Handoff Workflow

[繁體中文](HANDOFF_WORKFLOW.md) | [English](HANDOFF_WORKFLOW.en.md)

This workflow requires only two prompts. Claude, Codex, Cursor, Antigravity, Hermes, or another agent can be either the source or the receiver.

```text
Source agent --summarize--> HANDOFF.md --verify--> Target agent --continue--> updated HANDOFF.md
```

## Stage 1: summarize

Unified prompt:

> Please summarize the current work and create HANDOFF.md in the project root for the next agent.

The source agent should:

1. Read project rules and verify the current branch, HEAD, dirty files, and test status.
2. Produce `HANDOFF.md` from the current conversation and workspace rather than copying a chat summary alone.
3. Record the objective, authorized scope, completed work and evidence, decisions, remaining work, risks, next steps, and verification commands.
4. Mark unknown or irreproducible facts as “needs verification.” Never include passwords, tokens, or complete raw logs.

Recommended format:

```markdown
# HANDOFF

- Generated at:
- Source agent/session:
- Repository/branch/HEAD:
- Workspace state:

## Objective and scope
## Completed work and evidence
## Decisions and constraints
## Remaining work, risks, and blockers
## Next steps in priority order
## Verification run and still pending
```

### Importing a historical Claude Code session

Use the parser only when the historical session is outside the current context:

1. Run `cc-session list` to identify the session.
2. Repeat `cc-session inherit <id>` until `[inherit complete]` appears.
3. Verify the current workspace, then write relevant information to `HANDOFF.md`.

`cc-session` reads local Claude Code JSONL files. It is not a universal session database for Codex, Cursor, Antigravity, or Hermes. Other source agents summarize from their current context.

## Stage 2: take over

Unified prompt:

> Please read HANDOFF.md in the project root, verify the current state, take over, and continue with the next step.

The receiving agent should:

1. Read project rules before `HANDOFF.md`.
2. Verify the branch, HEAD, dirty files, relevant files, and test state. The current workspace and newer instructions override stale handoff content.
3. Briefly report the takeover state and any differences, then execute the first applicable next step.
4. Update `HANDOFF.md` when pausing or completing work so another agent can take over again.

## Interface differences

Keep the prompt unchanged. If an interface requires explicit file selection, such as some Cursor modes, attach or mention `@HANDOFF.md`. That is an interface detail, not a separate workflow.
