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

## Windows 本地體驗強化（2026-10-18）

- **PE 原生 App 圖示內嵌**：設計專屬多尺寸向量/圖素圖示 (`assets/icon.png`、`assets/icon.ico` 涵蓋 16~256px)，並透過 `.syso` 嵌入 Windows PE resource 段，使編譯執行檔具備原生現代化應用程式圖示。
- **獨立 Console 雙擊防閃退機制**：透過 Windows API `GetConsoleProcessList` 檢測是否由檔案總管雙擊建立獨立終端；若為雙擊且未帶參數或遇錯，自動進入 TUI 或提示暫停，徹底杜絕執行後閃退問題。
- **輕量終端互動面板 (TUI)**：實作無額外 Webview/HTTP 負擔的零依賴 TUI（`ui` / `tui` 子命令，或無參數終端自動啟動），提供數字選單、最近會話序號讀取、格式化預覽、一鍵複製至系統剪貼簿 (`clip`) 等友善按鈕化體驗。
- **Agent TUI 串流追加與方向鍵導航**：取消畫面清空行為，比照 Agent TUI 採用串流追加模式，完整保留終端歷史滾動條；並透過原生 Windows `msvcrt.dll` (`_getch`) 與 ANSI 局部游標回跳 (`\033[NA`)，支援 `↑ / ↓`、`j / k` 移動光棒游標、`Enter` 確認，且相容直接按數字鍵與管線非互動輸入。
