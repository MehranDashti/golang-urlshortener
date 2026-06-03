package database

import (
    "log/slog"
    "sync"

    "github.com/golang-migrate/migrate/v4"
    "github.com/golang-migrate/migrate/v4/database/mysql"
    "github.com/golang-migrate/migrate/v4/source/iofs"
    _ "github.com/go-sql-driver/mysql"
    gormmysql "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "urlshortener/migrations" 
)

var (
    db   *gorm.DB
    once sync.Once
)

func Connect(dsn string) *gorm.DB {
    once.Do(func() {
        gormDB, err := gorm.Open(gormmysql.Open(dsn), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Silent),
        })
        if err != nil {
            panic("failed to connect to database: " + err.Error())
        }

        if err := runMigrations(gormDB); err != nil {
            panic("failed to run migrations: " + err.Error())
        }

        db = gormDB
    })
    return db
}

func runMigrations(db *gorm.DB) error {
    sqlDB, err := db.DB()
    if err != nil {
        return err
    }

    driver, err := mysql.WithInstance(sqlDB, &mysql.Config{})
    if err != nil {
        return err
    }

    // Use embedded FS — no disk access needed
    sourceDriver, err := iofs.New(migrations.FS, ".")
    if err != nil {
        return err
    }

    m, err := migrate.NewWithInstance(
        "iofs",
        sourceDriver,
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