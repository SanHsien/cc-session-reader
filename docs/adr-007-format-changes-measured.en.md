# ADR-007: Five Output-Format Changes, Decided by Measurement

[繁體中文](adr-007-format-changes-measured.md) | [English](adr-007-format-changes-measured.en.md)

**Status:** Implemented. Measurement code lives on the unmerged `experiment/format-probes` branch.

After implementation, the complete `read` output for the same five sessions changed from 656,370 to 658,278 tokens (**+0.3%**). This is higher than the decision-time estimate of −0.4% because change 1 replaces skipped noise lines with the next, usually longer and more informative line. For example, `=== engine exports ===` (22 characters) becomes `32:export function createBaseGameSnapshot(): GameSnapshotV4 {` (57 characters). That is the intended result, not a regression.

This ADR replaces three ADR-003 decisions: how successful-result excerpts are selected, what Bash summaries contain, and how success is marked. All other ADR-003 decisions remain unchanged.

## Five decisions

| # | Change | Token impact |
|---|---|---:|
| 1 | Apply the failure path's noise filtering to successful excerpts and strip ANSI | Content changes; measured total +0.7% |
| 2 | Add a command verb to Bash summaries: program plus first argument after setup, capped at 30 characters | +1.9% |
| 3 | Stop printing `-> ok`; print only `-> FAILED` | −2.0% |
| 4 | Use `[HH:MM:SS]` timestamps and insert `--- YYYY-MM-DD ---` on date changes | −0.4% |
| 5 | Raise the page limit from 20,000 to 28,000 bytes | 27% fewer pages |

Changes 2 and 3 are shipped together and have a combined impact of **−0.0%**.

Measurement method: `cc-session formatbench` rendered every variant for five sessions ranging from 33K to 389K tokens and sent each result once to the Anthropic token-counting API. The table reports the mean across those five sessions.

## 1. Filter noise from successful excerpts, but do not take the last line

`ToolResult.Summary()` previously used asymmetric paths. Failures used `firstMeaningfulErrorLine`, which skips `cat` line numbers, bare `Exit code N`, and hook boilerplate. Successes used only `FirstLine`, the first non-empty line with no filtering.

The result was that banners appeared as conclusions:

```text
[Bash#Gj9T] check PR #407 status and CI -> ok: Work seamlessly with GitHub from the command line.
[Bash#XhoN] inspect workspace after agent interruption -> ok: --- HEAD ---
[Bash#tQvn] run the new archetype tests -> ok: \x1b[1m\x1b[46m RUN ...
```

Among 1,391 successful Bash results with output, **31.9% of first lines matched a noise shape**:

| Shape | Share |
|---|---:|
| Script section headers such as `=== x ===` or `--- x ---` | 13.7% |
| Code fragments starting with two or more spaces | 8.4% |
| ANSI escape sequences | 6.1% |
| Progress/status prefixes such as `[STARTED]`, `Checking `, or `> ` | 3.3% |
| Version banners | 0.4% |
| At least one shape | **31.9%** |

The successful path now uses `isNoiseExcerptLine`, with the measured banner and progress forms added, and applies `session.StripANSI` before selecting the excerpt. It skips a noisy line and chooses the next line; it does not discard the whole result.

After implementation, 292 of 1,258 Bash calls (23%) in the large sample received better excerpts, none lost an excerpt, and ANSI-bearing lines fell from 95 to 0. The 23% result is below the initial 31.9% because production rules are stricter: a section heading must use matching delimiters at both ends, and indented code is retained because skipping it often only finds another code line.

Plain prose banners have no reliable shape. The `gh` banner remains, but decision 2 adds `| gh pr`, which makes it clear that the banner is not the answer.

### Why not take the last line

A sample of 1,060 successful multi-line Bash results showed no systematic advantage for the last line. Many ended with `})`, a diff tail, or another code fragment. Position does not provide correctness; filtering does.

## 2. Bash summaries need a command verb, and extraction determines the cost

Previously, if a Bash call had a description, `summarizer.go` printed only that description and discarded the command. Readers could not verify what actually ran. Bash accounted for 1,258 of 1,604 tool calls (78%) in the example session, so most actions were unverifiable.

Appending a raw command prefix is expensive because setup consumes the budget:

```text
[Bash#KPso] update PR #423 body | cd /Users/maple/Desk
[Bash#6yGZ] verify two test assertions | echo "=== game-engin
```

Instead, scanning skips `cd`, environment assignments, `export`, and `echo`, then takes the first real program name and its first argument:

```text
[Bash#TV5d] find broken references with typecheck | pnpm tsc: src/...(19,3): error TS24
[Bash#DK5W] force-push rebased bento branch | git push: To https://github.com/...
[Bash#MS6r] inspect the index line | grep -n: 71:- [...]
```

Verb coverage was 1,258 of 1,261 lines (99.8%), at roughly half the cost of raw prefixes:

| Extraction | Token impact |
|---|---:|
| Raw 20-character prefix | +3.7% |
| Raw 60-character prefix | +9.8% |
| Extracted 20-character verb | +1.9% |
| Extracted 30-character verb | +2.0% |

Thirty characters cost nearly the same because extracted verbs rarely exceed 20 characters, so 30 was selected to reduce truncation.

The implementation also skips shell control keywords such as `until`, `while`, and `for`; otherwise `until gh pr checks` would identify `until` as the program. Absolute executable paths such as `/opt/homebrew/bin/gh` are reduced to the basename so they do not consume the budget.

## 3. Mark only failures

The example session had 1,550 successful summaries, 41 failures, and 13 calls without results. Success is the default, so repeating `ok` taxes nearly every line; 96 lines contained only a bare `-> ok`.

`-> FAILED` remains, while success is unmarked, saving 2.0% with the same meaning. `formatRetryCollapse`, which locates `FAILED` in the summary to insert `×N`, was updated at the same time. Successful lines retain `->` as a delimiter between the call and its excerpt.

## 4. Use second-resolution clocks with date-change markers

The old `01-02 15:04` format repeated the date on every message and omitted the year. `context` has a session header that can infer the year, but `read` does not.

The repeated date is wasteful, not the clock. Extracting the date into a marker allows second-level precision while reducing total tokens:

| Format | Example | Token impact |
|---|---|---:|
| Existing | `[08-06 03:06]` | — |
| Add year | `[2026-08-06 03:06]` | +0.6% |
| Add seconds | `[08-06 03:06:15]` | +0.4% |
| Clock + date marker | `[03:06]` plus `--- 2026-08-06 ---` | −0.8% |
| **Clock with seconds + date marker** | `[03:06:15]` plus `--- 2026-08-06 ---` | **−0.4%** |
| Relative minutes | `[+83]` | −1.0% |

The selected format adds seconds and a year while costing 0.4% less. Relative minutes are cheaper but make correlation with CI and server logs harder.

Date markers are necessary: sample sessions crossed two and three days, although transitions were sparse.

## 5. Raise the page limit to 28,000 bytes

`maxPageBytes = 20_000` existed to avoid Claude Code writing large Bash output to a file instead of returning stdout. Binary-search measurements found the actual boundary:

| Output bytes | Result |
|---:|---|
| 30,000 | Complete inline return |
| 31,000 | Written to a file; only a 2 KB preview returned |
| 32,768 | Written to a file |
| 40,000 | Written to a file |
| 50,000 | Written to a file |

The boundary is 30,000 characters. Allowing space for page markers and footers gives a 28,000-byte limit. An 871 KB session fell from 44 pages to 32, reducing round trips by 27%; the largest measured page was 28,073 bytes.

This constant depends on harness behavior rather than cc-session itself. It was verified against one Claude Code version and can also be configured by the user, so the measured 30,000 boundary is documented beside the constant for future retesting.

## Not yet decided

Cross-session search (`cc-session search`) is planned, but its interface and output format need a separate ADR. Search must operate on rendered, filtered text rather than shelling out to `grep`: rendered text exists only in memory, while `grep` can see only the raw JSONL that this feature is meant to replace.

## Boundaries of the measurements

Percentages are means from five sessions ranging from 33K to 389K tokens. Changes 2 and 3 depend heavily on the share of Bash calls. The 31.9% result for change 1 comes from 1,391 results across two sessions. The page boundary in change 5 was verified against only one harness version.
