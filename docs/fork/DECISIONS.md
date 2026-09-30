 # 上游 PR / Issue 審核與決策記錄
 
 ## PR #1 - #15 (已收斂納入 v0.1.78)
 
 - **#1 - #7**: CLI 核心重構、子命令收斂為 inherit/read/context/expand/stats、一行摘要與重複折疊（Adopt，保留於核心）。
 - **#8**: 英文第一、繁中雙語文檔（Adopt，本 fork 進一步調整為繁中 README.md 為主，README.en.md 為鏡像）。
 - **#9 - #14**: Harness 事件分類與 promptSource 角色推導（Adopt，核心已包含）。
 - **#15**: 新增 Claude Fable 5.1 作為定價模型（Adopt，核心已包含）。
 
 ## 本 Fork 專屬架構決策 (ADR-Fork-001)
 
 - **移除非 Windows 跨平台構建**：刪除 `install.sh`，`.goreleaser.yaml` 與 CI 只保留 Windows。
- **通用 Agent Handoff Skill**：以唯一的 `skills/agent-handoff` 實作兩階段 `HANDOFF.md` 交接；一般彙整不依賴 CLI，只有讀取歷史 Claude Code session 時使用 `cc-session`。

## 上游審視 baseline 重設（2026-09-30）

- `tools/upstream_baseline.json` 於 2026-09-30 重設到上游當日現況：release `v0.1.82`
  （`7a60e5c`）、PR `#19`、issue `#0`。
- 重設時**沒有**審視 `v0.1.78` 之後的項目：commit `47c3c83`、`5b20138`、`0faa51e`、`7a60e5c`，
  對應 PR `#16` 至 `#19`（新增 Opus 5.5 計價模型、harness 訊息分類、Sonnet 5.5／Opus 5.5 定價、
  `internal/models` 單一來源重構）。它們是「未審視」，不是「審視後略過」。
- 日後要評估這幾項，直接從上游查；本 fork 尚未含有這些變更。
