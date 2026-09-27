# 基準測試：cc-session 成本節省

[English](benchmark.md) | [繁體中文](benchmark.zh-TW.md)

比較 Claude API Prompt Cache 超過五分鐘 TTL 後，兩種情境的輸入 Token 成本：

- **情境 A**：留在原 session，第一次 API 呼叫重新寫入完整 context cache。
- **情境 B**：開新 session，以 cc-session 匯入壓縮歷史後繼續工作。

## 快速開始

### 1. 測量固定開銷（一次即可）

開一個新的 Claude Code session，輸入簡短訊息後執行：

```bash
cc-session stats <session-id>
```

記下 `Last turn context`。這是你的 session 固定開銷，包含 system prompt、工具定義、CLAUDE.md 與規則；每位使用者不同。

### 2. 執行基準測試

基準測試使用 Anthropic token counting API 計算壓縮歷史，因此必須設定 `ANTHROPIC_API_KEY` 或 `anthropic_api_key_file`。

```bash
cc-session benchmark --overhead <你的數值>
```

其他參數會從真實 session 資料自動推導。

### Flags

| Flag | 預設值 | 說明 |
|---|:---:|---|
| `--overhead` | 40000 | 實測的 session 固定開銷 token |
| `--days` | 30 | 回溯 session 的天數 |
| `--min-kb` | 100 | JSONL 最小檔案大小（KB） |
| `--n` | 10 | 最多回報幾個成功結果 |
| `--model` | opus | 定價與 token counting 模型：`opus`、`opus-4-6`、`opus-4-7`、`opus-4-8`、`sonnet` 或 `fable`（`fable-5-1`） |

### 輸出範例

```text
=== Compression ===
Session        Context      NewCtx   Saved
e61060b1       403,129     125,089  69.0%
977b7360       381,032      76,808  79.8%

Median: 74.4%   Mean: 74.4%   Range: 69.0% — 79.8%

=== Cost Savings Per Session (opus) ===
Session        Context      NewCtx      K  Break-even   10-turn   100-turn
e61060b1       403,129     125,089    4.5      turn 1       66%        52%
977b7360       381,032      76,808    5.5      turn 1       73%        45%

Median break-even: turn 1 | 10-turn saving: 69% | 100-turn saving: 48%
```

## 成本模型

### 定價

依 [Anthropic Prompt Caching 文件](https://platform.claude.com/docs/en/build-with-claude/prompt-caching)：

| 類別 | API 欄位 | Opus 費率 |
|---|---|:---:|
| Cache read | `cache_read_input_tokens` | $0.50/M（基礎費率 0.1 倍） |
| Cache write | `cache_creation_input_tokens` | $6.25/M（基礎費率 1.25 倍） |
| 未快取 input | `input_tokens` | $5.00/M（基礎費率 1 倍） |
| Output | `output_tokens` | $25/M；兩情境相同，因此排除 |

### 每次 API 呼叫的計費

每個請求的輸入 token 分成：

- **cache read**：與先前 cache 相符的 prefix；
- **cache write**：寫入 cache 的新內容，直到自動 breakpoint；
- **uncached**：最後 breakpoint 之後的內容，自動快取下通常接近零。

### 多輪 cache 行為

| 請求 | Cache 行為 |
|---|---|
| Request 1 | 全部寫入 cache |
| Request N | 從 cache 讀取 system 到 User(N-1)，把 Asst(N-1) + User(N) 寫入 cache |

跨輪 cache write 等於上一輪回覆 R 加上新 prompt P。

### 工具呼叫的 cache

依 [tool-use-with-prompt-caching](https://platform.claude.com/docs/en/agents-and-tools/tool-use/tool-use-with-prompt-caching)：

- bash/read/edit 等 client-side tools 是不同 API 呼叫，各自計算 cache。
- server-side tools 有自動 breakpoint，但不適用於 Claude Code 的 client tools。

若每個 user turn 有 K 次工具呼叫，就會有 K 次 API 請求，每次都對完整 prefix 付 cache-read 成本。K 為小數時，模擬器會計算完整額外呼叫，再按比例加入下一個呼叫，不做整數四捨五入。

### 公式

**情境 A**（cache 過期後留在原 session）：

```text
Turn 1:
  Call 1: (X + P) × CacheWrite
  Calls 2..K: prefix × CachedRead + s × CacheWrite

Turn N (N≥2):
  prefixFromPrev = (X + P) + (N-1)×toolIOPerTurn + (N-2)×growth
  Call 1: prefixFromPrev × CachedRead + growth × CacheWrite
  Calls 2..K: (prefixFromPrev + growth + s×(c-2)) × CachedRead + s × CacheWrite
```

**情境 B**（新 session + cc-session）：

```text
Setup:
  injectPages <= 1:
    (overhead + C) × CacheWrite

  injectPages > 1:
    overhead × CacheWrite
    每一頁 i:
      (overhead + 先前頁 token) × CachedRead
      + pageTokens × CacheWrite

  pageTokens ≈ C / injectPages

Turn 1:
  Call 1: base × CachedRead + P × CacheWrite
  Calls 2..K: (base + P + s×(c-2)) × CachedRead + s × CacheWrite

Turn N (N≥2): 與 A 相同，但 base 較小
```

### 參數

| 參數 | 來源 | 說明 |
|---|---|---|
| X | `stats.LastContextTokens` | 原 session context 大小 |
| C | 對 `filteredText` 呼叫 Anthropic token counting API | cc-session 壓縮歷史大小 |
| overhead | `--overhead` | System + tools + CLAUDE.md |
| NewCtx | `overhead + C` | 匯入後的新 session 總 context |
| injectPages | `inject.SplitPages(inject.RenderFullOutput(...))` | 匯入壓縮歷史需要的頁數 |
| K | `APICallCount / UserTurnCount` | 每個 user turn 的 API 呼叫數 |
| toolIOPerCall（s） | `Σ(InputChars + ResultChars) / Σ(CallCount) / 2` | 每次 API 呼叫的平均工具 I/O |
| avgResponse | `TotalOutputTokens / APICallCount` | 每次 API 呼叫平均 output token |
| prompt（P） | `(LastContext - overhead) / turns - avgResp - toolIO×(K-1)` | 平均 user prompt |
| growth | `avgResponse + prompt` | 跨輪 cache write（R + P） |

壓縮表比較相同層級的總 context：X 與 NewCtx。成本模擬仍分開保留 C、overhead 與 injectPages，因為 NewCtx 是匯入後的最終大小，不一定是單一請求成本。`cc-session inject` 會以每頁不超過 20K bytes 分頁，每頁都建模為獨立 API 呼叫；單頁則維持歷史的一次性 `NewCtx × CacheWrite` 行為。

X 來自 transcript API usage；C 來自 Anthropic token counting API。`--model` 同時控制定價與 tokenizer：`opus` 是 `opus-4-8` 別名；明確的 Opus 版本對應各自模型；`sonnet` 對應 `claude-sonnet-4-6`；`fable` 與 `fable-5-1` 都對應 `claude-fable-5-1`。Opus 4.6、4.7、4.8 使用相同費率。

只有 transcript 無法直接提供的行為，例如工具 I/O 資料太稀疏時，才使用 fallback 常數。Claude Fable 5.1 的 cache read 是基礎 input 的 2.5%，不是其他模型的 10%，因此節省數字不能直接與 Opus/Sonnet 橫向比較。

### 簡化假設

- **省略未快取 input**：自動快取會把 breakpoint 放在最後可快取區塊，未快取尾端接近零。這是由文件推論，不是明確保證；敏感度分析估計影響低於 2%。
- **排除 output token**：兩情境相同。
- **略過已 compact 的 session**：若有 `compact_boundary`，原 context 已被摘要，兩情境不再可比。

## 內部運作方式

```text
parser.ListAllSessions
  掃描 ~/.claude/projects/* 與 session-meta
        │
        ▼
reader.ReadAll(path)
  解析 JSONL、工具呼叫、compact boundary 與 API usage
        │
        ▼
analyzer.ComputeStats(events)
  累計 RawText、FilteredText、LastContextTokens、API 呼叫、turn 與工具 I/O
        │
        ▼
tokens.NewCounter(model)
  解析 API key 並重用 token counter
        │
        ▼
counter.Count(filteredText)
  跳過 compact 或缺 usage 的 session，直到取得 n 個成功結果
        │
        ▼
inject.RenderFull + SplitPages
  以實際匯入格式與分頁限制計算頁數
        │
        ▼
推導 K、toolIOPerCall、avgResponse、prompt、growth
        │
        ▼
cumulativeCostA / costBWithPages
  模擬 N turns × K API calls，累加 cache read/write
        │
        ▼
輸出壓縮表與成本節省表
```

### 主要資料來源

- **Session JSONL**：`~/.claude/projects/<project>/<session-id>.jsonl`，包含訊息、工具呼叫、結果與 API usage。
- **Session metadata**：`~/.claude/usage-data/session-meta/<session-id>.json`，提供專案路徑、時間與訊息數等索引。
- **API usage 欄位**：assistant message 事件內的 `input_tokens`、`cache_read_input_tokens`、`cache_creation_input_tokens`、`output_tokens`，用來計算真實 `LastContextTokens` 與 `TotalOutputTokens`。
