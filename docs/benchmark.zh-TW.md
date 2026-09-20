# 基準測試：cc-session 成本節省效益分析

[繁體中文](benchmark.zh-TW.md) | [English](benchmark.md)

比較在 Claude API Prompt Cache 過期（5 分鐘 TTL）後兩種情境的輸入 Token 成本：

- **情境 A**：停留在原始會話中。首次 API 呼叫時，整個上下文會被重新快取。
- **情境 B**：開闢新會話，透過 `cc-session` 注入壓縮後的歷史紀錄，然後繼續作業。

---

## 快速開始

### 1. 測量你的固定開銷 (只需一次)

開啟一個乾淨的 Claude Code 會話，輸入簡短文字（例如 "hi"），然後執行：

```bash
cc-session stats <session-id>
```

記下 `Last turn context` 的數值。這就是你的會話固定開銷（系統提示詞 + 工具定義 + CLAUDE.md + 規範規則）。每個使用者的數值不同。

### 2. 執行基準測試

```bash
# 掃描最近 20 個會話
cc-session benchmark -n 20
```

---

## 實測統計

在日常工具密集型的開發會話中，平均 Token 壓縮率達 **80%–88%**。使用者與助理的核心推理完整保留，但大量無效的工具 raw JSON 與結構被有效過濾，大幅降低快取失效後的重載成本。
