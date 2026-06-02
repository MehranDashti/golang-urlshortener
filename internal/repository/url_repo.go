package repository

import (
    "context"

    "gorm.io/gorm"
    "urlshortener/internal/model"
)

type URLRepository struct {
    db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
    return &URLRepository{db: db}
}

// WithContext passes the request context to GORM
// — DB queries cancel if the request is cancelled
func (r *URLRepository) Create(
    ctx context.Context, url *model.URL) error {
    return r.db.WithContext(ctx).Create(url).Error
}

func (r *URLRepository) FindByShortCode(
    ctx context.Context, code string) (*model.URL, error) {
    var url model.URL
    result := r.db.WithContext(ctx).
        Where("short_code = ?", code).First(&url)
    if result.Error == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &url, result.Error
}

func (r *URLRepository) IncrementClicks(
    ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Model(&model.URL{}).
        Where("id = ?", id).
        UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}

func (r *URLRepository) FindByUserID(
    ctx context.Context, userID string) ([]*model.URL, error) {
    var urls []*model.URL
    result := r.db.WithContext(ctx).
        Where("user_id = ?", userID).Find(&urls)
    return urls, result.Error
}

func (r *URLRepository) FindAll(
    ctx context.Context) ([]*model.URL, error) {
    var urls []*model.URL
    return urls, r.db.WithContext(ctx).Find(&urls).Error
}

func (r *URLRepository) Delete(
    ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Delete(&model.URL{}, "id = ?", id).Error
}