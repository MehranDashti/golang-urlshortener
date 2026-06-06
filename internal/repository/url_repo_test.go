//go:build integration

package repository_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"urlshortener/tests/testcontainer"
	"gorm.io/gorm"

	"urlshortener/internal/model"
	"urlshortener/internal/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    ctx := context.Background()

    db, cleanup, err := testcontainer.NewMySQL(ctx)
    require.NoError(t, err, "failed to start MySQL container")

    t.Cleanup(cleanup) // container terminates when test ends
    return db
}

func TestCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewURLRepository(db)

	url := &model.URL{
		OriginalURL: "https://google.com",
		ShortCode:   "abc123",
	}

	repo.Create(context.Background(), url)
	repo.FindByShortCode(context.Background(), "abc123")
	repo.IncrementClicks(context.Background(), url.ID)
}

func TestFindByShortCode(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewURLRepository(db)

	url := &model.URL{
		OriginalURL: "https://google.com",
		ShortCode:   "abc123",
	}
	repo.Create(context.Background(), url) // ← add context

	t.Run("found", func(t *testing.T) {
		found, err := repo.FindByShortCode(context.Background(), "abc123") // ← add context
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "https://google.com", found.OriginalURL)
	})

	t.Run("not found", func(t *testing.T) {
		found, err := repo.FindByShortCode(context.Background(), "xxxxxx") // ← add context
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, found)
	})
}

func TestIncrementClicks(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewURLRepository(db)

	url := &model.URL{
		OriginalURL: "https://google.com",
		ShortCode:   "abc123",
	}
	repo.Create(context.Background(), url) // ← add context

	err := repo.IncrementClicks(context.Background(), url.ID) // ← add context
	assert.NoError(t, err)

	found, _ := repo.FindByShortCode(context.Background(), "abc123") // ← add context
	assert.Equal(t, 1, found.Clicks)
}

func TestCreate_DuplicateShortCode(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewURLRepository(db)

	url1 := &model.URL{
		OriginalURL: "https://google.com",
		ShortCode:   "abc123",
		UserID:      "user-1",
	}
	url2 := &model.URL{
		OriginalURL: "https://github.com",
		ShortCode:   "abc123",
		UserID:      "user-1",
	}

	err1 := repo.Create(context.Background(), url1)
	assert.NoError(t, err1)

	err2 := repo.Create(context.Background(), url2)
	assert.Error(t, err2) 

	assert.Contains(t, err2.Error(), "URLRepository.Create")
	
	t.Logf("wrapped error: %v", err2)
}