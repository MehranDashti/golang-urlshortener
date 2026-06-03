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
    var user model.User
    result := r.db.WithContext(ctx).
        Where("email = ?", email).First(&user)
    if result.Error == gorm.ErrRecordNotFound {
        return nil, nil
    }
    if result.Error != nil {
        return nil, fmt.Errorf(
            "UserRepository.FindByEmail %s: %w", email, result.Error)
    }
    return &user, nil
}

func (r *UserRepository) FindAll(
    ctx context.Context) ([]*model.User, error) {
    var users []*model.User
    if err := r.db.WithContext(ctx).
        Find(&users).Error; err != nil {
        return nil, fmt.Errorf(
            "UserRepository.FindAll: %w", err)
    }
    return users, nil
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