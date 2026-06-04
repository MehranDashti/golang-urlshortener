package util_test

import (
	"testing"
	"urlshortener/internal/util"
)

func FuzzNormaliseURL(f *testing.F) {
	// Seed corpus — the fuzzer mutates these as starting points
	f.Add("https://google.com")
	f.Add("http://localhost:8080/path?q=1#anchor")
	f.Add("")
	f.Add("   ")
	f.Add("ftp://invalid-scheme.com")
	f.Add("https://")
	f.Add("not-a-url")
	f.Add("https://example.com?b=2&a=1")
	f.Add("HTTPS://EXAMPLE.COM/PATH/")

	f.Fuzz(func(t *testing.T, input string) {
		// Rule 1: must never panic on any input
		result, err := util.NormaliseURL(input)

		// Rule 2: if no error returned, result must be non-empty
		if err == nil && result == "" {
			t.Errorf("NormaliseURL(%q) returned empty string with no error", input)
		}

		// Rule 3: idempotent — normalising twice gives the same result
		if err == nil {
			result2, err2 := util.NormaliseURL(result)
			if err2 != nil {
				t.Errorf("NormaliseURL is not idempotent: second call on %q failed: %v", result, err2)
			}
			if result != result2 {
				t.Errorf("NormaliseURL is not idempotent: %q → %q → %q", input, result, result2)
			}
		}
	})
}