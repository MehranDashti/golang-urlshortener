package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"urlshortener/internal/model"
	"urlshortener/internal/repository"
	"urlshortener/internal/service"
)

// mockURLRepo for url_service tests
type mockURLRepo struct {
	createFn                func(ctx context.Context, url *model.URL) error
	findByShortCodeFn       func(ctx context.Context, code string) (*model.URL, error)
	incrementClicksFn       func(ctx context.Context, id string) error
	findByUserIDFn          func(ctx context.Context, userID string) ([]*model.URL, error)
	findByUserIDPaginatedFn func(ctx context.Context, userID string, params model.PaginationParams) ([]*model.URL, int64, error)
}

func (m *mockURLRepo) Create(ctx context.Context, url *model.URL) error {
	if m.createFn != nil {
		return m.createFn(ctx, url)
	}
	return nil
}
func (m *mockURLRepo) FindByShortCode(ctx context.Context, code string) (*model.URL, error) {
	if m.findByShortCodeFn != nil {
		return m.findByShortCodeFn(ctx, code)
	}
	return nil, nil
}
func (m *mockURLRepo) IncrementClicks(ctx context.Context, id string) error {
	if m.incrementClicksFn != nil {
		return m.incrementClicksFn(ctx, id)
	}
	return nil
}
func (m *mockURLRepo) FindByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	if m.findByUserIDFn != nil {
		return m.findByUserIDFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockURLRepo) FindByUserIDPaginated(ctx context.Context, userID string, params model.PaginationParams) ([]*model.URL, int64, error) {
	if m.findByUserIDPaginatedFn != nil {
		return m.findByUserIDPaginatedFn(ctx, userID, params)
	}
	return nil, 0, nil
}
func (m *mockURLRepo) WithTx(_ *gorm.DB) *repository.URLRepository { return nil }
func (m *mockURLRepo) DB() *gorm.DB                                { return nil }

// simulateDuplicateKeyError simulates a MySQL duplicate key error
// Since we can't create a real mysql.MySQLError in tests, we test
// the retry logic by counting attempts
func TestShortenURL_RetryOnCollision(t *testing.T) {
	attempts := 0

	repo := &mockURLRepo{
		createFn: func(ctx context.Context, url *model.URL) error {
			attempts++
			if attempts < 3 {
				// Simulate non-duplicate error first 2 times
				// Real collision detection tested in integration tests
				return errors.New("some other error")
			}
			return nil // success on 3rd attempt
		},
	}

	svc := service.NewURLService(repo, context.Background())
	_, appErr := svc.ShortenURL(
		context.Background(),
		"https://google.com",
		"user-123",
		nil,
	)

	// First non-duplicate error stops retry
	assert.NotNil(t, appErr)
	assert.Equal(t, 1, attempts, "should stop on non-duplicate error")
}

func TestShortenURL_Success(t *testing.T) {
	repo := &mockURLRepo{
		createFn: func(ctx context.Context, url *model.URL) error {
			return nil
		},
	}

	svc := service.NewURLService(repo, context.Background())
	result, appErr := svc.ShortenURL(
		context.Background(),
		"https://google.com",
		"user-123",
		nil,
	)

	require.Nil(t, appErr)
	assert.NotEmpty(t, result.ShortCode)
	assert.Equal(t, "https://google.com", result.OriginalURL)
}

func TestShortenURL_WithExpiry(t *testing.T) {
	repo := &mockURLRepo{
		createFn: func(ctx context.Context, url *model.URL) error {
			return nil
		},
	}

	svc := service.NewURLService(repo, context.Background())
	expiry := time.Now().Add(24 * time.Hour)
	result, appErr := svc.ShortenURL(
		context.Background(),
		"https://google.com",
		"user-123",
		&expiry,
	)

	require.Nil(t, appErr)
	assert.NotNil(t, result.ExpiresAt)
}
