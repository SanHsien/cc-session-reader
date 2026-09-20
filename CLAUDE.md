 # CLAUDE.md for cc-session-reader (Windows Fork)
 
 本專案為 `Mapleeeeeeeeeee/cc-session-reader` 的 Windows 維護 Fork。
 
 ## 原則與規範
 
 - **對外只打主人的 Repo**：PR、push 一律指向 `SanHsien/cc-session-reader`，預設遠端為 `origin`。嚴禁向 `upstream` 提出 PR 或推送。
 - **語言標準**：繁體中文為主要公開文檔（`README.md`），保留英文鏡像（`README.en.md`）。
 - **Windows-first**：所有腳本採用 PowerShell (`.ps1`)，僅維護 Windows amd64/arm64 構建。
 - **本地驗證**：提交前必須執行 `powershell -NoProfile -File tools/dev_check.ps1` 確認綠燈。
