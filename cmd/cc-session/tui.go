package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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

	for {
		clearScreen()
		fmt.Println("============================================================")
		fmt.Printf("  cc-session Reader %s - 終端互動面板 (TUI)\n", version)
		fmt.Println("============================================================")
		fmt.Println(" 請輸入選項編號 [1-6]，或輸入 [0] / [q] 離開：")
		fmt.Println("")
		fmt.Println("   [1] 📋 列出最近會話 (list)")
		fmt.Println("   [2] 🔍 讀取並壓縮指定會話 (read)")
		fmt.Println("   [3] 📦 產生精簡 Context 標頭格式 (context)")
		fmt.Println("   [4] 📊 Token 壓縮與成本評估 (benchmark)")
		fmt.Println("   [5] 📈 會話字元與 Token 統計 (stats)")
		fmt.Println("   [6] 🛠 檢視完整 CLI 子命令速查表 (help)")
		fmt.Println("   [0] 🚪 離開 (exit)")
		fmt.Println("------------------------------------------------------------")
		fmt.Print("請選擇操作 > ")

		choice := readLine(inReader)
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
		case "0", "q", "quit", "exit":
			fmt.Println("\n謝謝使用 cc-session！")
			pauseIfAlone()
			return
		default:
			fmt.Println("\n[無效輸入] 請輸入 0 到 6 的數字。")
			waitForEnter(inReader)
		}
	}
}

func handleListInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	clearScreen()
	fmt.Println("==> 最近的 Claude Code 會話列表：")
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
		waitForEnter(in)
		return
	}

	callerIDs := tracker.CallerSessionIDs()
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
		targetID := entries[idx-1].SessionID
		readAndDisplaySession(in, targetID, store, reader)
		return
	}
}

func handleReadInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	clearScreen()
	fmt.Println("==> 讀取並壓縮指定會話 (cc-session read)")
	fmt.Println("提示: 可輸入完整 Session ID 或前綴（如 e61060b1），輸入 'b' 返回")
	fmt.Print("請輸入 Session ID > ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") {
		return
	}

	readAndDisplaySession(in, id, store, reader)
}

func readAndDisplaySession(in *bufio.Reader, sessionID string, store parser.Store, reader claudecodec.Codec) {
	clearScreen()
	fmt.Printf("==> 會話讀取與壓縮結果: %s\n", sessionID)
	fmt.Println("================================================================================")

	var buf bytes.Buffer
	multiOut := io.MultiWriter(os.Stdout, &buf)

	err := runRead([]string{sessionID}, multiOut, os.Stderr, store, reader)
	fmt.Println("================================================================================")
	if err != nil {
		fmt.Printf("[錯誤] 讀取失敗: %v\n", err)
		waitForEnter(in)
		return
	}

	for {
		fmt.Print("\n[快捷操作] 按 [C] 複製完整內容至剪貼簿  |  按 [Enter] 返回主選單 > ")
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

func handleContextInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	clearScreen()
	fmt.Println("==> 產生精簡 Context 標頭格式 (cc-session context)")
	fmt.Print("請輸入 Session ID（輸入 'b' 返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") {
		return
	}

	clearScreen()
	fmt.Printf("==> Context 標頭輸出: %s\n", id)
	fmt.Println("================================================================================")

	var buf bytes.Buffer
	multiOut := io.MultiWriter(os.Stdout, &buf)

	err := runContext([]string{id}, multiOut, os.Stderr, store, reader)
	fmt.Println("================================================================================")
	if err != nil {
		fmt.Printf("[錯誤] 產生失敗: %v\n", err)
		waitForEnter(in)
		return
	}

	for {
		fmt.Print("\n[快捷操作] 按 [C] 複製至剪貼簿  |  按 [Enter] 返回主選單 > ")
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

func handleBenchmarkInteractive(in *bufio.Reader, store parser.Store, reader claudecodec.Codec) {
	clearScreen()
	fmt.Println("==> 掃描近期會話，評估壓縮率與成本節省效益 (cc-session benchmark)")
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
	clearScreen()
	fmt.Println("==> 會話字元與 Token 分佈統計 (cc-session stats)")
	fmt.Print("請輸入 Session ID（輸入 'b' 返回）> ")

	id := strings.TrimSpace(readLine(in))
	if id == "" || strings.EqualFold(id, "b") {
		return
	}

	clearScreen()
	fmt.Printf("==> 統計結果: %s\n", id)
	fmt.Println("--------------------------------------------------------------------------------")
	err := runStats([]string{id}, os.Stdout, os.Stderr, store, reader)
	fmt.Println("--------------------------------------------------------------------------------")
	if err != nil {
		fmt.Printf("[錯誤] 統計失敗: %v\n", err)
	}
	waitForEnter(in)
}

func handleHelpInteractive(in *bufio.Reader) {
	clearScreen()
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
	fmt.Print("\n請按 Enter 鍵返回...")
	_, _ = in.ReadString('\n')
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		// Windows terminal ANSI clear or fallback
		fmt.Print("\033[H\033[2J")
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
