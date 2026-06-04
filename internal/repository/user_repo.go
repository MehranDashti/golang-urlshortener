package repository

import (
    "context"
    "fmt"

    "gorm.io/gorm"
    "urlshortener/internal/model"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
    return &UserRepository{db: tx}
}

func (r *UserRepository) DB() *gorm.DB {
    return r.db
}

func (r *UserRepository) Create(
    ctx context.Context, user *model.User) error {
    if err := r.db.WithContext(ctx).
        Create(user).Error; err != nil {
        return fmt.Errorf(
            "UserRepository.Create %s: %w", user.Email, err)
    }
    return nil
}

func (r *UserRepository) FindByEmail(
    ctx context.Context,
    email string) (*model.User, error) {
    return findOne[model.User](ctx, r.db,
        "email = ?", email)
}

func (r *UserRepository) FindAll(
    ctx context.Context) ([]*model.User, error) {
    return findAll[model.User](ctx, r.db, "")
}

func (r *UserRepository) Delete(
    ctx context.Context, id string) error {
    if err := r.db.WithContext(ctx).
        Delete(&model.User{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf(
            "UserRepository.Delete %s: %w", id, err)
    }
    return nil
}