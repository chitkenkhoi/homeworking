package infrastructure

import (
	"fmt"

	"lqkhoi-go-http-api/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDBConnection(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",cfg.DBHost,cfg.DBUser,cfg.DBPassword,cfg.DBName,cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	return db, err
}
