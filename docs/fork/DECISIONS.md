 # 上游 PR / Issue 審核與決策記錄
 
 ## PR #1 - #15 (已收斂納入 v0.1.78)
 
 - **#1 - #7**: CLI 核心重構、子命令收斂為 inherit/read/context/expand/stats、一行摘要與重複折疊（Adopt，保留於核心）。
 - **#8**: 英文第一、繁中雙語文檔（Adopt，本 fork 進一步調整為繁中 README.md 為主，README.en.md 為鏡像）。
 - **#9 - #14**: Harness 事件分類與 promptSource 角色推導（Adopt，核心已包含）。
 - **#15**: 新增 Claude Fable 5.1 作為定價模型（Adopt，核心已包含）。
 
 ## 本 Fork 專屬架構決策 (ADR-Fork-001)
 
 - **移除非 Windows 跨平台構建**：刪除 `install.sh`，`.goreleaser.yaml` 與 CI 只保留 Windows。
 - **Codex 桌面版 Skill**：新增 `skills/claude-handoff`，提供自然語言免命令列會話接手功能。
