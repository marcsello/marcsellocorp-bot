package db

import (
	"database/sql"
	"gitlab.com/MikeTTh/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var db *gorm.DB

func Connect() (err error) {
	dsn := env.StringOrPanic("DATABASE_URL")
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true, // Epic performance improvement
	})
	if err != nil {
		return
	}

	var sqlDB *sql.DB
	sqlDB, err = db.DB()
	if err != nil {
		return
	}

	// hopefully this sets stuff globally
	sqlDB.SetConnMaxLifetime(time.Minute * 15)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(10)

	err = db.AutoMigrate(&Channel{}, &User{}, &PendingQuestion{}, &Token{})
	if err != nil {
		return
	}

	return
}
