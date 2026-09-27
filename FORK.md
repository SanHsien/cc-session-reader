 # Fork 維護說明
 
 本 repo fork 自 [`Mapleeeeeeeeeee/cc-session-reader`](https://github.com/Mapleeeeeeeeeee/cc-session-reader)，遵循 Apache 2.0 License 並保留完整的 Git 歷程。
 
 ## 為什麼維護此 fork
 
 - **Windows-first 專屬支援**：專注於 Windows 11 + PowerShell 環境，移除非 Windows 跨平台構建雜訊。
- **跨 Agent 兩階段交接**：新增通用 Skill（`skills/agent-handoff`），讓 Claude、Codex、Cursor、Antigravity、Hermes 等 Agent 以同一組提示詞產出或接手 `HANDOFF.md`；歷史 Claude Code session 才由解析器匯入。
 - **純靜態高倍率壓縮**：不使用 LLM、零額外 Token 成本，將 Claude Code / Desktop 的原始會話（數十萬 Token）壓縮 80%+，保留對話與關鍵工具結果。
 - **繁體中文首要文檔**：提供完整的繁體中文（zh-TW）與英文雙語鏡像文檔。
 - **規範化的治理與檢查 Gate**：建立 `tools/dev_check.ps1`，提供 Windows 本地一鍵驗證與上游追蹤。
 
 ## 核心配置與檔案
 
 | 檔案 | 說明 |
 |---|---|
 | `AGENTS.md` / `CLAUDE.md` | 本 fork 的 AI 治理規範與路由指引 |
 | `NOTICE.md` / `FORK.md` | 原始版權聲明、變更記錄與 fork 宗旨 |
 | `README.md` / `README.en.md` | 繁體中文主要說明與英文鏡像 |
| `skills/agent-handoff/SKILL.md` | Codex、Claude 與其他 Agent 共用的兩階段交接 Skill |
 | `tools/dev_check.ps1` | Windows 驗證 gate（二進位檢查 + Skill 完整性 + 上游檢查） |
 | `tools/upstream_baseline.json` | 上游已審核基線記錄 |
 | `docs/fork/DECISIONS.md` | 上游決策與變更記錄 |
 
 ## 遠端設定
 
 - `origin/main`：`SanHsien/cc-session-reader`（主要維護分支）。
 - `upstream/main`：`Mapleeeeeeeeeee/cc-session-reader`（上游母庫，僅拉取、不推送）。
