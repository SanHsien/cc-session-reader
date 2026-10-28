# 路線圖

> 本 repo 後續工作的單一真相源。做完一項，就在同一個 commit 刪掉那一列；新工作只加在這裡。
> 現況段的數字是 2026-10-29 依 `main` `67c60de9defd` 查的快照，開工前先重查 GitHub。

## 現況
這是 Windows 多 Agent 交接工具：日常以 HANDOFF.md 交接，只有匯入舊 Claude Code session 才使用靜態解析器（[README.md:9](README.md)、[README.md:31](README.md)）。2026-10-29 `gh` 查詢顯示 open issue、PR 與安全告警皆為 0，CI、CodeQL、Dependency freshness、Upstream check 最近結果成功，最近 release 為 v0.1.82（2026-10-29 `gh` 查詢）。近期中文截斷與全部顯示修正已在來源碼，但建置腳本會連帶覆寫本機安裝，深掃因而未執行它（commit `d8a6db0`、`b005f8f`；[tools/build_dist.ps1:37](tools/build_dist.ps1)；深掃報告（2026-10-27））。

## 近期：可以直接動手
| 項目 | 為什麼 | 驗收方式 | 建議執行者 | 證據 |
| --- | --- | --- | --- | --- |
| 將打包與本機安裝拆成明確選項 | 一般建置不應意外替換維護者正在使用的執行檔 | 預設只寫輸出目錄，安裝需明確參數；在隔離環境比對安裝檔雜湊未變，amd64／arm64 zip 與 checksums 正常，建置失敗回非零 | Codex | [tools/build_dist.ps1:16](tools/build_dist.ps1)、[tools/build_dist.ps1:35](tools/build_dist.ps1)、[tools/build_dist.ps1:40](tools/build_dist.ps1)；深掃報告（2026-10-27）「刻意不修」 |
| 補修正後 TUI 的實際操作驗收 | 自動測試已涵蓋資料完整性，仍需確認 Windows 主控台操作 | 以假造長會話啟動 TUI，核對中文預覽不亂碼、按 a 顯示至尾行且不重複、q 返回列表、終端歷史仍保留；不讀私人會話 | 維護者 | commit `76605a4`；[README.md:77](README.md)；深掃報告（2026-10-27）「已修」 |

## 中期：等條件成立再做
| 項目 | 觸發條件 | 驗收方式 | 建議執行者 | 證據 |
| --- | --- | --- | --- | --- |
| 發布含 TUI 修正的 fork 版本 | 打包隔離與 Windows TUI 驗收完成，維護者決定發版 | 用候選 SHA 建置 Windows 兩種架構，核對 checksums、版本與下載包，下載後測中文與長文分頁；確認新包可用後才依既定策略處理舊 release | Codex | [docs/fork/DECISIONS.md:14](docs/fork/DECISIONS.md)、[README.md:58](README.md)、[tools/build_dist.ps1:76](tools/build_dist.ps1) |
| 重評窄終端重繪與中文剪貼簿 | 在窄視窗、舊 conhost 或碼頁 950 重現問題 | 保存終端版本、寬度、碼頁與假資料；修正後驗證導航無錯位、中文複製無亂碼且不帶 BOM 字元 | Codex | 深掃報告（2026-10-27）「刻意不修」；[docs/fork/DECISIONS.md:29](docs/fork/DECISIONS.md) |
| 接入 agent-relay 的交接流程 | agent-relay Phase 2 協定與 ownership 邊界確認 | 用合成會話完成彙整、接手與過期 HEAD 拒絕案例；日常交接不強制讀完整歷史，保留對話與關鍵工具摘要 | Codex | [README.md:104](README.md)、[AGENTS.md:11](AGENTS.md)、[README.md:37](README.md) |

## 刻意不做
| 不做什麼 | 理由 | 什麼情況下重新考慮 |
| --- | --- | --- |
| 每次交接都重讀原始 JSONL 或呼叫模型壓縮 | 日常與歷史匯入分工見 [README.md:41](README.md) | 舊 session 不在上下文、確實需要歷史證據 |
| 在未重現前改剪貼簿編碼、暫停行為或 ANSI 控制 | 深掃列出的終端問題尚未證實，UTF-16 BOM 反而曾多出字元 | 有可重現環境與輸入 |
| 引進非 Windows 發布或修復無共同祖先 | 平台策略見 [AGENTS.md:18](AGENTS.md)；歷史壓縮是已知狀態 | 維護者另立平台或歷史遷移需求 |

## 維護方式
每週按 [tools/upstream_baseline.json](tools/upstream_baseline.json) 的 release 追蹤方式核對新增版本與票據；同一 SHA 不重審，理由留在 [docs/fork/DECISIONS.md](docs/fork/DECISIONS.md)。依賴由 freshness 與 Dependabot 提醒；產品修改跑 Go 檢查與 Windows gate，發版前另驗下載執行檔與交接 skill 的入口。
