package models

import (
	"github.com/lib/pq"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email       string `gorm:"unique"`
	Password    string
	Sex         string
	Location    string
	Description string        `gorm:"type:TEXT"`
	Liked       pq.Int64Array `gorm:"type:integer[]"`
	Disliked    pq.Int64Array `gorm:"type:integer[]"`
	Matches     pq.Int64Array `gorm:"type:integer[]"`
}
