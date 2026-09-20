---
name: claude-handoff
description: |
  接手 Claude Code / Claude Desktop 的 Session 對話與工作進度。
  當使用者在 Codex 說「接手 Claude 的 session」、「接手 Claude 最近的進度」、「讀取 Claude 對話」、「handoff from claude」、「繼承 Claude 上下文」時觸發。
  本 Skill 會自動在背景呼叫靜態解析器，過濾 80%+ 雜訊並注入精華對話，完全不消耗多餘 Token。
---

# Claude Session Handoff for Codex

本 Skill 專為 **Codex 桌面版** 設計，用來在 Claude Desktop / Claude Code 達到 5 小時用量限制（Rate Limit）或需要切換到 Codex 時，無縫繼承對話與任務脈絡。

## 運作機制

1. 靜態解析器位置：$env:LOCALAPPDATA\cc-session\cc-session.exe
2. 原始對話存放於 ~/.claude/projects/（包含 Claude Code 與 Claude Desktop 本機 Code Sessions）。
3. 使用者無需開啟任何命令列，Codex 會直接在背景執行指令並載入對話摘要。

## 執行流程

當使用者提出接手或讀取請求時：

### 1. 判斷目標 Session
- **如果使用者指名了 Session ID 或關鍵字**：直接使用該 ID。
- **如果使用者說「接手最近的」、「接手剛才的」**：
  在背景執行：
  `powershell
  &  C:\Users\SanHsien\AppData\Local\cc-session\cc-session.exe list -n 5
  `
  選取第一筆（最近更新）的 Session ID。

### 2. 取得清洗後的對話上下文
使用 context 子命令（或長對話用 inherit）提取已剔除大量 tool raw data、保留核心推理與對話的乾淨上下文：
`powershell
& C:\Users\SanHsien\AppData\Local\cc-session\cc-session.exe context <SESSION_ID>
`

### 3. 接手並回報使用者
讀取完成後，Codex 應立即：
1. **簡要總結接手進度**：說明該 Session 當時處理到哪裡（包含目標、已完成事項、卡點或最後進度）。
2. **無縫續做**：依照前一個 Session 的既定方向，主動接手下一個具體動作或開始編寫代碼。

