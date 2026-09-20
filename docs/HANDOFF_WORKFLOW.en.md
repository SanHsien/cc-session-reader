# Universal Two-Stage Multi-Agent Handoff Workflow

[繁體中文](HANDOFF_WORKFLOW.md) | [English](HANDOFF_WORKFLOW.en.md)

To prevent confusion and eliminate complex command-line arguments, cross-agent workflows are standardized into two clear steps: **Stage 1: Summarize & Export** and **Stage 2: Takeover & Execute**.

---

## Stage 1: Summarize & Export

When you complete planning in **Claude Desktop** or **Claude Code**, or reach the **5-hour rate limit**, prompt:

> **"Please summarize the current session and generate HANDOFF.md for the next agent."**

### Agent Automated Workflow:
1. Calls `cc-session` to extract clean context without raw JSON/tool noise (80%+ reduction).
2. Writes structured handoff card `HANDOFF.md` to the project root:
   ```markdown
   # Project Handoff Summary (HANDOFF.md)
   - **Source Session**: [Session ID & Timestamp]
   - **Core Mission Goal**: [One-sentence task goal]
   - **Completed Progress**:
     - [x] [Completed architecture or code]
     - [x] [Passed tests]
   - **Technical Constraints & Decisions**: [Decisions to avoid circular work]
   - **Handoff Trigger**: [e.g. 5h Rate Limit / Plan ready for execution]
   - **Actionable Next Steps**:
     1. [Specific files to modify or verification commands]
   ```

---

## Stage 2: Takeover & Execute

Switch to the downstream agent (**Codex Desktop, Cursor, Antigravity, Hermes, or a fresh Claude session**) without opening a terminal, and paste the unified prompt:

### 1. Codex Desktop / Antigravity (opencodex)
> **"Please read HANDOFF.md and continue with the next steps."**
*(Codex loads the structured handoff, generates code, and runs local Quality Gates)*

### 2. Cursor (Composer / Chat)
> **"@HANDOFF.md Please read this summary and implement the next steps."**
*(Cursor reads the refined summary and performs inline edits in the IDE)*

### 3. Hermes / CLI Agents
> **"Read HANDOFF.md and continue with the next action items."**

### 4. Claude Desktop / Code (Return after quota reset)
> **"Please read HANDOFF.md and review previous decisions."**

---

## Core Benefits
- **Zero Mental Overhead**: Always ask the source agent to generate `HANDOFF.md`, and ask the target agent to read `HANDOFF.md`.
- **Zero Token Waste**: The handoff artifact is statically filtered and curated, eliminating hundreds of thousands of tokens of transcript bloat.
