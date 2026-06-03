package repository

import (
    "context"
    "fmt"

    "gorm.io/gorm"
)

// findOne fetches a single record by condition.
// T can be any GORM model — URL, User, etc.
// Usage: findOne[model.URL](ctx, db, "short_code = ?", code)
func findOne[T any](
    ctx context.Context,
    db *gorm.DB,
    condition string,
    args ...interface{}) (*T, error) {

    var result T
    err := db.WithContext(ctx).
        Where(condition, args...).
        First(&result).Error

    if err == gorm.ErrRecordNotFound {
        return nil, nil // not found — not an error
    }
    if err != nil {
        return nil, fmt.Errorf("findOne[%T] %s: %w",
            result, condition, err)
    }
    return &result, nil
}

// findAll fetches all records matching condition.
// Usage: findAll[model.URL](ctx, db, "user_id = ?", userID)
func findAll[T any](
    ctx context.Context,
    db *gorm.DB,
    condition string,
    args ...interface{}) ([]*T, error) {

    var results []*T
    query := db.WithContext(ctx)
    if condition != "" {
        query = query.Where(condition, args...)
    }
    if err := query.Find(&results).Error; err != nil {
        return nil, fmt.Errorf("findAll[%T] %s: %w",
            *new(T), condition, err)
    }
    return results, nil
}