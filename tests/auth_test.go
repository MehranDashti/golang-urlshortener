//go:build integration

package tests

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"urlshortener/tests/testhelper"
	"urlshortener/tests/testserver"
)

// TestMain runs before all tests in this package.
// goleak checks for goroutine leaks after every test.
func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestAuthSignup_Success(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	assert.Equal(t, http.StatusCreated, w.Code)

	data := testhelper.GetData(w)
	require.NotNil(t, data)
	assert.Equal(t, "test@example.com", data["email"])
	assert.NotEmpty(t, data["id"])
}

func TestAuthSignup_DuplicateEmail(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	// First signup
	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	// Second signup with same email
	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthLogin_Success(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	// Signup first
	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	// Login
	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/login", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	assert.Equal(t, http.StatusOK, w.Code)

	data := testhelper.GetData(w)
	require.NotNil(t, data)
	assert.NotEmpty(t, data["access_token"])
	assert.NotEmpty(t, data["refresh_token"])
}

func TestAuthLogin_WrongPassword(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/login", `{
        "email": "test@example.com",
        "password": "wrongpassword"
    }`, "")

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRefresh_Success(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/signup", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	w := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/login", `{
        "email": "test@example.com",
        "password": "123456"
    }`, "")

	data := testhelper.GetData(w)
	refreshToken := data["refresh_token"].(string)

	w2 := testhelper.MakeRequest(s.Router, http.MethodPost, "/api/v1/auth/refresh",
		`{"refresh_token": "`+refreshToken+`"}`, "")

	assert.Equal(t, http.StatusOK, w2.Code)

	data2 := testhelper.GetData(w2)
	// New tokens must exist and be non-empty
	assert.NotEmpty(t, data2["access_token"])
	assert.NotEmpty(t, data2["refresh_token"])

	// New access token must be different from the refresh token
	// (they are different token types even if timestamps match)
	assert.NotEqual(t, data2["access_token"], data2["refresh_token"])
}

func TestLogout_TokenRevoked(t *testing.T) {
	s := testserver.New()
	defer s.CleanDB()

	// Signup + login
	testhelper.MakeRequest(s.Router, http.MethodPost,
		"/api/v1/auth/signup",
		`{"email":"logout@test.com","password":"123456"}`, "")

	w := testhelper.MakeRequest(s.Router, http.MethodPost,
		"/api/v1/auth/login",
		`{"email":"logout@test.com","password":"123456"}`, "")

	data := testhelper.GetData(w)
	accessToken := data["access_token"].(string)

	// Verify token works before logout
	w2 := testhelper.MakeRequest(s.Router, http.MethodGet,
		"/api/v1/client/links", "", accessToken)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Logout
	w3 := testhelper.MakeRequest(s.Router, http.MethodPost,
		"/api/v1/auth/logout", "", accessToken)
	assert.Equal(t, http.StatusOK, w3.Code)

	// Token should now be rejected
	w4 := testhelper.MakeRequest(s.Router, http.MethodGet,
		"/api/v1/client/links", "", accessToken)
	assert.Equal(t, http.StatusUnauthorized, w4.Code)

	body := testhelper.ParseBody(w4)
	assert.Equal(t, "token has been revoked", body["message"])
}
