package app

import (
	"fmt"

	"lqkhoi-go-http-api/config"
	"lqkhoi-go-http-api/internal/infrastructure"
	"lqkhoi-go-http-api/internal/repository"

	"github.com/gofiber/fiber/v2"
)

func New() *fiber.App{
	app := fiber.New()
	cfg := config.NewConfig().
			LoadDBConfig().
			LoadRedisConfig()


	db,err := infrastructure.NewDBConnection(cfg.DBConfig)
	if err!=nil{
		fmt.Println(err)
	}
	user_repository := repository.NewUserRepository(db)
	redis_client := infrastructure.NewRedisConnection(cfg.RedisConfig)

	return app
}