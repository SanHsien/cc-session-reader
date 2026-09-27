# AGENTS.md for SanHsien/cc-session-reader

本專案為 `Mapleeeeeeeeeee/cc-session-reader` 的維護 Fork，針對 Windows、Codex 桌面版與 Claude Desktop 提供最佳化整合與繁體中文支援。

## 核心原則與邊界

1. **對外只打主人的 Repo**：
   - 預設遠端一律指向 `SanHsien/cc-session-reader`。
   - 禁止向 upstream（母倉庫）直接 push 或開 PR。
2. **GUI 優先與兩階段交接**：
   - 本 repo 提供 `skills/agent-handoff`，讓 Claude、Codex、Cursor、Antigravity、Hermes 等 Agent 以同一組自然語言提示詞產出或接手 `HANDOFF.md`。
   - 只有匯入不在目前上下文的舊 Claude Code session 時，才由 Agent 調用已編譯的 `cc-session` 靜態解析工具。
3. **無損壓縮與 Token 防爆**：
   - 提取對話精華與關鍵 tool log，靜態過濾 80%+ harness 雜訊，防止上下文過載。
4. **語言標準**：
   - 繁體中文為主要公開文檔（`README.md`），上游公開 Markdown 保留繁體中文與英文鏡像；Fork 內部治理文件可只使用繁體中文。
5. **Windows-first**：
   - 所有腳本採用 PowerShell（`.ps1`），僅維護 Windows amd64/arm64 構建。
6. **本地驗證**：
   - 提交前必須執行 `powershell -NoProfile -File tools/dev_check.ps1` 確認綠燈。

