package util_test

import (
	"strings"
	"testing"
	"urlshortener/internal/util"
)

const expectedCodeLength = 6
const validCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func FuzzGenerateShortCode(f *testing.F) {
	// GenerateShortCode takes no input — use a dummy seed
	// so the fuzzer has something to vary between calls
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))

	f.Fuzz(func(t *testing.T, _ int64) {
		code := util.GenerateShortCode()

		// Rule 1: must always return exactly 6 characters
		if len(code) != expectedCodeLength {
			t.Errorf("GenerateShortCode() length = %d, want %d", len(code), expectedCodeLength)
		}

		// Rule 2: every character must be from the valid charset
		for _, ch := range code {
			if !strings.ContainsRune(validCharset, ch) {
				t.Errorf("GenerateShortCode() contains invalid char %q in %q", ch, code)
			}
		}
	})
}
