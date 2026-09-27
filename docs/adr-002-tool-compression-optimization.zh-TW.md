# ADR-002：工具結果壓縮最佳化與專案相對路徑

[English](adr-002-tool-compression-optimization.md) | [繁體中文](adr-002-tool-compression-optimization.zh-TW.md)

## 狀態

提案中

## 背景

目前 CLI 以輸出第一行壓縮工具結果，但對 `Read`、`Write`、`Edit`、`Agent` 等工具會產生雜訊：

- **`Read`**：顯示檔案第一行，例如 `import ...`，但檔名已知且成功讀取也已由狀態表達。
- **`Write` / `Edit`**：顯示包含本機絕對路徑的樣板確認訊息，既重複又洩漏環境細節。
- **`Agent`**：只顯示 subagent 已啟動；真正報告已存在 teammate 訊息中。

此外，`Write` 與 `Edit` 的 `ToolUse` 只顯示 basename；不同目錄出現同名檔案時會有歧義。

## 決定

最佳化上述工具的壓縮格式，並從原始 JSONL 的 `cwd` 欄位建立專案相對路徑。

### 1. 專案相對路徑

- parser 與 domain event 新增 `cwd` 欄位。
- `Read`、`Write`、`Edit` 使用 `filepath.Rel`，把絕對路徑轉成相對專案根目錄的路徑。
- 目標在專案外時，退回乾淨路徑或 home-relative 的 `~/...`。

### 2. 工具專用壓縮格式

#### Read

- 輸入含行號範圍時，顯示 `[Read] path/to/file:offset:limit`，例如 `api/mutations.ts:21:120`。
- 成功結果只顯示乾淨的 `-> ok`，不重複行數。

#### Write 與 Edit

- **方案 A（最小化）**：成功只顯示 `-> ok`，例如 `[Edit] src/routes/v1/courses.ts -> ok`。細節由後續 assistant 說明或 `expand` 提供。
- 失敗時永遠保留原始錯誤訊息。

#### Agent

- 成功啟動壓縮為 `-> ok`。
- 實際 subagent 報告仍完整保留於 teammate 訊息。

## 下一階段最佳化

未來可在不使用 LLM 的前提下產生更有意義的程式碼變更摘要：

- 解析 `structuredPatch`，統計增刪行與位置，例如 `-> ok (+15, -3 lines at L20)`。
- 靜態擷取函式／類別名稱或極短 diff 預覽。
- Bash 失敗時跳過 `Exit code 1` 樣板，直接擷取真正的編譯或測試錯誤。

## 後果

- 時間線更乾淨，節省 context token。
- 專案相對路徑消除同名檔案歧義。
- 輸出不再洩漏本機絕對路徑。
- 未來仍需在更多專案歷史上量測 Token 節省效果。

## 基準結果

以 `./cc-session benchmark -n 5 -no-api` 比較五個具代表性的歷史 session：

### 最佳化前

- **Token 節省中位數**：**82.1%**（平均 81.5%，範圍 78.9%–82.2%）
- **Opus 成本節省**：10 turn 72%，100 turn 38%

### 最佳化後（專案相對路徑與方案 A）

- **Token 節省中位數**：**81.9%**（平均 81.3%，範圍 78.9%–82.0%）
- **Opus 成本節省**：10 turn 72%，100 turn 38%

### 0.2% 差異分析

相對路徑（例如 `src/features/myinccu/api/mutations.ts`）比 basename（例如 `mutations.ts`）稍長；在 500K token 的極長 session 約增加 1,000 token。這是解決同名歧義與路徑洩漏所付出的合理成本。
