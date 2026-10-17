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

// cmdTUI launches the interactive terminal user interface.
func cmdTUI(args []string, reader claudecodec.Codec) {
	runInteractiveMenu(reader)
}

func runInteractiveMenu(reader claudecodec.Codec) {
	inReader := bufio.NewReader(os.Stdin)
	store := parser.DefaultStoreWith(reader)

	commitSuffix := ""
	if commit != "none" && commit != "" {
		short := commit
		if len(short) > 7 {
			short = short[:7]
		}
		commitSuffix = fmt.Sprintf(" (%s)", short)
	}
	fmt.Println("============================================================")
	fmt.Printf("  cc-session Reader %s%s - 終端互動面板 (TUI)\n", version, commitSuffix)
	fmt.Println("============================================================")

	for {
		entries, warnings := store.ListAllSessions()
		for _, w := range warnings {
			fmt.Fprintf(os.Stderr, "警告: %s\n", w)
		}

		limit := 15
		if len(entries) < limit {
			limit = len(entries)
		}

		var action string
		if isInteractiveTerminal() {
			action = runSessionWorkspaceInteractive(inReader, entries[:limit], store, reader)
		} else {
			action = runSessionWorkspaceNonInteractive(inReader, entries[:limit], store, reader)
		}

		if action == "exit" {
			fmt.Println("\n謝謝使用 cc-session！")
			pauseIfAlone()
			return
		}
	}
}

func runSessionWorkspaceInteractive(in *bufio.Reader, entries []parser.SessionListEntry, store parser.Store, reader claudecodec.Codec) string {
	fmt.Println("\n==> 最近的 Claude Code 會話列表：")
	fmt.Println("--------------------------------------------------------------------------------")

	limit := len(entries)
	callerIDs := tracker.CallerSessionIDs()

	if limit == 0 {
		fmt.Println("未找到任何 Claude Code 會話。")
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println(" 操作選項: [M] 手動輸入 Session ID  |  [B] 成本評估  |  [H] 命令速查  |  [Q] 離開")
		fmt.Print("請選擇操作 > ")
		choice := strings.ToLower(strings.TrimSpace(readLine(in)))
		switch choice {
		case "m":
			handleManualSessionInteractive(in, store, reader)
			return "continue"
		case "b":
			handleBenchmarkInteractive(in, store, reader)
			return "continue"
		case "h":
			handleHelpInteractive(in)
			return "continue"
		case "q", "0", "exit":
			return "exit"
		default:
			return "continue"
		}
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
		fmt.Printf("\033[2K\r [↑/↓] 選擇會話 | [Enter] 讀取 | [C] Context | [S] 統計 | [Y] 複製ID | [B] 評估 | [Q] 離開\n")
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
			readAndDisplaySession(in, entries[selected].SessionID, store, reader)
			return "continue"
		case keyEscape:
			return "exit"
		case keyRune:
			ch := strings.ToLower(string(ev.char))
			switch ch {
			case "k":
				selected = (selected - 1 + limit) % limit
				renderList(false)
			case "j":
				selected = (selected + 1) % limit
				renderList(false)
			case "c":
				handleContextForSession(in, entries[selected].SessionID, store, reader)
				return "continue"
			case "s":
				handleStatsForSession(in, entries[selected].SessionID, store, reader)
				return "continue"
			case "y", "i":
				_ = copyToClipboard(entries[selected].SessionID)
				fmt.Printf("\n已成功複製 Session ID 至剪貼簿：%s\n", entries[selected].SessionID)
				waitForEnter(in)
				return "continue"
			case "b":
				handleBenchmarkInteractive(in, store, reader)
				return "continue"
			case "h", "?":
				handleHelpInteractive(in)
				return "continue"
			case "m", "f":
				handleManualSessionInteractive(in, store, reader)
				return "continue"
			case "cls":
				clearScreen()
				return "continue"
			case "q":
				return "exit"
			default:
				if ch >= "1" && ch <= "9" {
					idx := int(ch[0] - '0')
					if idx <= limit {
						readAndDisplaySession(in, entries[idx-1].SessionID, store, reader)
						return "continue"
					}
				}
			}
		}
	}
}

func runSessionWorkspaceNonInteractive(in *bufio.Reader, entries []parser.SessionListEntry, store parser.Store, reader claudecodec.Codec) string {
	fmt.Println("\n==> 最近的 Claude Code 會話列表：")
	fmt.Println("--------------------------------------------------------------------------------")

	limit := len(entries)
	callerIDs := tracker.CallerSessionIDs()

	if limit == 0 {
		fmt.Println("未找到任何 Claude Code 會話。")
		fmt.Println("--------------------------------------------------------------------------------")
		fmt.Println(" 操作選項: [M] 手動輸入 Session ID  |  [B] 成本評估  |  [H] 命令速查  |  [Q] 離開")
		fmt.Print("請選擇操作 > ")
	} else {
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
		fmt.Println(" 操作選項: [<序號> 讀取] | [C <序號> Context] | [S <序號> 統計] | [Y <序號> 複製ID] | [B 評估] | [M 手動輸入] | [Q 離開]")
		fmt.Print("請輸入操作 > ")
	}

	input := strings.TrimSpace(readLine(in))
	if input == "" || input == "0" || strings.EqualFold(input, "q") || strings.EqualFold(input, "exit") {
		return "exit"
	}

	parts := strings.Fields(input)
	cmd := strings.ToLower(parts[0])

	switch cmd {
	case "b", "benchmark":
		handleBenchmarkInteractive(in, store, reader)
		return "continue"
	case "h", "help", "?":
		handleHelpInteractive(in)
		return "continue"
	case "m", "manual":
		handleManualSessionInteractive(in, store, reader)
		return "continue"
	case "y", "copy":
		idx := 1
		if len(parts) > 1 {
			if parsed, err := strconv.Atoi(parts[1]); err == nil && parsed >= 1 && parsed <= limit {
				idx = parsed
			}
		}
		if limit > 0 && idx <= limit {
			_ = copyToClipboard(entries[idx-1].SessionID)
			fmt.Printf("\n已成功複製 Session ID 至剪貼簿：%s\n", entries[idx-1].SessionID)
			waitForEnter(in)
		}
		return "continue"
	case "c", "context":
		if len(parts) > 1 {
			if idx, err := strconv.Atoi(parts[1]); err == nil && idx >= 1 && idx <= limit {
				handleContextForSession(in, entries[idx-1].SessionID, store, reader)
				return "continue"
			}
		}
		handleManualSessionInteractive(in, store, reader)
		return "continue"
	case "s", "stats":
		if len(parts) > 1 {
			if idx, err := strconv.Atoi(parts[1]); err == nil && idx >= 1 && idx <= limit {
				handleStatsForSession(in, entries[idx-1].SessionID, store, reader)
				return "continue"
			}
		}
		handleManualSessionInteractive(in, store, reader)
		return "continue"
	default:
		if idx, err := strconv.Atoi(cmd); err == nil && idx >= 1 && idx <= limit {
			readAndDisplaySession(in, entries[idx-1].SessionID, store, reader)
			return "continue"
		}
	}
	return "continue"
}

func handleManualSessionInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	fmt.Println("\n==> 手動輸入 Session ID 查詢 (支援前綴如 e61060b1)")
	fmt.Print("請輸入 Session ID（按 [Enter / Esc / q] 取消返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") || strings.EqualFold(id, "q") {
		return
	}

	fmt.Println("操作選項: [1/Enter] 讀取壓縮  |  [2/C] Context標頭  |  [3/S] Token統計  |  [Q] 返回")
	fmt.Print("請選擇操作 > ")
	act := strings.ToLower(strings.TrimSpace(readLine(in)))
	switch act {
	case "2", "c":
		handleContextForSession(in, id, store, reader)
	case "3", "s":
		handleStatsForSession(in, id, store, reader)
	case "q", "0", "b":
		return
	default:
		readAndDisplaySession(in, id, store, reader)
	}
}

func readAndDisplaySession(in *bufio.Reader, sessionID string, store parser.Store, reader claudecodec.Codec) {
	fmt.Printf("\n==> 會話讀取與壓縮結果: %s\n", sessionID)
	fmt.Println("提示: 長會話自動分頁串流，閱讀中途可隨時按 [q / Esc] 跳出返回")
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
		fmt.Print("\n[快捷操作] 按 [C] 複製完整內容至剪貼簿  |  按 [Enter / Esc / q] 返回會話列表 > ")
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

func handleContextForSession(in *bufio.Reader, id string, store parser.Store, reader claudecodec.Codec) {
	fmt.Printf("\n==> Context 標頭輸出: %s\n", id)
	fmt.Println("提示: 長內容自動分頁串流，閱讀中途可隨時按 [q / Esc] 跳出返回")
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
		fmt.Print("\n[快捷操作] 按 [C] 複製至剪貼簿  |  按 [Enter / Esc / q] 返回會話列表 > ")
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

func handleStatsForSession(in *bufio.Reader, id string, store parser.Store, reader claudecodec.Codec) {
	fmt.Printf("\n==> 統計結果: %s\n", id)
	fmt.Println("--------------------------------------------------------------------------------")
	err := runStats([]string{id}, os.Stdout, os.Stderr, store, reader)
	fmt.Println("--------------------------------------------------------------------------------")
	if err != nil {
		fmt.Printf("[錯誤] 統計失敗: %v\n", err)
	}
	waitForEnter(in)
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
			fmt.Println("\n[已中途跳出，返回會話列表]")
			return false
		case keyDown, keyEnter:
			// continue to next chunk
		case keyRune:
			ch := strings.ToLower(string(ev.char))
			switch ch {
			case "q", "b":
				fmt.Println("\n[已中途跳出，返回會話列表]")
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
		fmt.Print("\n請按 [Enter / Esc / q] 返回會話列表...")
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
