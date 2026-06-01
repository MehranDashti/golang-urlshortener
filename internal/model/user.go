package model

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type Role string

const (
    RoleAdmin Role = "admin"
    RoleUser  Role = "user"
)

type User struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"`
    Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    Password  string    `gorm:"type:varchar(255);not null"`
    Role      Role      `gorm:"type:varchar(20);default:'user'"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.ID = uuid.New().String()
    if u.Role == "" {
        u.Role = RoleUser
    }
    return nil
}