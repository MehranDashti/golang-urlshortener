package repository_test

import (
    "testing"
    "context"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"

    "urlshortener/internal/model"
    "urlshortener/internal/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    err = db.AutoMigrate(&model.URL{})
    require.NoError(t, err)
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
        assert.NoError(t, err)
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