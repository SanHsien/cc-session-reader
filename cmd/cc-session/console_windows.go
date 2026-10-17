//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                 = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleProcessList = kernel32.NewProc("GetConsoleProcessList")
)

// isAloneOnConsole checks whether this process was launched in its own
// dedicated console window (e.g., by double-clicking cc-session.exe in File Explorer)
// rather than being invoked from an existing shell like PowerShell or cmd.
func isAloneOnConsole() bool {
	if procGetConsoleProcessList.Find() != nil {
		return false
	}
	var pids [2]uint32
	r, _, _ := procGetConsoleProcessList.Call(
		uintptr(unsafe.Pointer(&pids[0])),
		uintptr(len(pids)),
	)
	return r == 1
}

func pauseIfAlone() {
	if isAloneOnConsole() {
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "============================================================")
		fmt.Fprintln(os.Stderr, " [提示] cc-session 為命令列工具 (CLI)，請在 PowerShell 或 CMD 中使用。")
		fmt.Fprintln(os.Stderr, " 常用指令範例：")
		fmt.Fprintln(os.Stderr, "   cc-session list      列出近期 session")
		fmt.Fprintln(os.Stderr, "   cc-session help      檢視完整子命令速查表")
		fmt.Fprintln(os.Stderr, "============================================================")
		fmt.Fprint(os.Stderr, "請按 Enter 鍵關閉此視窗...")
		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadBytes('\n')
	}
}
