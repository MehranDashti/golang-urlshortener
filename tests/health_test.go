//go:build integration

package tests

import (
    "net/http"
    "testing"

    "github.com/stretchr/testify/assert"
    "urlshortener/tests/testhelper"
    "urlshortener/tests/testserver"
)

func TestHealthCheck(t *testing.T) {
    s := testserver.New()
    defer s.CleanDB()

    w := testhelper.MakeRequest(s.Router,
        http.MethodGet, "/health", "", "")

    assert.Equal(t, http.StatusOK, w.Code)

    body := testhelper.ParseBody(w)
    assert.Equal(t, "ok", body["status"])
    assert.Equal(t, "ok", body["db"])
}

func TestHealthCheck_ReturnsVersion(t *testing.T) {
    s := testserver.New()
    defer s.CleanDB()

    w := testhelper.MakeRequest(s.Router,
        http.MethodGet, "/health", "", "")

    body := testhelper.ParseBody(w)
    assert.NotEmpty(t, body["version"])
}