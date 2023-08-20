package models

import "gorm.io/gorm"

type Photo struct {
	gorm.Model
	UserID    uint   `gorm:"not null"`
	ImageName string `gorm:"unique"`
	Hash      uint32 `gorm:"unique"`
	Caption   string
}
