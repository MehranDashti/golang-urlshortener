//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"urlshortener/tests/testhelper"
	"urlshortener/tests/testserver"
)

// loginUser is a helper that signs up and logs in, returns access token
func loginUser(t *testing.T, s *testserver.TestServer, email, password string) string {
	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup",
		`{"email":"`+email+`","password":"`+password+`"}`, "")

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/login",
		`{"email":"`+email+`","password":"`+password+`"}`, "")

	data := testhelper.GetData(w)
	require.NotNil(t, data)
	return data["access_token"].(string)
}

func TestShorten_Integration_Success(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	token := loginUser(t, s, "test@example.com", "123456")

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/client/shorten",
		`{"url": "https://google.com"}`, token)

	assert.Equal(t, http.StatusCreated, w.Code)

	data := testhelper.GetData(w)
	require.NotNil(t, data)
	assert.NotEmpty(t, data["short_code"])
	assert.Contains(t, data["short_url"], "http://localhost:8080/")
}

func TestShorten_Integration_Unauthorized(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/client/shorten",
		`{"url": "https://google.com"}`, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestListLinks_Integration(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	token := loginUser(t, s, "test@example.com", "123456")

	// Create two links
	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/client/shorten",
		`{"url": "https://google.com"}`, token)
	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/client/shorten",
		`{"url": "https://github.com"}`, token)

	w := testhelper.MakeRequest(s.Router, http.MethodGet,
		"/api/v1/client/links", "", token)

	assert.Equal(t, http.StatusOK, w.Code)

	body := testhelper.ParseBody(w)
	links := body["data"].([]interface{})
	assert.Len(t, links, 2)
}

func TestRedirect_Integration_Expired(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	token := loginUser(t, s, "test@example.com", "123456")

	// Create a link that expired in the past
	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/client/shorten",
		`{"url": "https://google.com", "expires_at": "2020-01-01T00:00:00Z"}`, token)

	data := testhelper.GetData(w)
	shortCode := data["short_code"].(string)

	// Try to redirect — should get 410 Gone
	w2 := testhelper.MakeRequest(s.Router, http.MethodGet,
		"/api/v1/"+shortCode, "", "")

	assert.Equal(t, http.StatusGone, w2.Code)
}

func TestShorten_CollisionRetry(t *testing.T) {
	s := testserver.New()
	defer s.Close()

	token := loginUser(t, s, "test@example.com", "123456")

	// Create 50 short links rapidly — collision probability increases
	// with more links but GenerateShortCode should handle it
	for i := 0; i < 50; i++ {
		w := testhelper.MakeRequest(s.Router,
			http.MethodPost, "/api/v1/client/shorten",
			`{"url":"https://google.com"}`, token)
		assert.Equal(t, http.StatusCreated, w.Code,
			"request %d should succeed", i)
	}
}
