# ADR-003：工具結果狀態判定與錯誤／Diff 摘要

[English](adr-003-tool-result-status-and-summaries.md) | [繁體中文](adr-003-tool-result-status-and-summaries.zh-TW.md)

## 狀態

已接受

## 背景

cc-session 會為每個工具結果顯示 `-> ok` 或 `-> FAILED`。讀取注入 context 的 LLM 會依賴這個狀態，所以錯誤的成功標記比缺少細節更危險。

現行實作只讀 `toolUseResult.success`，欄位缺失時預設為 true；真實 Claude Code JSONL 顯示這個預設不成立：

- Bash 結果只有 `stdout`、`stderr`、`interrupted`、`isImage`、`noOutputExpected`，沒有 `success`。因此非零退出碼也會顯示成自相矛盾的 `-> ok: Exit code 1`。
- `Read` 結果也沒有 `success`。
- 只有少數工具（例如 Agent lifecycle）明確設定 `success`。
- codec 沒有解析 `tool_result` content block 的 `is_error`；即使解析，它也不是可靠的成功訊號，因為真實 transcript 中有 `is_error: false` 但文字為 `Exit code 1` 的 Bash 結果。

另外，ADR-002 建議替 Edit/Write 產生 diff 摘要並改善 Bash 錯誤摘錄。真實 Edit 結果包含 `structuredPatch` hunk，因此可以不使用 LLM 就算出一行摘要。

## 決定

### 1. 狀態判定階梯

依第一個適用規則判定：

1. 有 `toolUseResult.success`：採用明確值。
2. content block 的 `is_error: true`：失敗。
3. 以保守、列舉式樣式檢查結果文字：
   - 出現 `Exit code N` 且 N 不為 0；
   - 以 `... hook error` 開頭。
4. 其餘視為成功。

文字檢查只使用已知失敗特徵的 allowlist，不做模糊啟發式。每個新樣式都必須有真實 transcript 證據與回歸測試。

### 2. 錯誤摘錄

失敗時保留第一個**有意義**的錯誤行，而不是機械式取第一行：

- 跳過 `cat -n` 行號、hook 樣板與裸 `Exit code N`。
- 失敗結果使用較大的單行預算，約 200 字元，因為錯誤最不應被丟失。

狀態文字仍維持 `ok` / `FAILED`，避免破壞既有輸出契約。

### 3. Edit/Write diff 摘要

成功的 Edit/Write 從裸 `-> ok` 升級：

- **Edit** 且 `structuredPatch` 非空：`-> ok (+A, -D @ L<newStart>)`；多個 hunk 時附加 `H hunks`。
- **Write** 新檔且 patch 為空：依 `content` 顯示 `-> ok (new file, N lines)`。
- patch 缺失或無法解析：退回 `-> ok`。

### 4. 未知工具備援

summarizer 的預設分支改為顯示前 2–3 個輸入 key/value，每個值約截至 60 字元。Claude Code 新增工具時，至少仍保留基本上下文。

## 不可刪除的負面知識

- Bash、Read 等高頻工具沒有 `toolUseResult.success`；欄位缺失時無論預設 true 或 false 都只是猜測。
- `is_error: false` 不是成功證據，只能在 true 時當成失敗證據。
- 因此不能把文字檢查簡化成「只信 flags」。

## 後果

- 注入 context 對成功／失敗的描述更可信。
- Diff 摘要雖增加少量 token，卻能直接回答「該 session 改了什麼」。
- 未知失敗形式仍可能暫時判為成功；`audit` 命令負責協助找出這些缺口。
