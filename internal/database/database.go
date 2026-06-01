package database

import (
	"log"

	"gorm.io/driver/mysql"
    "gorm.io/gorm"

    "urlshortener/internal/model"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect to database: ", err)
    }

    err = db.AutoMigrate(&model.User{}, &model.URL{})
    if err != nil {
        log.Fatal("failed to migrate: ", err)
    }

    return db
}