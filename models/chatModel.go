package models

import (
	"time"

	"gorm.io/gorm"
)

type Chat struct {
	gorm.Model
	User1ID     uint
	User2ID     uint
	LastMessage string
	Messages    []Message
}

type Message struct {
	gorm.Model
	ChatID    uint
	SenderID  uint
	Body      string
	CreatedAt time.Time
}
