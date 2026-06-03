package service_test

import (
    "bytes"
    "context"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "urlshortener/internal/model"
    "urlshortener/internal/service"
)

// mockURLRepoCSV for CSV tests
type mockURLRepoCSV struct {
    urls   []*model.URL
    created []*model.URL
}

func (m *mockURLRepoCSV) FindAll(
    ctx context.Context) ([]*model.URL, error) {
    return m.urls, nil
}

func (m *mockURLRepoCSV) Create(
    ctx context.Context, url *model.URL) error {
    m.created = append(m.created, url)
    return nil
}

// Satisfy full interface with no-ops
func (m *mockURLRepoCSV) FindByShortCode(ctx context.Context, code string) (*model.URL, error) { return nil, nil }
func (m *mockURLRepoCSV) IncrementClicks(ctx context.Context, id string) error                 { return nil }
func (m *mockURLRepoCSV) FindByUserID(ctx context.Context, userID string) ([]*model.URL, error) { return nil, nil }
func (m *mockURLRepoCSV) FindByUserIDPaginated(ctx context.Context, userID string, params model.PaginationParams) ([]*model.URL, int64, error) { return nil, 0, nil }
func (m *mockURLRepoCSV) FindAll2(ctx context.Context) ([]*model.URL, error)                    { return nil, nil }
func (m *mockURLRepoCSV) Delete(ctx context.Context, id string) error                           { return nil }
func (m *mockURLRepoCSV) DeleteByUserID(ctx context.Context, userID string) error               { return nil }

func TestWriteLinksCSV(t *testing.T) {
    repo := &mockURLRepoCSV{
        urls: []*model.URL{
            {ID: "1", ShortCode: "abc123",
                OriginalURL: "https://google.com", Clicks: 5},
            {ID: "2", ShortCode: "def456",
                OriginalURL: "https://github.com", Clicks: 10},
        },
    }

    adminRepo := &mockAdminURLRepo{
        findAllFn: func(ctx context.Context) ([]*model.URL, error) {
            return repo.urls, nil
        },
    }
    userRepo := &mockAdminUserRepo{}
    svc := service.NewAdminService(adminRepo, userRepo)

    // Write CSV to a buffer — bytes.Buffer implements io.Writer
    var buf bytes.Buffer
    err := svc.WriteLinksCSV(context.Background(), &buf)
    require.NoError(t, err)

    csv := buf.String()
    t.Logf("CSV output:\n%s", csv)

    // Check header
    assert.Contains(t, csv, "id,short_code,original_url")

    // Check data rows
    assert.Contains(t, csv, "abc123")
    assert.Contains(t, csv, "https://google.com")
    assert.Contains(t, csv, "def456")
    assert.Contains(t, csv, "https://github.com")
}

func TestImportLinksCSV(t *testing.T) {
    repo := &mockURLRepoCSV{}
    svc := service.NewURLService(repo, context.Background())

    // CSV content as a string reader — strings.NewReader implements io.Reader
    csvContent := `original_url
https://google.com
https://github.com
https://golang.org
`
    results, err := svc.ImportLinksCSV(
        context.Background(),
        strings.NewReader(csvContent), // ← strings.NewReader = io.Reader
        "user-123",
    )

    require.NoError(t, err)
    assert.Len(t, results, 3)
    assert.Len(t, repo.created, 3)

    for _, r := range results {
        assert.Empty(t, r.Error,
            "row %d should not have error", r.Row)
        assert.NotEmpty(t, r.ShortCode)
    }
}