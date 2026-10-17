# 上游 PR / Issue 審核與決策記錄

## PR #1 - #15 (已收斂納入 v0.1.78)

- **#1 - #7**: CLI 核心重構、子命令收斂為 inherit/read/context/expand/stats、一行摘要與重複折疊（Adopt，保留於核心）。
- **#8**: 英文第一、繁中雙語文檔（Adopt，本 fork 進一步調整為繁中 README.md 為主，README.en.md 為鏡像）。
- **#9 - #14**: Harness 事件分類與 promptSource 角色推導（Adopt，核心已包含）。
- **#15**: 新增 Claude Fable 5.1 作為定價模型（Adopt，核心已包含）。

## 本 Fork 專屬架構決策 (ADR-Fork-001)

- **移除非 Windows 跨平台構建**：刪除 `install.sh`，`.goreleaser.yaml` 與 CI 只保留 Windows。
- **通用 Agent Handoff Skill**：以唯一的 `skills/agent-handoff` 實作兩階段 `HANDOFF.md` 交接；一般彙整不依賴 CLI，只有讀取歷史 Claude Code session 時使用 `cc-session`。
- **嚴格保持單一 Release / Tag**：本倉庫在 GitHub 與本機只保留唯一最新版本之單一 Tag 與 Release，升版時清理舊 Tag，不留歷史 Tag 堆疊。

## PR #16 - #19 與 Release v0.1.82 審核納入（2026-10-18）

- **#16** (`47c3c83`): 新增 Claude Opus 5.5 作為 benchmark 定價模型（Adopt）。
- **#17** (`5b20138`): 修復 2026-09-02 以後 harness 事件格式漂移（Compaction continuation 優先順序修復、CLI 2.1.257–2.1.278 事件分類擴充，Adopt）。
- **#18** (`0faa51e`): 更新 Sonnet 5.5 / Opus 5.5 定價並重設別名（Adopt）。
- **#19** (`7a60e5c`): 新增 `internal/models` 作為模型定價與別名的單一真實來源（Single Source of Truth, SSOT，Adopt）。
- **雙語文檔同步**：全面補齊 `docs/benchmark.zh-TW.md` 與 `docs/adr-008-harness-event-drift.en.md` 鏡像。
- **升版與單一 Tag 策略**：正式發布本 fork `v0.1.82`，並清理舊版 `v0.1.80` Release 與 Tag。
