# 貢獻指引 (Contributing Guide)

[繁體中文](CONTRIBUTING.md) | [English](CONTRIBUTING.en.md)

感謝你對本專案的關注！本倉庫是專為 Windows 與多 AI Agent（Codex、Claude、Cursor、Antigravity、Hermes）協同開發設計的維護 Fork。

## 原則與規範

1. **對外只打主人的倉庫**：
   - PR、push 一律指向 `SanHsien/cc-session-reader`，預設遠端為 `origin`。
   - **嚴禁向 upstream 母倉庫直接提出 PR 或推送**。
2. **Windows 優先**：
   - 移除非 Windows 的跨平台建置雜訊，所有本機自動化均以 PowerShell (`.ps1`) 撰寫。
3. **本地驗證標準**：
   - 提交前必須在本地 Windows 終端執行：
     ```powershell
     powershell -NoProfile -File tools/dev_check.ps1
     ```
   - 確保二進位檔可執行、Skill 結構完整且上游更新已審核（顯示 `WINDOWS DEV CHECK GREEN`）。
4. **雙語公開文件同步**：
   - 上游公開 Markdown 的繁體中文與英文版本必須同步更新；Fork 內部治理文件可只使用繁體中文。
