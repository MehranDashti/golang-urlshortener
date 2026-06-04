package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type URL struct {
	ID          string `gorm:"type:varchar(36);primaryKey"`
	UserID      string `gorm:"type:varchar(36);not null;index"`
	OriginalURL string `gorm:"type:text;not null"`
	ShortCode   string `gorm:"type:varchar(10);uniqueIndex;not null"`
	Clicks      int    `gorm:"default:0"`
	ExpiresAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *URL) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New().String()
	return nil
}
