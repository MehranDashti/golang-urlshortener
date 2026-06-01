package util_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "urlshortener/internal/util"
)

func TestGenerateShortCode_Length(t *testing.T) {
    code := util.GenerateShortCode()
    assert.Len(t, code, 6)
}

func TestGenerateShortCode_ValidCharacters(t *testing.T) {
    code := util.GenerateShortCode()
    for _, c := range code {
        assert.Contains(t,
            "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
            string(c),
        )
    }
}

func TestGenerateShortCode_Uniqueness(t *testing.T) {
    codes := make(map[string]bool)
    for i := 0; i < 1000; i++ {
        code := util.GenerateShortCode()
        codes[code] = true
    }
    assert.Greater(t, len(codes), 990)
}