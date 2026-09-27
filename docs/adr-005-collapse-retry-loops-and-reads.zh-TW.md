# ADR-005：折疊重試迴圈與連續同檔 Read

[English](adr-005-collapse-retry-loops-and-reads.md) | [繁體中文](adr-005-collapse-retry-loops-and-reads.zh-TW.md)

## 狀態

已接受

## 背景

真實 transcript 有兩種高頻重複模式，每次呼叫都占一行，但資訊幾乎相同：

- **重試迴圈**：模型以小幅變化重試失敗命令。真實樣本中，11 次連續的 `git worktree` 診斷都得到相同的 `Path ... does not exist`。
- **分段讀取**：模型以不同 offset 連續 2–4 次讀取同一大型檔案。

ADR-001 已確立折疊相鄰冗餘行與 `verbose` 例外。本 ADR 把規則延伸到上述兩種模式；這些輸出會被其他 LLM 當作 context，因此規則視為契約。

## 決定

### 規則一：折疊重試迴圈

只有同時符合以下條件的連續工具呼叫才折疊：

- 工具名稱相同；
- 每次呼叫都失敗；
- 正規化後的第一行指令相同，或其中一個是另一個的前綴，且前綴結束於 token 邊界並包含多於一個 token。下一字元必須是空白，裸程式名不能作為前綴依據；例如 `git add` 可匹配 `git add --intent-to-add`，但不能匹配 `git add-on-something`，而裸 `git` 不能吸收 `git add`；
- **錯誤摘錄完全相同**。同一指令產生不同錯誤時不得合併。

格式為：`[Bash#<last-id>] <description> -> FAILED ×N: <last error excerpt>`。保留最後一次呼叫的 tool ID，因為那是重試收斂後的最終狀態；錯誤相等條件也確保顯示的摘錄代表每一次呼叫。

以下情況會中斷折疊：穿插成功、工具不同、任何會被渲染的事件（assistant 文字可能引用單次結果），或設定 `-verbose-bash`。不會渲染的 harness 雜訊、system reminder 與非 verbose thinking 不會中斷群組。

### 規則二：折疊同檔 Read

連續成功讀取同一 `file_path` 的 `Read`（offset/limit 可不同）折疊為 `[Read#<last-id> ×N] <path> -> ok`。任何穿插事件或失敗都會中斷群組。

### 不變條件

- **不得因折疊而丟失失敗資訊**：只有同一錯誤的重複才會壓縮；不同錯誤各自保留。第一版缺少此條件，會被最後一個錯誤蓋掉，已在發布前由對抗式審查發現。
- 單次呼叫不顯示 `×1`。
- 比較策略刻意保守：寧可少折疊，也不要誤折疊。擴大相等規則必須附真實 transcript 與回歸測試。

## 後果

- 11 行重試樣本縮成 6 行；不同命令仍保留，三組同錯誤重試分別顯示 ×2、×3、×4。
- `analyzer` 的 KEPT 分類在折疊後由 render sink 計量，統計仍可對帳。
- 已接受的限制：
  - 指令相等只看第一行。多行 Bash script 若第一行與錯誤都相同仍可能合併；錯誤相等限制降低資訊損失，且每個 ID 仍可 `expand`。
  - 空白正規化不理解 shell quoting；引號內空白不同的指令可能得到同一 signature。沒有真實案例前不增加複雜度。
  - Read 以原始 `file_path` 分組，不 canonicalize。不同 cwd 的相同相對路徑可能誤合併；同一檔案的 symlink、`..`、絕對／相對寫法則可能不合併。考量 cwd 在同一批次中很少變動，並偏好少折疊，接受此限制。
