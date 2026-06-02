package repository

import (
    "context"

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
    return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepository) FindByEmail(
    ctx context.Context, email string) (*model.User, error) {
    var user model.User
    result := r.db.WithContext(ctx).
        Where("email = ?", email).First(&user)
    if result.Error == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &user, result.Error
}

func (r *UserRepository) FindAll(
    ctx context.Context) ([]*model.User, error) {
    var users []*model.User
    return users, r.db.WithContext(ctx).Find(&users).Error
}

func (r *UserRepository) Delete(
    ctx context.Context, id string) error {
    return r.db.WithContext(ctx).
        Delete(&model.User{}, "id = ?", id).Error
}