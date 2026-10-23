# ADR-010：快取過期後的無損對話重建與接手協議（Dialogue-Preserving Resume）

[繁體中文](adr-010-dialogue-preserving-resume.md) | [English](adr-010-dialogue-preserving-resume.en.md)

**狀態**：已實作（2026-10-23）。

## 1. 背景與核心痛點

在長時間運行或跨日執行的大型 Coding Agent 專案中，開發者常面臨「上下文快取過期（Cache Expiration）」與「長對話記憶斷裂」的兩難：

1. **官方 `/compact` 摘要資訊耗損嚴重**：
   - 官方提供的 Compact 行為傾向高度抽象總結，容易丟失具體編輯過的檔案路徑、精確的錯誤堆疊回溯、未完成的細部待辦清單，以及使用者曾經給予的關鍵架構約束。
   - 在多檔案重構或大型專案中，Compact 後的 Agent 常出現「失憶」、「覆寫已修復程式碼」或「方向偏離」等重大幻覺。
2. **直接載入原始 JSONL 會話導致 Token 瞬間爆炸**：
   - 原始會話記錄（Transcript）充斥大量未過濾的工具輸出、重複讀檔（Read）內容與環境日誌（單個長會話往往高達數百 KB 甚至數 MB，超過數十萬 Token）。
   - 若直接餵入新會話，不僅耗盡 Context 窗口，亦觸發巨額 API 成本。

## 2. 實測基準與壓縮量測數據

以維護者本機真實超長會話 `6b6c9288`（總行數 1,235 行，涵蓋 212 次 API 呼叫、145 次 Bash 調用、跨多個子任務調度）進行實測：

- **原始 Transcript 檔案大小**：`8,554.6 KB`
- **末輪 Model Context**：`493,971 tokens`（約 500k 上下文）
- **字符過濾成效（`cc-session stats`）**：
  - 原始總字符（Raw）：`495,435`
  - 核心保留字符（Filtered）：`73,741`
  - 雜訊消除率（Saved）：**`421,694 (85.1%)`**
- **保留內容結構**：
  - 人類使用者輸入（`user text`）：`11,245 chars` (100% 完整保留)
  - 模型回覆文字（`assistant text`）：`18,686 chars` (100% 完整保留)
  - 關鍵工具調用摘要（`tool summaries`）：`39,438 chars` (保留一行關鍵行為與狀態)
  - 裁剪之底層雜訊（`tool input/result raw`）：`456,905 chars` (徹底消除重複檔案讀取與龐大 stdout 輸出)

## 3. 無損對話重建協議（Dialogue-Preserving Resume Protocol）

本協議定義標準化 3 步驟，讓任何新啟動的 Agent session 能夠安全、完整且低成本地接續前次會話脈絡：

### 步驟 1：定位目標中斷會話
```bash
cc-session list -n 5
```
快速檢視最近中斷或跨日需接續的 session ID、時間與最後對話意圖。

### 步驟 2：評估量體與健康度
```bash
cc-session stats <session-id>
```
檢視過濾比例與保留字元數，確認會話是否包含過期前的完整決策。

### 步驟 3：分頁注入新會話接手
```bash
cc-session inherit <session-id> -reset
```
在新開啟的 Agent 工作流中重複執行，直到顯示 `[inherit complete]`。
- 每頁嚴格限制在 28,000 字符內，安全控制單次 Prompt 吞吐量。
- 注入內容具備時鐘標記 `[HH:MM:SS]`、換日線 `--- YYYY-MM-DD ---` 與精準角色標籤（`user:`, `assistant:`, `harness:`）。
- 保留完整的思考脈絡與架構決定，徹底終結過度抽象摘要引起的失憶問題。

## 4. 效益比對總結

| 評估維度 | 官方 `/compact` 摘要 | 直接讀取原始 JSONL | `cc-session` 無損重建協議 |
| :--- | :--- | :--- | :--- |
| **Token 開銷** | 極低（約 3k~5k） | 極高（400k+，易爆窗） | **最佳平衡（約 20k~30k）** |
| **關鍵路徑保留率** | 偏低（常遭遺漏） | 100%（但被雜訊淹沒） | **100%（提煉為單行摘要）** |
| **未完成待辦追蹤** | 模糊概括 | 需人工全文搜尋 | **精確還原至最後一輪** |
| **跨 Agent 相容性** | 僅限 Claude Code 自身 | 依賴自訂腳本 | **全 Agent 通用（Claude/Codex/Antigravity）** |
