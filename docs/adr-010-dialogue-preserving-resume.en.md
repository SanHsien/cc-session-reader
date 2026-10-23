# ADR-010: Dialogue-Preserving Resume Protocol for Cache Expiration Recovery

[繁體中文](adr-010-dialogue-preserving-resume.md) | [English](adr-010-dialogue-preserving-resume.en.md)

**Status**: Implemented (2026-10-23).

## 1. Context and Core Problem

In long-running or multi-day agent workflows, developers face a recurring dilemma when prompt caches expire (typically after 1 hour of idle time) or when resuming across session boundaries:

1. **Over-abstracted `/compact` summaries lose critical precision**:
   - Native compaction produces generalized text summaries that frequently strip away specific edited file paths, exact error stack traces, fine-grained pending task lists, and user-specified architectural constraints.
   - When handling multi-file refactoring or complex codebases, resumed agents suffer from severe hallucination, repetitive edits, or regressive fixes.
2. **Raw JSONL transcript injection overwhelms context windows**:
   - Raw transcripts are flooded with redundant tool inputs, bulk file reads, and noisy runtime outputs (a single extended session easily grows to several megabytes and hundreds of thousands of tokens).
   - Injecting uncurated JSONL transcripts exhausts context limits and incurs excessive API costs.

## 2. Benchmark and Empirical Compression Measurements

Evaluated against an actual extended production session `6b6c9288` (1,235 lines, 212 model API turns, 145 Bash invocations, cross-task dispatches):

- **Raw Transcript Size**: `8,554.6 KB`
- **Last Turn Model Context**: `493,971 tokens` (~500k context)
- **Character Compression Metrics (`cc-session stats`)**:
  - Raw Characters: `495,435`
  - Core Filtered Characters: `73,741`
  - Noise Eliminated: **`421,694 (85.1% reduction)`**
- **Content Breakdown**:
  - Human Prompts (`user text`): `11,245 chars` (100% preserved)
  - Assistant Responses (`assistant text`): `18,686 chars` (100% preserved)
  - Tool Invocation Summaries (`tool summaries`): `39,438 chars` (concise single-line operation & status)
  - Pruned Raw Payloads (`tool input/result raw`): `456,905 chars` (massive file content and bulk stdout stripped)

## 3. The Dialogue-Preserving Resume Protocol

The protocol defines a standardized 3-step operational procedure for new agent sessions to inherit prior context losslessly and economically:

### Step 1: Identify the Interrupted Session
```bash
cc-session list -n 5
```
Inspect recent session IDs, timestamps, and closing intent to locate the resumption point.

### Step 2: Validate Volume and Compression Ratio
```bash
cc-session stats <session-id>
```
Confirm the noise reduction ratio and retained characters to ensure prior decisions remain intact.

### Step 3: Paged Lossless Context Inheritance
```bash
cc-session inherit <session-id> -reset
```
Execute repeatedly in the fresh agent session until `[inherit complete]` is reached.
- Each page is strictly bounded to 28,000 characters, safeguarding single-turn prompt throughput.
- Injected records retain wall-clock headers `[HH:MM:SS]`, date separators `--- YYYY-MM-DD ---`, and explicit role attribution (`user:`, `assistant:`, `harness:`).
- Preserves full reasoning continuity and architectural choices, completely resolving compaction amnesia.

## 4. Comparative Evaluation

| Evaluation Metric | Native `/compact` | Raw JSONL Transcript | `cc-session` Resume Protocol |
| :--- | :--- | :--- | :--- |
| **Token Overhead** | Minimal (3k~5k tokens) | Extreme (400k+ tokens, blowout risk) | **Optimal balance (20k~30k tokens)** |
| **Critical Path Retention** | Low (frequently omitted) | 100% (drowned in raw noise) | **100% (distilled to one-line summaries)** |
| **Pending Task Tracking** | Vague generalization | Requires manual fulltext digging | **Faithfully preserved to the final turn** |
| **Cross-Agent Portability** | Claude Code only | Custom ad-hoc scripts | **Universal across Claude, Codex, Antigravity** |
