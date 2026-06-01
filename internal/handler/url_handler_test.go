package handler_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware" 
    "urlshortener/internal/model"
	"urlshortener/internal/apperror"
)

type mockURLService struct {
    shortenFn        func(originalURL string, userID string) (*model.URL, *apperror.AppError)
    getByShortCodeFn func(code string) (*model.URL, *apperror.AppError)
}

func (m *mockURLService) ShortenURL(originalURL string, userID string) (*model.URL, *apperror.AppError) {
    return m.shortenFn(originalURL, userID)
}

func (m *mockURLService) GetByShortCode(code string) (*model.URL, *apperror.AppError) {
    return m.getByShortCodeFn(code)
}

func setupRouter(svc handler.URLService) *gin.Engine {
    gin.SetMode(gin.TestMode)
    h := handler.NewURLHandler(svc, "http://localhost:8080")
    r := gin.New()

    fakeAuth := func(c *gin.Context) {
        c.Set(middleware.UserIDKey, "test-user-id")
        c.Next()
    }

    r.POST("/shorten", fakeAuth, h.Shorten)
    r.GET("/:code", h.Redirect)
    return r
}

// --- Tests for POST /shorten ---

func TestShorten_Success(t *testing.T) {
    mock := &mockURLService{
        shortenFn: func(originalURL string, userID string) (*model.URL, *apperror.AppError) {
            return &model.URL{
                OriginalURL: originalURL,
                ShortCode:   "abc123",
            }, nil
        },
    }

    r := setupRouter(mock)

    body := bytes.NewBufferString(`{"url": "https://example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/shorten", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    var resp map[string]string
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)

    assert.Contains(t, resp["short_url"], "http://localhost:8080/")
    assert.NotEmpty(t, resp["short_code"])
    assert.Equal(t, "https://example.com", resp["original_url"])
}

func TestShorten_MissingURL(t *testing.T) {
    mock := &mockURLService{}

    r := setupRouter(mock)

    body := bytes.NewBufferString(`{}`)
    req := httptest.NewRequest(http.MethodPost, "/shorten", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestShorten_RepoError(t *testing.T) {
    mock := &mockURLService{
        shortenFn: func(originalURL string, userID string) (*model.URL, *apperror.AppError) {
            return nil, apperror.Internal("could not create short url")
        },
    }

    r := setupRouter(mock)

    body := bytes.NewBufferString(`{"url": "https://example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/shorten", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- Tests for GET /:code ---

func TestRedirect_Success(t *testing.T) {
    mock := &mockURLService{
        getByShortCodeFn: func(code string) (*model.URL, *apperror.AppError) {
            return &model.URL{
                ID:          "some-uuid",
                ShortCode:   code,
                OriginalURL: "https://example.com",
            }, nil
        },
    }

    r := setupRouter(mock)

    req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusMovedPermanently, w.Code)
    assert.Equal(t, "https://example.com", w.Header().Get("Location"))
}

func TestRedirect_NotFound(t *testing.T) {
    mock := &mockURLService{
        getByShortCodeFn: func(code string) (*model.URL, *apperror.AppError) {
            return nil, apperror.NotFound("short url not found")
        },
    }

    r := setupRouter(mock)

    req := httptest.NewRequest(http.MethodGet, "/notexist", nil)
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusNotFound, w.Code)
}