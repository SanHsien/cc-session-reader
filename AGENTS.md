# AGENTS.md for SanHsien/cc-session-reader

本專案為 `Mapleeeeeeeeeee/cc-session-reader` 的維護 Fork，針對 Windows、Codex 桌面版與 Claude Desktop 提供最佳化整合與繁體中文支援。

## 核心原則與邊界

1. **對外只打主人的 Repo**：
   - 預設遠端一律指向 `SanHsien/cc-session-reader`。
   - 禁止向 upstream（母倉庫）直接 push 或開 PR。
2. **零命令列 GUI 導向**：
   - 本 repo 提供 `skills/claude-handoff`，讓使用者在 **Codex 桌面版** 對話框直接以自然語言接手 Claude Desktop / Claude Code 的 Session。
   - 底層調用已編譯好的靜態解析工具，不依賴手動執行 CLI。
3. **無損壓縮與 Token 防爆**：
   - 提取對話精華與關鍵 tool log，靜態過濾 80%+ harness 雜訊，防止上下文過載。

