package repository

import (
    "gorm.io/gorm"
    "urlshortener/internal/model"
)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
    return r.db.Create(user).Error
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
    var user model.User
    result := r.db.Where("email = ?", email).First(&user)
    if result.Error == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &user, result.Error
}