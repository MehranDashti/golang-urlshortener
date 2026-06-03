package repository

import (
    "context"
    "fmt"

    "gorm.io/gorm"
    "urlshortener/internal/model"
)

type URLRepository struct {
    db *gorm.DB
}

func NewURLRepository(db *gorm.DB) *URLRepository {
    return &URLRepository{db: db}
}

func (r *URLRepository) Create(
    ctx context.Context, url *model.URL) error {
    if err := r.db.WithContext(ctx).Create(url).Error; err != nil {
        return fmt.Errorf("URLRepository.Create: %w", err)
    }
    return nil
}

func (r *URLRepository) FindByShortCode(
    ctx context.Context,
    code string) (*model.URL, error) {
    return findOne[model.URL](ctx, r.db,
        "short_code = ?", code)
}

func (r *URLRepository) IncrementClicks(
    ctx context.Context, id string) error {
    err := r.db.WithContext(ctx).
        Model(&model.URL{}).
        Where("id = ?", id).
        UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
    if err != nil {
        return fmt.Errorf(
            "URLRepository.IncrementClicks %s: %w", id, err)
    }
    return nil
}

func (r *URLRepository) FindByUserID(
    ctx context.Context,
    userID string) ([]*model.URL, error) {
    var urls []*model.URL
    if err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Find(&urls).Error; err != nil {
        return nil, fmt.Errorf(
            "URLRepository.FindByUserID %s: %w", userID, err)
    }
    return urls, nil
}

func (r *URLRepository) FindByUserIDPaginated(
    ctx context.Context,
    userID string,
    params model.PaginationParams) ([]*model.URL, int64, error) {
    var urls  []*model.URL
    var total int64

    if err := r.db.WithContext(ctx).
        Model(&model.URL{}).
        Where("user_id = ?", userID).
        Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf(
            "URLRepository.FindByUserIDPaginated count %s: %w",
            userID, err)
    }

    if err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Offset(params.Offset()).
        Limit(params.Limit).
        Order("created_at DESC").
        Find(&urls).Error; err != nil {
        return nil, 0, fmt.Errorf(
            "URLRepository.FindByUserIDPaginated fetch %s: %w",
            userID, err)
    }
    return urls, total, nil
}

func (r *URLRepository) FindAll(
    ctx context.Context) ([]*model.URL, error) {
    return findAll[model.URL](ctx, r.db, "")
}

func (r *URLRepository) Delete(
    ctx context.Context, id string) error {
    if err := r.db.WithContext(ctx).
        Delete(&model.URL{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf(
            "URLRepository.Delete %s: %w", id, err)
    }
    return nil
}

func (r *URLRepository) DeleteByUserID(
    ctx context.Context, userID string) error {
    if err := r.db.WithContext(ctx).
        Where("user_id = ?", userID).
        Delete(&model.URL{}).Error; err != nil {
        return fmt.Errorf(
            "URLRepository.DeleteByUserID %s: %w", userID, err)
    }
    return nil
}