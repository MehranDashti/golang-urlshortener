package testcontainer

import (
	"context"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"urlshortener/internal/model"
)

func NewMySQL(ctx context.Context) (*gorm.DB, func(), error) {
	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("testdb"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("secret"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start mysql container: %w", err)
	}

	cleanup := func() {
		_ = container.Terminate(context.Background())
	}

	dsn, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("get connection string: %w", err)
	}

	var db *gorm.DB
	// Retry — container may be ready before MySQL is accepting connections
	for i := range 10 {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err == nil {
			break
		}
		time.Sleep(time.Duration(i+1) * 200 * time.Millisecond)
	}
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("connect to mysql container: %w", err)
	}

	// Run migrations so the schema matches production
	if err := db.AutoMigrate(&model.User{}, &model.URL{}); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("migrate test schema: %w", err)
	}

	return db, cleanup, nil
}
