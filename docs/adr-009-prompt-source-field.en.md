# ADR-009: Prefer the `promptSource` Field for Message Origin, with String Matching as Fallback

[繁體中文](adr-009-prompt-source-field.md) | [English](adr-009-prompt-source-field.en.md)

**Status:** Implemented on 2026-09-02. Decision 4 was corrected the same day to exclude `sdk` as described below. An unclassified `sdk` message is rendered as `user (sdk):`; the compact `context` prefix is `S:`, alongside `U:` and `H:`.

Conflict handling between `promptSource` and string classification, and the “unrecognized system shape” fallback, both live in the parser layer (`claudecodec.parseLineWithToolCalls`). String classification runs first and `PromptSource` is always attached. Only the human sources `typed`, `queued`, and `suggestion_accepted` reset a conflicting harness classification to plain text; `sdk` does not.

The render layer therefore does not repeat string classification. Only when no shape was recognized does it use `PromptSource` to select `harness:`, `user (sdk):`, or `user:` through `session.UserMessage.IsClassifiedAsHarness` and `formatter.plainTextRole`. In session `b11858cf`, all eight `system` messages were already-recognized `<task-notification>` shapes, and the `read` output MD5 stayed unchanged.

ADR-008 classified harness-injected user messages by matching text. Harness wording changes frequently; eight more forms appeared within two days. Claude Code 2.1.165 and later adds `promptSource` to user entries. This ADR measures how much string matching that field can replace.

Measurement set: main and `subagents/` transcripts under `~/.claude/projects`, covering 11,064 user text entries from the most recent 60 days on versions 2.1.165 and later, measured 2026-09-02.

## Field values

| Value | Count | Meaning | Share that started a turn |
|---|---:|---|---:|
| `typed` | 3,547 | Typed by a person in the terminal | 89.6% |
| `sdk` | 1,213 | Sent programmatically through the SDK or `claude -p` | 98.3% |
| `system` | 1,206 | Injected by the harness | 93.9% |
| `queued` | 158 | Queued by a person | 93.0% |
| `suggestion_accepted` | 50 | A person accepted a suggestion | 96.0% |
| absent | 4,890 | See below | |

“Started a turn” uses ADR-008's definition: scan until the next user text entry and check whether an assistant message with usage appears in between. All five values are near or above 90%, so the presence of `promptSource` identifies a prompt that started a turn.

## The 44% without the field are a different category, not missing data

The top 22 shapes among the 4,890 entries without `promptSource` are harness injections or attachments:

| Shape | Count |
|---|---:|
| Teammate messages: 1,080 with preamble plus 769 beginning directly with a tag | 1,849 |
| Skill injection beginning `Base directory for this skill:` | 464 |
| `<system-reminder>` | 439 |
| `[Request interrupted by user]` | 244 |
| `[Image: …]` placeholder | 237 |
| `<local-command-caveat>`, `<local-command-stdout>`, and `<command-*>` | 394 |
| Compaction continuation, system notification, coordinator, fork boilerplate, nudge, and mid-turn user message | 271 |

The semantics are therefore: **present means a turn-starting prompt; absent means an injected or attached entry**. Absence is itself a signal.

Coverage varies from 25% to 89% by version because versions produce different proportions of injected messages, not because the field is intermittently missing.

## Conflicts with string matching

| Case | Count |
|---|---:|
| Human source (`typed`, `queued`, `suggestion_accepted`) classified as harness by string matching | **0** |
| `system` source whose shape string matching did not recognize | 15 |

The 15 unrecognized `system` entries were scheduler or loop prompts written as natural language, such as background-task checks. Only `promptSource` can identify them.

## Decisions

1. **When present, prefer `promptSource`.** `system` is a harness role; `typed`, `queued`, and `suggestion_accepted` are user roles. `sdk` also occupies the user position in that transcript but is labeled `user (sdk):` so readers know it was not typed manually. `sdk` describes how the session is driven, not necessarily who authored an individual message: harness injections in an SDK-driven session also carry `sdk`.
2. **`CountsAsTurn()` returns true for messages with `promptSource`**, subject to the `sdk` exception in decision 4. The ADR-008 per-shape table remains relevant only to messages without the field.
3. **String matching remains for field-less messages.** It is still needed to choose compact forms. A `promptSource=system` message with no recognized shape is rendered in full as `harness:`.
4. **When field and string classification disagree, the field wins except for `sdk`.** The three human values occur only for typing, queuing, or suggestion acceptance, so a conflict indicates a string-classification error. `sdk` is session-wide and can also label harness injection; if a known harness shape appears with `sdk`, keep the compact harness classification. Only an unclassified `sdk` message becomes `user (sdk):`. `CountsAsTurn()` follows the same boundary: known non-turn shapes such as interrupted, agents-stopped, stop-hook, and skill injection remain non-turns even with `sdk`.

   The original rule allowed `sdk` to override string classification because the measurement found no conflicts. After v0.1.76 implemented it, real SDK sessions showed harness shapes expanding in full as `user (sdk):`; this correction fixes that behavior.
5. **Transcripts before 2.1.165 keep the existing path.**

## Rejected alternatives

**Do not treat absence of `promptSource` as harness by itself.** Image placeholders and mid-turn human messages can lack the field. Absence means only “not the prompt that started the turn”; authorship still requires content classification.

**Do not use `isMeta` as a second source.** It appears on skill injection, image placeholders, and stop-hook feedback. Its meaning is “attached metadata,” not “harness.”

## Expected effects

- Fifteen previously unrecognized scheduler prompts gain the `harness:` label.
- Future harness wording changes affect compact formatting only and degrade to full text, rather than corrupting role labels and K.
- K remains effectively unchanged because current policy already counted nearly all messages with `promptSource`; 91% of the `system` group are task notifications counted by ADR-008.
