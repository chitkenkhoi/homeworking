package app

import (
	"log"
	"fmt"

	cfg "lqkhoi-go-http-api/config"
	"lqkhoi-go-http-api/internal/infrastructure"
	"lqkhoi-go-http-api/internal/migration"
	"lqkhoi-go-http-api/internal/routes"
	// "lqkhoi-go-http-api/internal/repository"

	"github.com/gofiber/fiber/v2"
)
type App struct{
	App *fiber.App
	Config *cfg.Config
}
func New() *App{
	cfg := cfg.NewConfig()

	app := fiber.New()

	return &App{
		App: app,
		Config: cfg,
	}
	// user_repository := repository.NewUserRepository(db)
	// redis_client := infrastructure.NewRedisConnection(cfg.RedisConfig)
	// app := fiber.New()
	// return app
}

func (app *App) Setup()*App{
	app.Config.LoadAppConfig().LoadDBConfig().LoadRedisConfig()

	db,err := infrastructure.NewDBConnection(app.Config.DBConfig)
	if err!=nil{
		log.Fatal(err)
	}

	if err := migration.AutoMigrate(db);err !=nil{
		log.Fatal(err)
	}

	routes.SetupUserRoutes(app.App)
	
	return app
}

func (app *App)Run(){
	log.Printf("Starting server on port %s...",app.Config.AppConfig.Port)
	log.Println(app.App.Listen(fmt.Sprintf(":%s",app.Config.AppConfig.Port)))
}