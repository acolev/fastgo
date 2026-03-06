package database

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

var db *gorm.DB

func SetDB(d *gorm.DB) {
	db = d
}

func DB() *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db
}

func SQLDB() (*sql.DB, error) {
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	return db.DB()
}

func Ping(ctx context.Context) error {
	sqlDB, err := SQLDB()
	if err != nil {
		return err
	}

	return sqlDB.PingContext(ctx)
}

func Close() error {
	if db == nil {
		return nil
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	db = nil
	return sqlDB.Close()
}
