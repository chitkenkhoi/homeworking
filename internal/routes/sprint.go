package routes

import (
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/middlewares"
	"lqkhoi-go-http-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func SetupSprintRoutes(app *fiber.App, h *handler.SprintHandler) {
	authenticated := app.Group("/")
	authenticated.Use(middlewares.AuthMiddleware)

	projectManagerSprint := authenticated.Group("/sprints")
	projectManagerSprint.Use(middlewares.RequireRoleIs(models.ProjectManager))

	projectManagerSprint.Post("/", h.CreateSprint)
	projectManagerSprint.Get("/", h.ListSprints)
	projectManagerSprint.Get("/:sprintId", h.GetSprint)
	projectManagerSprint.Put("/:sprintId")
	projectManagerSprint.Delete("/:sprintId", h.DeleteSprint)
}
