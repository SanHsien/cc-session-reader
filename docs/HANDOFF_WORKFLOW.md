# 多 Agent 兩階段標準交接工作流

[繁體中文](HANDOFF_WORKFLOW.md) | [English](HANDOFF_WORKFLOW.en.md)

本工作流只要求使用者記住兩句提示詞。來源可以是 Claude、Codex、Cursor、Antigravity、Hermes 或其他 Agent；目標端同樣不限工具。

```text
來源 Agent ──彙整──> HANDOFF.md ──核對──> 接手 Agent ──續做──> 更新 HANDOFF.md
```

## 階段一：彙整

統一提示詞：

> 請彙整目前工作並在專案根目錄產出 HANDOFF.md，供下一個 Agent 接手。

來源 Agent 應：

1. 讀取專案規範，核對目前 branch、HEAD、dirty files 與測試狀態。
2. 根據目前會話與工作區產出 `HANDOFF.md`，不得只複製聊天摘要。
3. 記錄目標、授權範圍、完成事項與證據、關鍵決策、未完成事項、風險、下一步及驗證命令。
4. 將未知或無法重現的資訊標成「待確認」；不得寫入密碼、Token 或完整原始 log。

建議格式：

```markdown
# HANDOFF

- 產生時間：
- 來源 Agent／session：
- Repo／branch／HEAD：
- 工作區狀態：

## 目標與範圍
## 已完成與證據
## 關鍵決策與限制
## 未完成、風險與阻塞
## 下一步（依優先順序）
## 已跑／尚未跑的驗證
```

### 匯入舊 Claude Code session

只有歷史 session 不在目前上下文時才使用解析器：

1. 執行 `cc-session list` 找 session ID。
2. 重複執行 `cc-session inherit <id>`，直到看到 `[inherit complete]`。
3. 核對目前工作區，再將有效資訊寫入 `HANDOFF.md`。

`cc-session` 只讀 Claude Code 本機 JSONL；它不是 Codex、Cursor、Antigravity 或 Hermes 的通用 session 資料庫。其他 Agent 直接依目前上下文產出交接檔。

## 階段二：接手

統一提示詞：

> 請讀取專案根目錄的 HANDOFF.md，核對目前狀態，接手並繼續執行下一步。

接手 Agent 應：

1. 先讀專案規範，再讀 `HANDOFF.md`。
2. 核對 branch、HEAD、dirty files、關鍵檔案與驗證狀態；交接內容與現況不同時，以現況及較新指令為準。
3. 簡要回報接手狀態與差異，立即執行第一個仍適用的下一步。
4. 工作告一段落時更新 `HANDOFF.md`，形成可重複的交接循環。

## 工具差異

提示詞保持一致。若介面要求明確選檔（例如部分 Cursor 模式），附加或 `@HANDOFF.md` 即可；這是介面操作差異，不需要另背一套交接 Prompt。
