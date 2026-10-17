package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/Mapleeeeeeeeeee/cc-session-reader/internal/claudecodec"
	"github.com/Mapleeeeeeeeeee/cc-session-reader/internal/parser"
	"github.com/Mapleeeeeeeeeee/cc-session-reader/internal/tracker"
)

type menuItem struct {
	key   string
	label string
	desc  string
}

var mainMenuItems = []menuItem{
	{key: "1", label: "📋 列出最近會話 (list)", desc: "列出最近的 Claude Code 會話，支援方向鍵/序號選取讀取"},
	{key: "2", label: "🔍 讀取並壓縮指定會話 (read)", desc: "完整對話 + 工具一行摘要，可複製至剪貼簿"},
	{key: "3", label: "📦 產生精簡 Context 標頭格式 (context)", desc: "精簡注入格式（帶 metadata header）"},
	{key: "4", label: "📊 Token 壓縮與成本評估 (benchmark)", desc: "掃描近期會話，計算壓縮率與成本比較"},
	{key: "5", label: "📈 會話字元與 Token 統計 (stats)", desc: "字元與 Token 分佈統計"},
	{key: "6", label: "🛠 檢視完整 CLI 子命令速查表 (help)", desc: "顯示 CLI 完整子命令速查表"},
	{key: "0", label: "🚪 離開 (exit)", desc: "關閉並離開程式"},
}

// cmdTUI launches the interactive terminal user interface.
func cmdTUI(args []string, reader claudecodec.Codec) {
	runInteractiveMenu(reader)
}

func runInteractiveMenu(reader claudecodec.Codec) {
	inReader := bufio.NewReader(os.Stdin)
	store := parser.DefaultStoreWith(reader)

	fmt.Println("============================================================")
	fmt.Printf("  cc-session Reader %s - 終端互動面板 (TUI)\n", version)
	fmt.Println("============================================================")

	for {
		var choice string
		if isInteractiveTerminal() {
			choice = selectMenuInteractive(inReader, mainMenuItems)
		} else {
			choice = selectMenuNonInteractive(inReader, mainMenuItems)
		}

		choice = strings.ToLower(strings.TrimSpace(choice))

		switch choice {
		case "1", "list":
			handleListInteractive(inReader, store, reader)
		case "2", "read":
			handleReadInteractive(inReader, store, reader)
		case "3", "context":
			handleContextInteractive(inReader, store, reader)
		case "4", "benchmark":
			handleBenchmarkInteractive(inReader, store, reader)
		case "5", "stats":
			handleStatsInteractive(inReader, store, reader)
		case "6", "help":
			handleHelpInteractive(inReader)
		case "cls", "clear":
			clearScreen()
		case "0", "q", "quit", "exit":
			fmt.Println("\n謝謝使用 cc-session！")
			pauseIfAlone()
			return
		default:
			fmt.Println("\n[無效輸入] 請輸入 0 到 6 的數字。")
		}
	}
}

func selectMenuNonInteractive(in *bufio.Reader, items []menuItem) string {
	fmt.Println("\n------------------------------------------------------------")
	for _, it := range items {
		fmt.Printf("   [%s] %s\n", it.key, it.label)
	}
	fmt.Println("------------------------------------------------------------")
	fmt.Print("請選擇操作 [1-6, 0] > ")
	return readLine(in)
}

func selectMenuInteractive(in *bufio.Reader, items []menuItem) string {
	fmt.Println("\n------------------------------------------------------------")

	selected := 0
	numItems := len(items)
	totalRenderLines := numItems + 2

	render := func(initial bool) {
		if !initial {
			fmt.Printf("\033[%dA", totalRenderLines)
		}
		for i, it := range items {
			prefix := "    "
			colorStart := ""
			colorEnd := ""
			if i == selected {
				prefix = "  ❯ "
				colorStart = "\033[1;36m"
				colorEnd = "\033[0m"
			}
			fmt.Printf("\033[2K\r%s[%s] %s%s\n", prefix, it.key, colorStart+it.label+colorEnd, "")
		}
		fmt.Printf("\033[2K\r------------------------------------------------------------\n")
		fmt.Printf("\033[2K\r 使用 [↑ / ↓] 或 [j / k] 移動游標，[Enter] 確認，或按快捷鍵 [1-6, 0]\n")
	}

	render(true)

	for {
		ev := readKeyInteractive()
		switch ev.kind {
		case keyUp:
			selected = (selected - 1 + numItems) % numItems
			render(false)
		case keyDown:
			selected = (selected + 1) % numItems
			render(false)
		case keyEnter:
			return items[selected].key
		case keyEscape:
			return "0"
		case keyRune:
			ch := strings.ToLower(string(ev.char))
			switch ch {
			case "k":
				selected = (selected - 1 + numItems) % numItems
				render(false)
			case "j":
				selected = (selected + 1) % numItems
				render(false)
			case "q":
				return "0"
			default:
				for _, it := range items {
					if strings.EqualFold(it.key, ch) {
						return it.key
					}
				}
			}
		}
	}
}

func handleListInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 最近的 Claude Code 會話列表：")
	fmt.Println("--------------------------------------------------------------------------------")

	entries, warnings := store.ListAllSessions()
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "警告: %s\n", w)
	}

	limit := 15
	if len(entries) < limit {
		limit = len(entries)
	}

	if limit == 0 {
		fmt.Println("未找到任何 Claude Code 會話。")
		return
	}

	callerIDs := tracker.CallerSessionIDs()

	if !isInteractiveTerminal() {
		for i := 0; i < limit; i++ {
			entry := entries[i]
			projectName := "?"
			if entry.ProjectPath != "" {
				projectName = filepath.Base(entry.ProjectPath)
			}
			dateStr := "??-??"
			if entry.StartTime != "" {
				dateStr = parser.FormatTimestamp(entry.StartTime)
			}
			refMarker := ""
			if callerIDs[entry.SessionID] {
				refMarker = " [refs]"
			}
			prompt := entry.FirstPrompt
			if len(prompt) > 35 {
				prompt = prompt[:32] + "..."
			}
			fmt.Printf(" [%2d] %s  %s  %-15s %3dm  u:%-2d a:%-2d%s  %s\n",
				i+1, entry.SessionID, dateStr, projectName, entry.DurationMinutes,
				entry.UserMessageCount, entry.AssistantMessageCount, refMarker, prompt)
		}
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Print("請輸入序號 [1-", limit, "] 直接讀取壓縮，或按 Enter 返回主選單 > ")

		input := strings.TrimSpace(readLine(in))
		if input == "" {
			return
		}

		idx, err := strconv.Atoi(input)
		if err == nil && idx >= 1 && idx <= limit {
			readAndDisplaySession(in, entries[idx-1].SessionID, store, reader)
		}
		return
	}

	selected := 0
	totalRenderLines := limit + 2
	renderList := func(initial bool) {
		if !initial {
			fmt.Printf("\033[%dA", totalRenderLines)
		}
		for i := 0; i < limit; i++ {
			entry := entries[i]
			projectName := "?"
			if entry.ProjectPath != "" {
				projectName = filepath.Base(entry.ProjectPath)
			}
			dateStr := "??-??"
			if entry.StartTime != "" {
				dateStr = parser.FormatTimestamp(entry.StartTime)
			}
			refMarker := ""
			if callerIDs[entry.SessionID] {
				refMarker = " [refs]"
			}
			prompt := entry.FirstPrompt
			if len(prompt) > 35 {
				prompt = prompt[:32] + "..."
			}
			prefix := "   "
			colorStart := ""
			colorEnd := ""
			if i == selected {
				prefix = " ❯ "
				colorStart = "\033[1;36m"
				colorEnd = "\033[0m"
			}
			fmt.Printf("\033[2K\r%s[%2d] %s%s  %s  %-15s %3dm  u:%-2d a:%-2d%s  %s%s\n",
				prefix, i+1, colorStart, entry.SessionID, dateStr, projectName, entry.DurationMinutes,
				entry.UserMessageCount, entry.AssistantMessageCount, refMarker, prompt, colorEnd)
		}
		fmt.Printf("\033[2K\r--------------------------------------------------------------------------------\n")
		fmt.Printf("\033[2K\r 使用 [↑ / ↓] 選擇會話，[Enter] 直接讀取，[Esc / b] 返回主選單\n")
	}

	renderList(true)

	for {
		ev := readKeyInteractive()
		switch ev.kind {
		case keyUp:
			selected = (selected - 1 + limit) % limit
			renderList(false)
		case keyDown:
			selected = (selected + 1) % limit
			renderList(false)
		case keyEnter:
			targetID := entries[selected].SessionID
			readAndDisplaySession(in, targetID, store, reader)
			return
		case keyEscape:
			return
		case keyRune:
			ch := strings.ToLower(string(ev.char))
			switch ch {
			case "k":
				selected = (selected - 1 + limit) % limit
				renderList(false)
			case "j":
				selected = (selected + 1) % limit
				renderList(false)
			case "b", "q":
				return
			default:
				if ch >= "1" && ch <= "9" {
					idx := int(ch[0] - '0')
					if idx <= limit {
						readAndDisplaySession(in, entries[idx-1].SessionID, store, reader)
						return
					}
				}
			}
		}
	}
}

func handleReadInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 讀取並壓縮指定會話 (cc-session read)")
	fmt.Print("請輸入 Session ID（輸入 'q' / 'b' 或直接按 Enter 跳出返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") || strings.EqualFold(id, "q") {
		return
	}

	readAndDisplaySession(in, id, store, reader)
}

func readAndDisplaySession(in *bufio.Reader, sessionID string, store parser.Store, reader claudecodec.Codec) {
	fmt.Printf("\n==> 會話讀取與壓縮結果: %s\n", sessionID)
	fmt.Println("================================================================================")

	var buf bytes.Buffer
	err := runRead([]string{sessionID}, &buf, os.Stderr, store, reader)
	if err != nil {
		fmt.Println("================================================================================")
		fmt.Printf("[錯誤] 讀取失敗: %v\n", err)
		return
	}

	completed := displayWithPaging(in, buf.String())
	fmt.Println("================================================================================")
	if !completed {
		return
	}

	for {
		fmt.Print("\n[快捷操作] 按 [C] 複製完整內容至剪貼簿  |  按 [Enter / Esc / q] 返回主選單 > ")
		if isInteractiveTerminal() {
			ev := readKeyInteractive()
			if ev.kind == keyEscape || ev.kind == keyEnter {
				fmt.Println("")
				break
			}
			if ev.kind == keyRune {
				ch := strings.ToLower(string(ev.char))
				if ch == "c" {
					if err := copyToClipboard(buf.String()); err != nil {
						fmt.Printf("\n剪貼簿複製失敗: %v\n", err)
					} else {
						fmt.Println("\n已成功複製完整壓縮內容至剪貼簿！")
					}
				} else if ch == "q" || ch == "b" {
					fmt.Println("")
					break
				}
			}
		} else {
			act := strings.ToLower(strings.TrimSpace(readLine(in)))
			if act == "c" {
				if err := copyToClipboard(buf.String()); err != nil {
					fmt.Printf("剪貼簿複製失敗: %v\n", err)
				} else {
					fmt.Println("已成功複製完整壓縮內容至剪貼簿！")
				}
			} else {
				break
			}
		}
	}
}

func handleContextInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 產生精簡 Context 標頭格式 (cc-session context)")
	fmt.Print("請輸入 Session ID（輸入 'q' / 'b' 或直接按 Enter 跳出返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") || strings.EqualFold(id, "q") {
		return
	}

	fmt.Printf("\n==> Context 標頭輸出: %s\n", id)
	fmt.Println("================================================================================")

	var buf bytes.Buffer
	err := runContext([]string{id}, &buf, os.Stderr, store, reader)
	if err != nil {
		fmt.Println("================================================================================")
		fmt.Printf("[錯誤] 產生失敗: %v\n", err)
		return
	}

	completed := displayWithPaging(in, buf.String())
	fmt.Println("================================================================================")
	if !completed {
		return
	}

	for {
		fmt.Print("\n[快捷操作] 按 [C] 複製至剪貼簿  |  按 [Enter / Esc / q] 返回主選單 > ")
		if isInteractiveTerminal() {
			ev := readKeyInteractive()
			if ev.kind == keyEscape || ev.kind == keyEnter {
				fmt.Println("")
				break
			}
			if ev.kind == keyRune {
				ch := strings.ToLower(string(ev.char))
				if ch == "c" {
					if err := copyToClipboard(buf.String()); err != nil {
						fmt.Printf("\n剪貼簿複製失敗: %v\n", err)
					} else {
						fmt.Println("\n已成功複製至剪貼簿！")
					}
				} else if ch == "q" || ch == "b" {
					fmt.Println("")
					break
				}
			}
		} else {
			act := strings.ToLower(strings.TrimSpace(readLine(in)))
			if act == "c" {
				if err := copyToClipboard(buf.String()); err != nil {
					fmt.Printf("剪貼簿複製失敗: %v\n", err)
				} else {
					fmt.Println("已成功複製至剪貼簿！")
				}
			} else {
				break
			}
		}
	}
}

func displayWithPaging(in *bufio.Reader, text string) bool {
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	total := len(lines)
	pageSize := 30

	if total <= pageSize || !isInteractiveTerminal() {
		for _, line := range lines {
			fmt.Println(line)
		}
		return true
	}

	current := 0
	showAll := false

	for current < total {
		end := current + pageSize
		if end > total {
			end = total
		}

		for i := current; i < end; i++ {
			fmt.Println(lines[i])
		}
		current = end

		if current >= total || showAll {
			break
		}

		fmt.Printf("\n-- 閱讀中 (%d/%d 行) -- [Enter/Space/↓] 繼續閱讀  |  [q/Esc] 中途跳出  |  [a] 全部顯示 > ", current, total)

		ev := readKeyInteractive()
		switch ev.kind {
		case keyEscape:
			fmt.Println("\n[已中途跳出，返回主選單]")
			return false
		case keyDown, keyEnter:
			// continue to next chunk
		case keyRune:
			ch := strings.ToLower(string(ev.char))
			switch ch {
			case "q", "b":
				fmt.Println("\n[已中途跳出，返回主選單]")
				return false
			case "a":
				showAll = true
			}
		}
		fmt.Println("")
	}
	return true
}

func handleBenchmarkInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 掃描近期會話，評估壓縮率與成本節省效益 (cc-session benchmark)")
	fmt.Println("正在掃描最近會話與計算... 請稍候...")
	fmt.Println("--------------------------------------------------------------------------------")

	err := runBenchmark([]string{"--n", "10"}, os.Stdout, os.Stderr, store, reader)
	fmt.Println("--------------------------------------------------------------------------------")
	if err != nil {
		fmt.Printf("[提示/錯誤] 評估結果: %v\n", err)
	}
	waitForEnter(in)
}

func handleStatsInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 會話字元與 Token 分佈統計 (cc-session stats)")
	fmt.Print("請輸入 Session ID（輸入 'q' / 'b' 或直接按 Enter 跳出返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") || strings.EqualFold(id, "q") {
		return
	}

	fmt.Printf("\n==> 統計結果: %s\n", id)
	fmt.Println("--------------------------------------------------------------------------------")
	err := runStats([]string{id}, os.Stdout, os.Stderr, store, reader)
	fmt.Println("--------------------------------------------------------------------------------")
	if err != nil {
		fmt.Printf("[錯誤] 統計失敗: %v\n", err)
	}
	waitForEnter(in)
}

func handleHelpInteractive(in *bufio.Reader) {
	fmt.Println("")
	cmdHelp(nil)
	waitForEnter(in)
}

func readLine(in *bufio.Reader) string {
	line, err := in.ReadString('\n')
	if err != nil && len(line) == 0 {
		return ""
	}
	return strings.TrimRight(line, "\r\n")
}

func waitForEnter(in *bufio.Reader) {
	if isInteractiveTerminal() {
		fmt.Print("\n請按 [Enter / Esc / q] 返回主選單...")
		ev := readKeyInteractive()
		_ = ev
		fmt.Println("")
		return
	}
	fmt.Print("\n請按 Enter 鍵返回...")
	_, _ = in.ReadString('\n')
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	} else {
		fmt.Print("\033[H\033[2J")
	}
}

func copyToClipboard(text string) error {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	if _, err := exec.LookPath("xclip"); err == nil {
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	}
	return fmt.Errorf("未找到剪貼簿命令 (clip/pbcopy/xclip)")
}
