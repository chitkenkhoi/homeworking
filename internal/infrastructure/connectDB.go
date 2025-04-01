package infrastructure

import (
	"fmt"

	"lqkhoi-go-http-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewDBConnection(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",cfg.Host,cfg.User,cfg.Password,cfg.Name,cfg.Port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	return db, err
}
