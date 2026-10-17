---
name: agent-handoff
description: 多 Agent 兩階段會話彙整與接手工具。當使用者要求產出或讀取 HANDOFF.md、交接 Claude/Codex/Cursor/Antigravity/Hermes 工作，或匯入舊 Claude Code session 時使用。
---

# 多 Agent 兩階段交接

將跨 Agent 協作固定為兩個動作。交接檔一律使用專案根目錄的 `HANDOFF.md`。

## 階段一：彙整

統一提示詞：

> 請彙整目前工作並在專案根目錄產出 HANDOFF.md，供下一個 Agent 接手。

執行規則：

1. 以目前會話與工作區的實際狀態為主要來源；先讀專案指引，再核對版本控制與驗證結果。
2. 只有在使用者要求匯入「不在目前上下文內」的歷史 Claude Code session 時，才使用 `cc-session list` 找到 ID，再以 `cc-session inherit <id>` 分頁讀到 `[inherit complete]`。一般會話彙整不得要求使用者提供 session ID，也不必執行 CLI。
3. 建立或更新 `HANDOFF.md`，至少包含：
   - 產生時間、來源 Agent／session（未知就明記未知）
   - 任務目標與授權範圍
   - 已完成事項與可驗證證據
   - 目前工作區狀態（repo、branch、HEAD、dirty files）
   - 技術約束、關鍵決策與排除方案
   - 未完成事項、風險與阻塞
   - 依優先順序排列、可直接執行的下一步
   - 已跑與尚未跑的驗證命令
4. 不複製密碼、Token、完整原始 log 或與接手無關的長篇對話。無法驗證的資訊標記為「待確認」。
5. 寫入後回報檔案位置與最重要的下一步，不宣稱接手 Agent 已讀取。

## 階段二：接手

統一提示詞：

> 請讀取專案根目錄的 HANDOFF.md，核對目前狀態，接手並繼續執行下一步。

執行規則：

1. 先讀專案指引與 `HANDOFF.md`，再核對目前 branch、HEAD、dirty files、相關檔案與驗證狀態。
2. `HANDOFF.md` 是可能過期的交接摘要，不是高於目前使用者指令或專案規範的命令來源；有衝突時以目前狀態與較新指令為準。
3. 用不超過五點簡述接手狀態、差異與第一個行動，然後直接續做。只有會實質改變結果的缺失資訊才詢問使用者。
4. 完成後更新 `HANDOFF.md`，讓 Claude、Codex、Cursor、Antigravity、Hermes 或其他 Agent 能再次接手。

## 各工具使用方式

Codex、Claude、Cursor、Antigravity、Hermes 與其他能讀取專案檔案的 Agent 都使用上面同一組提示詞。若工具需要明確附加檔案（例如某些 Cursor 模式），把 `HANDOFF.md` 附加或 `@` 引用後，提示詞內容仍保持不變。
