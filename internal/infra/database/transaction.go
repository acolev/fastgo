package database

import "gorm.io/gorm"

func Transaction(fn func(tx *gorm.DB) error) error {

	return DB().Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})

}
