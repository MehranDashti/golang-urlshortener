//go:build integration

package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"urlshortener/tests/testhelper"
	"urlshortener/tests/testserver"
)

func TestHealthCheck(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	w := testhelper.MakeRequest(s.Router,
		http.MethodGet, "/health", "", "")

	assert.Equal(t, http.StatusOK, w.Code)

	body := testhelper.ParseBody(w)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "ok", body["db"])
}

func TestHealthCheck_ReturnsVersion(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	w := testhelper.MakeRequest(s.Router,
		http.MethodGet, "/health", "", "")

	body := testhelper.ParseBody(w)
	assert.NotEmpty(t, body["version"])
}

func TestTraceID_InResponse(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	w := testhelper.MakeRequest(s.Router,
		http.MethodGet, "/health", "", "")

	// Every response should have a trace ID header
	traceID := w.Header().Get("X-Trace-ID")
	assert.NotEmpty(t, traceID,
		"response should include X-Trace-ID header")
	assert.Len(t, traceID, 8,
		"trace ID should be 8 characters")
}

func TestTraceID_PropagatesFromRequest(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	// Send a custom trace ID
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Trace-ID", "custom12")
	w := httptest.NewRecorder()
	s.Router.ServeHTTP(w, req)

	// Same ID should come back
	assert.Equal(t, "custom12",
		w.Header().Get("X-Trace-ID"))
}
