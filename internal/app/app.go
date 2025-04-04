package app

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"lqkhoi-go-http-api/internal/config"
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/infrastructure"
	"lqkhoi-go-http-api/internal/middlewares"
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
	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // Logs: INFO, WARN, ERROR (but not DEBUG)
	}

	// Create logger with the level restriction
	logger := slog.New(slog.NewTextHandler(os.Stdout, opts))

	cfg, err := config.LoadConfig("./internal/config")
	if err != nil {
		logger.Error("Failed to load configuration","erorr",err.Error())
	}
	app.config = &cfg

	db, err := infrastructure.NewDBConnection(app.config.Database)
	if err != nil {
		logger.Error("Failed to connect to database","error",err)
		return err
	}

	if err := migration.AutoMigrate(db); err != nil {
		logger.Error("Failed to migrate database","error",err)
		return err
	}
	userRepository := repository.NewUserRepository(db)
	projectRepository := repository.NewProjectRepository(db, cfg.DateTime)
	sprintRepository := repository.NewSprintRepository(db, cfg.DateTime)
	taskRepository := repository.NewTaskRepository(db, cfg.DateTime)

	userService := service.NewUserService(userRepository)
	projectService := service.NewProjectService(projectRepository, userService)
	sprintService := service.NewSprintService(sprintRepository, projectService, cfg.DateTime)
	taskService := service.NewTaskService(taskRepository, projectService, sprintService, userService)

	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService, cfg.DateTime)
	sprintHandler := handler.NewSprintHandler(sprintService, cfg.DateTime)
	taskHandler := handler.NewTaskHandler(taskService, cfg.DateTime)


	
	lm := middlewares.NewLoggingMiddleware(logger)
	routes.SetupUserRoutes(app.server, userHandler, lm)
	routes.SetupProjectRoutes(app.server, projectHandler, lm)
	routes.SetupSprintRoutes(app.server, sprintHandler, lm)
	routes.SetupTaskRoutes(app.server, taskHandler, lm)

	return nil
}

func (app *App) Run() {
	log.Printf("Starting server on port %s...", app.config.Server.Port)
	log.Println(app.server.Listen(fmt.Sprintf(":%s", app.config.Server.Port)))
}
