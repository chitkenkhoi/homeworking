package app

import (
	"fmt"
	"log"

	"lqkhoi-go-http-api/config"
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/infrastructure"
	"lqkhoi-go-http-api/internal/migration"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/internal/routes"
	"lqkhoi-go-http-api/internal/service"

	// "lqkhoi-go-http-api/internal/repository"

	"github.com/gofiber/fiber/v2"
)

type App struct {
	server *fiber.App
	config *config.Config
}

func New() *App {

	cfg := &config.Config{}
	app := fiber.New()

	return &App{
		server: app,
		config: cfg,
	}
}

func (app *App) Setup() error {
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	app.config = &cfg

	db, err := infrastructure.NewDBConnection(app.config.Database)
	if err != nil {
		log.Fatal(err)
		return err
	}

	if err := migration.AutoMigrate(db); err != nil {
		log.Fatal(err)
		return err
	}
	userRepository := repository.NewUserRepository(db)
	projectRepository := repository.NewProjectRepository(db)

	userService := service.NewUserService(userRepository)
	projectService := service.NewProjectService(projectRepository,userRepository)

	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)

	routes.SetupUserRoutes(app.server, userHandler)
	routes.SetupProjectRoutes(app.server,projectHandler)

	return nil
}

func (app *App) Run() {
	log.Printf("Starting server on port %s...", app.config.Server.Port)
	log.Println(app.server.Listen(fmt.Sprintf(":%s", app.config.Server.Port)))
}
