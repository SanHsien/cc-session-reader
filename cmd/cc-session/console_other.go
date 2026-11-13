//go:build !windows

package main

func pauseIfAlone() {}

func enableVirtualTerminal() (bool, func()) { return true, func() {} }
