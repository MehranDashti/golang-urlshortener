package handler_test

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware" 
    "urlshortener/internal/model"
	"urlshortener/internal/apperror"
)

type mockURLService struct {
    shortenFn                func(ctx context.Context, originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError)
    getByShortCodeFn         func(ctx context.Context, code string) (*model.URL, *apperror.AppError)
    getUserLinksFn           func(ctx context.Context, userID string) ([]*model.URL, *apperror.AppError)
    getUserLinksPaginatedFn  func(ctx context.Context, userID string, params model.PaginationParams) (*model.PaginatedResult, *apperror.AppError) // ← add
}

func (m *mockURLService) ShortenURL(ctx context.Context, originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError) {
    return m.shortenFn(ctx, originalURL, userID, expiresAt)
}

func (m *mockURLService) GetByShortCode(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
    return m.getByShortCodeFn(ctx, code)
}

func (m *mockURLService) GetUserLinks(ctx context.Context, userID string) ([]*model.URL, *apperror.AppError) {
    if m.getUserLinksFn != nil {
        return m.getUserLinksFn(ctx, userID)
    }
    return nil, nil
}

func (m *mockURLService) GetUserLinksPaginated(ctx context.Context, userID string, params model.PaginationParams) (*model.PaginatedResult, *apperror.AppError) {
    if m.getUserLinksPaginatedFn != nil {
        return m.getUserLinksPaginatedFn(ctx, userID, params)
    }
    return nil, nil
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
        shortenFn: func(ctx context.Context, originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError) {
            return &model.URL{OriginalURL: originalURL, ShortCode: "abc123"}, nil
        },
    }

    r := setupRouter(mock)

    body := bytes.NewBufferString(`{"url": "https://example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/shorten", body)
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)

    // New response shape — use map[string]interface{} since values are mixed types
    var resp map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &resp)
    require.NoError(t, err)

    // Check top level
    assert.Equal(t, true, resp["success"])

    // data is a nested object — type assert it to map[string]interface{}
    data, ok := resp["data"].(map[string]interface{})
    require.True(t, ok, "data should be an object")

    assert.Contains(t, data["short_url"], "http://localhost:8080/")
    assert.NotEmpty(t, data["short_code"])
    assert.Equal(t, "https://example.com", data["original_url"])
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
        shortenFn: func(ctx context.Context, originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError) {
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
        getByShortCodeFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
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
        getByShortCodeFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
            return nil, apperror.NotFound("short url not found")
        },
    }

    r := setupRouter(mock)

    req := httptest.NewRequest(http.MethodGet, "/notexist", nil)
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, http.StatusNotFound, w.Code)
}