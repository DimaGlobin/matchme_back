package services

import (
	// "strconv"

	"github.com/DimaGlobin/matchme/initializers"
	"github.com/DimaGlobin/matchme/models"
)

func CreateChat(user1ID, user2ID uint) (uint, error) {
	chat := models.Chat{
		User1ID: user1ID,
		User2ID: user2ID,
	}

	result := initializers.DB.Create(&chat)
	if err := result.Error; err != nil {
		return 0, err
	}

	// chatChannel := "chat:" + strconv.FormatUint(uint64(chat.ID), 10)


	return chat.ID, nil
}

func SendMessageToChat(chatId uint, senderID uint, text string) error {
	message := models.Message{
		ChatID:   chatId,
		SenderID: senderID,
		Body:     text,
	}

	result := initializers.DB.Create(&message)
	if err := result.Error; err != nil {
		return err
	}

	return nil
}
