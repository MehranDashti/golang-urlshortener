package util

import (
    "fmt"
    "net/url"
    "sort"
    "strings"
)

// NormaliseURL brings a URL to a canonical form:
// - scheme lowercased
// - host lowercased
// - trailing slash removed from path
// - query params sorted alphabetically
// - fragment removed (not useful for short links)
//
// Returns error if the URL cannot be parsed.
func NormaliseURL(rawURL string) (string, error) {
    // Parse the URL
    u, err := url.Parse(rawURL)
    if err != nil {
        return "", fmt.Errorf("invalid URL: %w", err)
    }

    // Validate scheme
    scheme := strings.ToLower(u.Scheme)
    if scheme != "http" && scheme != "https" {
        return "", fmt.Errorf(
            "unsupported scheme %q — only http and https allowed",
            u.Scheme)
    }
    u.Scheme = scheme

    // Lowercase host
    u.Host = strings.ToLower(u.Host)

    // Remove trailing slash from path (but keep root /)
    if len(u.Path) > 1 {
        u.Path = strings.TrimRight(u.Path, "/")
    }

    // Sort query parameters for consistency
    // ?b=2&a=1 → ?a=1&b=2
    if u.RawQuery != "" {
        params := u.Query()
        keys := make([]string, 0, len(params))
        for k := range params {
            keys = append(keys, k)
        }
        sort.Strings(keys)

        sorted := url.Values{}
        for _, k := range keys {
            sorted[k] = params[k]
        }
        u.RawQuery = sorted.Encode()
    }

    // Remove fragment — not useful for redirect targets
    u.Fragment = ""

    return u.String(), nil
}