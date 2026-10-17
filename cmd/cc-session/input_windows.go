//go:build windows

package main

import (
	"syscall"
)

var (
	msvcrt    = syscall.NewLazyDLL("msvcrt.dll")
	procGetch = msvcrt.NewProc("_getch")
	procKbhit = msvcrt.NewProc("_kbhit")
)

type keyType int

const (
	keyUnknown keyType = iota
	keyUp
	keyDown
	keyEnter
	keyEscape
	keyRune
)

type keyEvent struct {
	kind keyType
	char rune
}

func readKeyInteractive() keyEvent {
	r, _, _ := procGetch.Call()
	ch := int(r)
	if ch == 0 || ch == 224 {
		r2, _, _ := procGetch.Call()
		switch int(r2) {
		case 72:
			return keyEvent{kind: keyUp}
		case 80:
			return keyEvent{kind: keyDown}
		default:
			return keyEvent{kind: keyUnknown}
		}
	}
	switch ch {
	case 13, 10:
		return keyEvent{kind: keyEnter}
	case 27:
		rHit, _, _ := procKbhit.Call()
		if rHit != 0 {
			rNext, _, _ := procGetch.Call()
			if int(rNext) == '[' {
				rCode, _, _ := procGetch.Call()
				switch int(rCode) {
				case 'A':
					return keyEvent{kind: keyUp}
				case 'B':
					return keyEvent{kind: keyDown}
				}
			}
		}
		return keyEvent{kind: keyEscape}
	case 3: // Ctrl+C
		return keyEvent{kind: keyEscape}
	}
	return keyEvent{kind: keyRune, char: rune(ch)}
}
