package initializers

import models "github.com/DimaGlobin/matchme/models"

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Photo{})
	DB.AutoMigrate(&models.Chat{})
	DB.AutoMigrate(&models.Message{})
}
