package initializers

import (
	"github.com/DimaGlobin/matchme/models"
)

func SyncDatabase() {
	DB.AutoMigrate(&models.User{})
}