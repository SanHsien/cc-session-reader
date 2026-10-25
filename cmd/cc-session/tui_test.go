package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncatePrompt_GivenLongChinesePrompt_ThenResultIsValidUTF8(t *testing.T) {
	prompt := strings.Repeat("修正登入頁面的錯誤", 5)

	got := truncatePrompt(prompt)

	if !utf8.ValidString(got) {
		t.Fatalf("truncatePrompt cut inside a multi-byte character: %q", got)
	}
	if n := utf8.RuneCountInString(got); n != 35 {
		t.Fatalf("truncated prompt should be 32 characters plus ellipsis (35), got %d", n)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("truncated prompt should end with ellipsis, got %q", got)
	}
}

func TestTruncatePrompt_GivenShortChinesePrompt_ThenUnchanged(t *testing.T) {
	// 20 characters = 60 bytes, but only 20 display characters.
	prompt := strings.Repeat("修正登入", 5)

	if got := truncatePrompt(prompt); got != prompt {
		t.Fatalf("short prompt was altered: %q", got)
	}
}

func TestTruncatePrompt_GivenASCIIWithinLimit_ThenUnchanged(t *testing.T) {
	prompt := strings.Repeat("a", 35)

	if got := truncatePrompt(prompt); got != prompt {
		t.Fatalf("prompt at the limit was altered: %q", got)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	done := make(chan string)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()
	defer func() { os.Stdout = orig }()
	fn()
	w.Close()
	return <-done
}

func numberedLines(n int) string {
	var sb strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&sb, "line-%03d\n", i)
	}
	return sb.String()
}

func stubKeys(t *testing.T, keys ...keyEvent) {
	t.Helper()
	origI, origK := isInteractive, readKey
	t.Cleanup(func() { isInteractive, readKey = origI, origK })
	isInteractive = func() bool { return true }
	i := 0
	readKey = func() keyEvent {
		if i >= len(keys) {
			t.Fatalf("readKey called more than the %d scripted times", len(keys))
		}
		k := keys[i]
		i++
		return k
	}
}

func TestDisplayWithPaging_GivenShowAllKey_ThenEveryLineIsPrinted(t *testing.T) {
	stubKeys(t, keyEvent{kind: keyRune, char: 'a'})
	text := numberedLines(100)

	var completed bool
	out := captureStdout(t, func() {
		completed = displayWithPaging(bufio.NewReader(strings.NewReader("")), text)
	})

	if !completed {
		t.Fatal("showing everything should report completion")
	}
	if strings.Count(out, "line-031") != 1 {
		t.Fatal("lines after the first page must be printed exactly once")
	}
	for _, want := range []string{"line-001", "line-060", "line-100"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output is missing %s after pressing [a]; later lines were dropped", want)
		}
	}
}

func TestDisplayWithPaging_GivenContinueKeys_ThenEveryLineIsPrinted(t *testing.T) {
	stubKeys(t, keyEvent{kind: keyEnter}, keyEvent{kind: keyEnter}, keyEvent{kind: keyEnter})

	var completed bool
	out := captureStdout(t, func() {
		completed = displayWithPaging(bufio.NewReader(strings.NewReader("")), numberedLines(100))
	})

	if !completed {
		t.Fatal("paging to the end should report completion")
	}
	if !strings.Contains(out, "line-100") {
		t.Fatal("paging with Enter never reached the last line")
	}
}

func TestDisplayWithPaging_GivenEscape_ThenStopsEarlyAndReportsIncomplete(t *testing.T) {
	stubKeys(t, keyEvent{kind: keyEscape})

	var completed bool
	out := captureStdout(t, func() {
		completed = displayWithPaging(bufio.NewReader(strings.NewReader("")), numberedLines(100))
	})

	if completed {
		t.Fatal("Esc should report an incomplete read")
	}
	if strings.Contains(out, "line-031") {
		t.Fatal("output continued past the first page after Esc")
	}
}
