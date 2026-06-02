package database

import (
    "log/slog"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    _ "github.com/golang-migrate/migrate/v4/source/file" 
    _ "github.com/go-sql-driver/mysql"                  
    gormmysql "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
    gormDB, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{})
    if err != nil {
        panic("failed to connect to database: " + err.Error())
    }

    if err := runMigrations(gormDB, dsn); err != nil {
        panic("failed to run migrations: " + err.Error())
    }

    return gormDB
}

func runMigrations(db *gorm.DB, dsn string) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }

    driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
    if err != nil {
        return err
    }

    m, err := migrate.NewWithDatabaseInstance(
        "file://migrations", // path to migration files
        "mysql",
        driver,
    )
    if err != nil {
        return err
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }

    version, _, _ := m.Version()
    slog.Info("migrations applied", "version", version)
    return nil
}