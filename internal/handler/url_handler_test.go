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
    "urlshortener/internal/apperror"
    "urlshortener/internal/handler"
    "urlshortener/internal/middleware"
    "urlshortener/internal/model"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockURLService struct {
    shortenFn               func(ctx context.Context, originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError)
    getByShortCodeFn        func(ctx context.Context, code string) (*model.URL, *apperror.AppError)
    getUserLinksFn          func(ctx context.Context, userID string) ([]*model.URL, *apperror.AppError)
    getUserLinksPaginatedFn func(ctx context.Context, userID string, params model.PaginationParams) (*model.PaginatedResult[*model.URL], *apperror.AppError)
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
func (m *mockURLService) GetUserLinksPaginated(ctx context.Context, userID string, params model.PaginationParams) (*model.PaginatedResult[*model.URL], *apperror.AppError) {
    if m.getUserLinksPaginatedFn != nil {
        return m.getUserLinksPaginatedFn(ctx, userID, params)
    }
    return nil, nil
}

// ── Router setup ──────────────────────────────────────────────────────────────

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

// ── Table-driven: POST /shorten ───────────────────────────────────────────────

func TestShorten(t *testing.T) {
    tests := []struct {
        name       string
        body       string
        mockFn     func(ctx context.Context, url, userID string, exp *time.Time) (*model.URL, *apperror.AppError)
        wantStatus int
        wantData   bool // whether to check data fields
    }{
        {
            name: "success",
            body: `{"url": "https://example.com"}`,
            mockFn: func(ctx context.Context, url, userID string, exp *time.Time) (*model.URL, *apperror.AppError) {
                return &model.URL{
                    OriginalURL: url,
                    ShortCode:   "abc123",
                }, nil
            },
            wantStatus: http.StatusCreated,
            wantData:   true,
        },
        {
            name:       "missing url",
            body:       `{}`,
            mockFn:     nil, // never reaches service
            wantStatus: http.StatusBadRequest,
            wantData:   false,
        },
        {
            name: "invalid url format",
            body: `{"url": "not-a-url"}`,
            mockFn: nil, // caught by validator
            wantStatus: http.StatusBadRequest,
            wantData:   false,
        },
        {
            name: "service error",
            body: `{"url": "https://example.com"}`,
            mockFn: func(ctx context.Context, url, userID string, exp *time.Time) (*model.URL, *apperror.AppError) {
                return nil, apperror.Internal("db is down")
            },
            wantStatus: http.StatusInternalServerError,
            wantData:   false,
        },
    }

    for _, tt := range tests {
        // tt captured per iteration — important for parallel tests
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockURLService{
                shortenFn: func(ctx context.Context, url, userID string, exp *time.Time) (*model.URL, *apperror.AppError) {
                    if tt.mockFn != nil {
                        return tt.mockFn(ctx, url, userID, exp)
                    }
                    t.Fatal("shortenFn called but not expected")
                    return nil, nil
                },
            }

            r := setupRouter(mock)
            req := httptest.NewRequest(http.MethodPost, "/shorten",
                bytes.NewBufferString(tt.body))
            req.Header.Set("Content-Type", "application/json")
            w := httptest.NewRecorder()
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code, "status mismatch")

            if tt.wantData {
                var resp map[string]interface{}
                err := json.Unmarshal(w.Body.Bytes(), &resp)
                require.NoError(t, err)

                assert.Equal(t, true, resp["success"])

                data, ok := resp["data"].(map[string]interface{})
                require.True(t, ok, "data should be an object")
                assert.Contains(t, data["short_url"], "http://localhost:8080/")
                assert.NotEmpty(t, data["short_code"])
                assert.Equal(t, "https://example.com", data["original_url"])
            }
        })
    }
}

// ── Table-driven: GET /:code ──────────────────────────────────────────────────

func TestRedirect(t *testing.T) {
    tests := []struct {
        name           string
        code           string
        mockFn         func(ctx context.Context, code string) (*model.URL, *apperror.AppError)
        wantStatus     int
        wantLocation   string
    }{
        {
            name: "success",
            code: "abc123",
            mockFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
                return &model.URL{
                    ID:          "some-uuid",
                    ShortCode:   code,
                    OriginalURL: "https://example.com",
                }, nil
            },
            wantStatus:   http.StatusMovedPermanently,
            wantLocation: "https://example.com",
        },
        {
            name: "not found",
            code: "notexist",
            mockFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
                return nil, apperror.NotFound("short url not found")
            },
            wantStatus: http.StatusNotFound,
        },
        {
            name: "expired link",
            code: "expired",
            mockFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
                return nil, apperror.Gone("short url has expired")
            },
            wantStatus: http.StatusGone,
        },
        {
            name: "server error",
            code: "errcode",
            mockFn: func(ctx context.Context, code string) (*model.URL, *apperror.AppError) {
                return nil, apperror.Internal("something went wrong")
            },
            wantStatus: http.StatusInternalServerError,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            mock := &mockURLService{
                getByShortCodeFn: tt.mockFn,
            }

            r := setupRouter(mock)
            req := httptest.NewRequest(http.MethodGet,
                "/"+tt.code, nil)
            w := httptest.NewRecorder()
            r.ServeHTTP(w, req)

            assert.Equal(t, tt.wantStatus, w.Code)

            if tt.wantLocation != "" {
                assert.Equal(t, tt.wantLocation,
                    w.Header().Get("Location"))
            }
        })
    }
}