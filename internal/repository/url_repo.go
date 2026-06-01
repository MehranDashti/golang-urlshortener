package repository

import (
    "errors"

    "gorm.io/gorm"

    "urlshortener/internal/model"
)

type URLRepository struct {
    db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
    return &URLRepository{db: db}
}

func (r *URLRepository) Create(url *model.URL) error {
    result := r.db.Create(url)
    return result.Error
}

func (r *URLRepository) FindByShortCode(code string) (*model.URL, error) {
    var url model.URL
    result := r.db.Where("short_code = ?", code).First(&url)
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &url, result.Error
}

func (r *URLRepository) IncrementClicks(id string) error {
    result := r.db.Model(&model.URL{}).
        Where("id = ?", id).
        UpdateColumn("clicks", gorm.Expr("clicks + 1"))
    return result.Error
}