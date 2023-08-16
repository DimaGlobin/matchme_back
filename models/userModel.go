package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email            string `gorm:"unique"`
	Password         string
	Sex              string
	Location         string
	Description      string `gorm:"type:TEXT"`
	liked            []int  `gorm:"type:integer[]"`
	disliked         []int  `gorm:"type:integer[]"`
	opened_chat_with []int  `gorm:"type:integer[]"`
}
