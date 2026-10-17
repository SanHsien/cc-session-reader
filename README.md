# cc-session-reader（Windows 多 Agent 會話交接 Fork）

[![CI](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml/badge.svg)](https://github.com/SanHsien/cc-session-reader/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20x64%20%7C%20arm64-blue.svg)](#)

[繁體中文](README.md) | [English](README.en.md)

這是 [`Mapleeeeeeeeeee/cc-session-reader`](https://github.com/Mapleeeeeeeeeee/cc-session-reader) 的 Windows-first 維護 Fork。它保留原專案以 Go 靜態解析 Claude Code JSONL、壓縮工具雜訊與分頁匯入歷史 session 的能力，並新增 `agent-handoff` Skill，讓 Claude、Codex、Cursor、Antigravity、Hermes 等 Agent 以同一份 `HANDOFF.md` 接續工作。

```text
任一來源 Agent
  │  階段一：彙整目前狀態
  ▼
專案根目錄 HANDOFF.md
  │  階段二：目標 Agent 核對後接手
  ▼
Claude / Codex / Cursor / Antigravity / Hermes / 其他 Agent
```

## 統一工作流

所有 Agent 都使用同一組提示詞，不需要記憶各工具專用指令。

### 階段一：彙整

> 請彙整目前工作並在專案根目錄產出 HANDOFF.md，供下一個 Agent 接手。

來源 Agent 會根據目前會話及工作區狀態，記錄目標、完成事項、branch/HEAD、驗證結果、關鍵決策、風險與下一步。一般彙整不需要 session ID，也不必執行 CLI。

只有要匯入「不在目前上下文內」的舊 Claude Code session 時，才讓 Agent 使用 `cc-session list` 與 `cc-session inherit <id>` 讀取經過濾的歷史內容，再更新 `HANDOFF.md`。

### 階段二：接手

> 請讀取專案根目錄的 HANDOFF.md，核對目前狀態，接手並繼續執行下一步。

接手 Agent 先讀專案規範，再核對 branch、HEAD、dirty files 與驗證狀態；`HANDOFF.md` 若已過期，以目前工作區與較新的使用者指令為準。Cursor 等需要明確附檔的介面，可先附加或 `@HANDOFF.md`，提示詞本身不變。

完整說明：[兩階段交接工作流](docs/HANDOFF_WORKFLOW.md)

## 兩個 Skills 的分工

| Skill | 用途 | 適用情境 |
|---|---|---|
| `agent-handoff` | 產出、核對與接續 `HANDOFF.md` | 所有來源與目標 Agent 的日常交接 |
| `cc-session` | 讀取與壓縮歷史 Claude Code JSONL | 舊 session 不在目前上下文，需要匯入時 |

`agent-handoff` 不綁定特定模型或 harness。它會要求 Agent 先核對實際工作區，避免把過期交接摘要當成目前事實。

## 快速安裝（Windows PowerShell）

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/SanHsien/cc-session-reader/main/install.ps1 | iex"
```

安裝程式會：

1. 從本 Fork 的最新 release 下載 Windows `cc-session.exe` 至 `$env:LOCALAPPDATA\cc-session\`。
2. 從本 Fork 安裝 Claude Code 的 `cc-session` 與 `agent-handoff` Skills。
3. 從本 Fork 安裝 Codex 的 `agent-handoff` Skill 至 `~/.codex/skills/agent-handoff`。
4. 產生 `$env:LOCALAPPDATA\cc-session\agent-handoff-claude.zip`。

Claude Desktop 的自訂 Skill 必須由使用者在 **Customize > Skills > + Create skill > Upload a skill** 上傳 ZIP；本機腳本不能替代帳號層的上傳。詳見 [Claude 官方說明](https://support.claude.com/en/articles/12512180-use-skills-in-claude)。

## CLI 速查（進階）

| 命令 | 說明 | 範例 |
|---|---|---|
| `list` | 列出最近的 Claude Code sessions | `cc-session list -n 10` |
| `inherit` | 分頁匯入完整壓縮內容 | `cc-session inherit <id>` |
| `context` | 輸出精簡格式與 metadata | `cc-session context <id>` |
| `read` | 顯示對話與工具摘要 | `cc-session read <id>` |
| `expand` | 展開指定工具呼叫 | `cc-session expand <id> <tool-id>` |
| `stats` | 統計字元與 Token 分布 | `cc-session stats <id>` |
| `ui` | 啟動終端互動面板 (TUI) | `cc-session ui` |

### 終端互動面板 (TUI) 操作指南

雙擊 `cc-session.exe` 或在終端輸入 `cc-session ui`（或無參數啟動）即可開啟互動工作台：

- **會話中心工作台**：啟動後直接呈現在地最近會話列表，無需先查 ID 再手動輸入。
- **方向鍵選取**：使用 `↑ / ↓` 或 `j / k` 移動光棒游標，直接定位目標會話。
- **一鍵觸發操作**：
  - `[Enter]`：直接讀取並壓縮對話內容（read）。
  - `[C]`：直接提取該會話的 Context 精簡標頭（context）。
  - `[S]`：直接統計該會話的字元與 Token 分佈（stats）。
  - `[Y]`：一鍵複製該會話 ID 至 Windows 剪貼簿。
  - `[B]`：執行全域 Token 壓縮率與成本評估（benchmark）。
  - `[M]`：手動輸入自訂或更早的 Session ID。
  - `[Q / Esc]`：離開程式。
- **長會話串流分頁與中途跳出**：長文自動分頁串流，**閱讀中途可隨時按 `q` 或 `Esc` 立即中斷跳出並返回會話列表**，無需等待全篇跑完。
- **歷史保留**：比照 Agent TUI 串流追加模式，不抹除終端歷史紀錄，可隨時向上滾動查閱過去的對話與輸出。

## 開發與驗證

```powershell
pwsh -NoProfile -File tools/dev_check.ps1
```

本 Fork 只維護 Windows amd64/arm64。公開上游文件提供繁體中文與英文鏡像；Fork 內部治理文件可只使用繁體中文。

## 授權與來源

本專案依 Apache License 2.0 授權。原始作品與上游來源詳見 [NOTICE.md](NOTICE.md)、[FORK.md](FORK.md) 與 [LICENSE](LICENSE)。
