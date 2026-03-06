package database

import "gorm.io/gorm"

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
