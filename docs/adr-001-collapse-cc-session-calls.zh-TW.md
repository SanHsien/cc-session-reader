# ADR-001：折疊渲染輸出中的 cc-session 工具呼叫

[English](adr-001-collapse-cc-session-calls.md) | [繁體中文](adr-001-collapse-cc-session-calls.zh-TW.md)

## 狀態

已接受

## 背景

Claude Code session 使用 `cc-session inherit`（舊稱 `cc-session inject`）載入另一個 session 時，JSONL 會把每一頁記成獨立的 Bash 工具呼叫。四頁的匯入會產生四行幾乎相同的紀錄：

```text
[Bash#Y1dg] cc-session inherit 16d06326-... -> ok: [page 1/4 | lines 1-377 of 1320]
[Bash#Ybwi] cc-session inherit 16d06326-... -> ok: [page 2/4 | lines 378-720 of 1320]
[Bash#iNMY] cc-session inherit 16d06326-... -> ok: [page 3/4 | lines 721-1051 of 1320]
[Bash#iMqP] cc-session inherit 16d06326-... -> ok: [page 4/4 | lines 1052-1320 of 1320]
```

這些行對讀者沒有額外資訊：匯入內容已被當時的 AI 消化，結論也已出現在後續 assistant 訊息。當輸出再送入另一個 session 時，保留四行重複紀錄只會浪費空間與 context token。

## 決定

把連續、指向同一 session 的 `cc-session inherit/read/context` Bash 呼叫折疊成一行：

```text
(cc-session: inherited session 16d06326 here, 1320 lines omitted)
```

折疊規則：

- **自動套用**：不需要 flag。原始工具內容仍可透過 `cc-session expand <session> <tool-id>` 查看。
- **只限 cc-session CLI**：其他 Bash 呼叫，包括模型嘗試 `which cc-session`、`node cc-session.mjs` 等，仍照常顯示一行摘要。
- **同時套用於 `read` 與 `context`**：共用 `collapseCCSessionTools()`，由兩種 renderer 的 flush 路徑呼叫。
- **相容舊的 `inject`**：改名以前的 transcript 仍會被折疊，並保留歷史用語「injected」。

### 偵測

`parseCCSessionCommand(cmd)` 檢查 Bash 指令是否符合 `cc-session {inherit|inject|read|context} <session-id>`，並回傳子命令與 session ID 前八碼。以 `-` 開頭的參數會被排除，避免把 `-h` 等 flag 誤認為 session。

### 總行數擷取

`parseTotalLines(text)` 只在工具結果第一行比對 page marker 的 `of N]`，避免 session 內容其他位置出現「of」而誤判。

### 動詞選擇

- `inherit` → `inherited session X here`
- `inject`（舊命令）→ `injected session X here`
- `read` / `context` → `loaded session X here`

## 後果

- 讀者看到一行清楚描述，而不是 N 行重複工具摘要。
- 不再顯示被匯入 session 的實際內容，因為原 session 的 AI 已經整理過。
- 需要時仍可用 `cc-session expand` 檢查單一工具呼叫。
- 未來可考慮 `--follow-refs` 直接解析被引用 session，但那需要定位另一份 JSONL，屬於不同架構。
