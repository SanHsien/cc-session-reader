# ADR-008: Track Harness Event-Format Drift and Separate Turn Counting from Character Statistics

[繁體中文](adr-008-harness-event-drift.md) | [English](adr-008-harness-event-drift.en.md)

**Status:** Implemented.

The reader recognizes Claude Code transcript events by matching literal strings. Those strings change, and several already had. An audit of 120 transcripts from the latest 60 days (35,940 entries) found three categories of drift. This ADR records six decisions.

| # | Decision | Impact |
|---|---|---|
| 1 | Add eight later CLI entry types to `noiseTypes` | 597 KB moves from “not parsed” to `system_noise` |
| 2 | Classify six harness-injected user-message forms in the parser | They no longer impersonate user messages |
| 3 | Add a `harness:` role label | Readers can distinguish typed text from injected text |
| 4 | Detect teammate messages only by XML tags and accept both variants | Prose rewrites no longer break detection |
| 5 | Add `UserMessage.CountsAsTurn()` and calibrate it from measurements | Aggregate turn-count error across eight sessions falls from 70 to 29 |
| 6 | Detect skill injection through the `sourceToolUseID` structural link | Bundled skills without the base-directory line are no longer rendered in full |

The original measurement sampled the latest 120 `.jsonl` files by modification time under `~/.claude/projects`, covering the 60 days ending 2026-08-30.

**Additional sample, 2026-09-02:** The first audit skipped `<session>/subagents/*.jsonl`. Adding the 120 largest subagent transcripts found five cases handled by extending existing decisions:

- `relocated` entry type: 600 entries / 82 KB, added to the decision-1 allowlist.
- A prose warning before `<task-notification>` in CLI 2.1.200–2.1.235: detection now recognizes the tag anywhere rather than requiring it at the start, using decision 4's “tags, not prose” rationale.
- `The coordinator sent a message while you were working:`: three entries. A coordinator starts work for a subagent, so it is harness content and counts as a turn under decisions 2 and 5.
- `<local-command-stderr>`: two entries, handled like the existing stdout variant.
- `Continue from where you left off.`: two top-level `isMeta: true` entries. They are tails of invocations started elsewhere, so they do not count as turns, like stop-hook notifications.

Scanning the combined main-session and subagent corpus found three more forms:

- `<fork-boilerplate>...</fork-boilerplate>`: five subagent entries of about 1,000 characters. Compress the fixed preamble to `[fork]`, preserve the actual instruction after the closing tag, and count it as a turn because it starts the fork's work.
- `[Your previous response had no visible output. Please continue and produce a user-visible response.]`: 36 exact main-session messages. Compress to `[nudge: no visible output]` and count it as a turn because it requests a new response.
- `The user sent a new message while you were working:` plus explanatory wrapper: 35 entries. Preserve the human body with a `user:` label, strip the harness wrapper, and do not count a new turn because the wrapper explicitly says it arrived during an existing turn.

## 1. `noiseTypes` missed eight entry types added later by the CLI

`noiseTypes` is a handwritten allowlist containing 13 types. Eight observed types were missing:

| Type | Count | Bytes |
|---|---:|---:|
| `atis-latch` | 1,600 | 191,680 |
| `frame-link` | 664 | 96,167 |
| `worktree-state` | 285 | 115,708 |
| `file-history-delta` | 168 | 80,618 |
| `artifact-autoreact-ledger` | 159 | 80,936 |
| `artifact-comment-monitor` | 49 | 17,522 |
| `agent-setting` | 37 | 3,947 |
| `cost-state` | 14 | 10,608 |
| Total | 2,976 | 597,186 |

These types have no `message`. In `parseLineWithToolNames`, `raw.Message == nil`; if the type is not allowlisted, the line returns no event. Filtered output is therefore already correct, but statistics are wrong: `EventNoise` contributes characters to `system_noise`, while fall-through makes these 597 KB invisible.

Decision: add the types to the allowlist. A reverse allowlist was rejected because a new type containing real conversation content could then be silently discarded, a more dangerous failure than undercounted statistics. `cost-state` contains `totalCostUSD` and `modelUsage`; this ADR classifies it as noise but does not consume it.

## 2. Six harness-injected user-message forms were not classified

`classifyHarnessUserMessage` already recognized system reminders, skill injection, teammate messages, context usage, and command injection. Six observed forms were missing:

| Prefix | Count | Characters | Mean |
|---|---:|---:|---:|
| `<task-notification>` | 64 | 147,340 | 2,302 |
| `This session is being continued…` | 13 | 190,493 | 14,653 |
| `[Request interrupted by user` | 13 | 403 | 31 |
| `N background agents were stopped…` | 4 | 1,605 | 401 |
| `A session-scoped Stop hook is now active` | 2 | 1,034 | 517 |
| `Skill /X was loaded earlier…` | 2 | 357 | 178 |

`<task-notification>` was a special inconsistency: the formatter recognized and compressed it to `[summary]`, but the parser did not. Detection moved to the parser so formatter and statistics branch on domain fields rather than rematching tags.

Compact forms follow the existing `[kind: id]` convention:

| Message | Rendering | Characters |
|---|---|---:|
| Stop hook | `[goal] <condition>` | 517 → 50 |
| Agents stopped | `[agents stopped: 7]` | 401 → 19 |
| Repeated skill load | `[skill: X] (repeat)` | 178 → 33 |
| Request interrupted | `[interrupted]` | 31 → 13 |
| Compaction continuation | `[compaction summary]` plus full body | 14,653 → 14,320 |
| Task notification | Existing compression, new role only | unchanged |

The compaction continuation body is preserved because it is exactly the prior-conversation summary needed by the inheriting reader. Only the harness preamble and “read the full transcript at” footer are removed.

This change is not primarily about token savings: all six forms save only about 7,300 characters across 60 days. Its value is accurate role labeling and turn counting.

## 3. Add a `harness:` role label

Previously every user-role entry rendered as `user:`. In session `b11858cf`, 46 entries were labeled `user:`, but ten were harness injections. After the change, the session has 36 `user:` and ten `harness:` entries with the same total. The compact `context` prefix similarly separates `H:` from `U:`.

`system:` was rejected because transcript entry types already use that word, and `hook:` was too narrow. `harness:` describes all six forms without overloading an existing concept.

## 4. The primary teammate-message check was dead code

```go
teammateWarning = "IMPORTANT: This is NOT from your user"
```

Among 133 teammate messages in 60 days, this string matched zero. Detection survived only because the fallback preamble `Another Claude session sent a message:` matched 133 of 133. One more prose rewrite would have turned every teammate message into a normal user message.

Even matching several prose markers was insufficient: among 1,112 teammate messages, 80 used `<agent-message>` instead of `<teammate-message>` and lacked both the preamble and disclaimer; three sampled entries had no attributes.

Decision: detect the XML tags themselves, not prose. Accept `<teammate-message>` with `teammate_id` and `<agent-message>` with optional `from`. Store variants once in `session.TeammateTagVariants`, shared by `classify.go` and `CompactTeammateMessage`, so the two paths cannot drift independently. Prose is written for humans and changes often; structural tags are part of the harness protocol and cost more to change.

The old warning-removal code in `CompactTeammateMessage` is also dead but harmless, because extraction already stops at the closing tag. It was left unchanged.

## 5. One flag answered two unrelated questions

The user branch of `ComputeStats` both classified characters and counted turns. Classification used early `continue` statements, while `userTurnCount++` sat at the bottom of the loop. Every early exit written for character statistics therefore also skipped turn counting. That is correct for system reminders and command output, but wrong for teammate messages.

K equals API calls divided by turns. `cost.go` uses `extraCallsPerTurn(K) = K - 1` to estimate calls after the first call in each turn. A turn is therefore a unit of work started by an external prompt and ending when the agent stops. Teammate messages start exactly such units.

The decisive point is numerator/denominator consistency, not whether teammate messages resemble human text. `apiCallCount` includes every assistant message with distinct usage, including those caused by teammate messages. Before this change, K divided all API calls by only human-started turns. Session `fc65aeeb` became 165 calls / 4 turns = 41.2 even though it had 11 teammate messages.

Decision: put the policy in `UserMessage.CountsAsTurn()` and evaluate it before character-statistics branches.

### Measure which forms count rather than guessing

The first policy counted teammate messages but excluded all other harness injection. It performed worse than the old behavior. The revised policy measures whether an assistant message with usage occurs before the next true user message:

| Type | Started a turn | Did not | Share |
|---|---:|---:|---:|
| Teammate | 128 | 8 | 94% |
| Compaction continuation | 12 | 0 | 100% |
| Task notification | 63 | 8 | 89% |
| Stop hook | 2 | 0 | 100% |
| Repeated skill load | 3 | 0 | 100% |
| Interrupted | 0 | 14 | 0% |
| Agents stopped | 0 | 4 | 0% |

The metric is invalid for stop hooks and repeated skill loads because they are tails of invocations started by earlier entries. Teammate, compaction, and task-notification messages have no earlier user entry in the same turn.

| Message type | Rendering | Counts as turn |
|---|---|---|
| System reminder / context usage | Dropped | No |
| Command output / caveat | Dropped | No |
| Skill / command injection | Compressed | No; counting would duplicate the triggering user message |
| Stop hook / repeated skill load | Compressed | No; tail of an existing invocation |
| Interrupted / agents stopped | Compressed | No; measurements show no work starts |
| Teammate message | Compressed | **Yes** |
| Task notification | Compressed | **Yes** |
| Compaction continuation | Preserved | **Yes** |
| Normal message | Preserved | Yes |

### Calibration

Ground truth was approximated as each assistant message with usage that was not preceded by another assistant message. Across eight sessions, total absolute error was:

| Policy | Error |
|---|---:|
| Before the change | 70 |
| First version: teammate only | 81 |
| Adopted version | **29** |

Ground truth is itself heuristic, so the comparison is relative. In six sessions with similar 25–26% compression, cold-cache ten-turn savings fell as K rose from 2.8 to 8.6 (68% to 59%). Overestimating K therefore underestimates savings. Warm-cache results did not show the same trend; this is observational, not a controlled experiment.

## 6. Skill injection recognized only a text prefix

`classifyHarnessUserMessage` recognized skill injection only when text began with `Base directory for this skill:`. Bundled skills such as `artifact-design` omit that line, so their full body rendered as `user:`. Fifty such messages appeared in 60 days, about 12% of skill injections; one could be several KB and consumed one tenth of session `b11858cf`'s output.

These JSONL entries contain `isMeta: true` and a `sourceToolUseID` pointing to the preceding assistant `Skill` tool use, whose `input.skill` provides the name. `isMeta` alone is not sufficient because image placeholders and stop-hook feedback also carry it.

Decision: extend the reader's existing tool-use ID map to record Skill `skill` and `args`. User messages first follow that structural link; the textual prefix remains a stateless `ParseLine` fallback. Real transcripts showed the tool input name and path-derived name agree, and `seenSkills` deduplicates both paths.

This is another instance of decision 2: the harness already supplied a structural field while the reader still matched prose. Teammate messages lack an equivalent field, so decision 4 must continue matching structural tags in text.

## Not resolved

**Slash and bang commands still do not count as turns.** A slash command such as `/goal` starts work but appears as three entries: a command-name marker, command-message injection, and sometimes a skill body. A future decision must select exactly one representative entry. Bang commands do not start API turns and cannot share the same rule.

**`cost-state` is not consumed.** It contains `totalCostUSD` and `modelUsage`, while benchmark still reconstructs cost from per-message API usage. Choosing whether to use it is a separate decision.
