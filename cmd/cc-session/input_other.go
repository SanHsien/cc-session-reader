//go:build !windows

package main

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
	return keyEvent{kind: keyUnknown}
}
