package model

import (
	"time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

type User struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"`
    Email     string    `gorm:"type:varchar(255);uniqueIndex;not null"`
    Password  string    `gorm:"type:varchar(255);not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
    u.ID = uuid.New().String()
    return nil
}