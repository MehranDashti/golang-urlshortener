package util

import (
    "math/rand"
    "time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const codeLength = 6

func GenerateShortCode() string {
    rng := rand.New(rand.NewSource(time.Now().UnixNano()))
    code := make([]byte, codeLength)
    for i := range code {
        code[i] = charset[rng.Intn(len(charset))]
    }
    return string(code)
}