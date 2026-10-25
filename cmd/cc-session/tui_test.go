package main

import (
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
