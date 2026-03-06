package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Number struct {
	ID     string `json:"id" gorm:"type:uuid;primaryKey"`
	Number int    `json:"number" gorm:"not null;uniqueIndex"`
}

func (Number) TableName() string {
	return "numbers"
}

func (n *Number) BeforeCreate(*gorm.DB) error {
	if n.ID == "" {
		n.ID = uuid.NewString()
	}

	return nil
}
