package migration

import (
	"lqkhoi-go-http-api/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB){
	db.AutoMigrate(&models.User{})
}