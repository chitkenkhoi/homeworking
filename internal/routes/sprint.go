package routes

import (
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/middlewares"
	"lqkhoi-go-http-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func SetupSprintRoutes(prefixApp fiber.Router, h *handler.SprintHandler, lm fiber.Handler) {
	log := prefixApp.Group("/")
	log.Use(lm)

	authenticated := log.Group("/")
	authenticated.Use(middlewares.AuthMiddleware)

	projectManagerSprint := authenticated.Group("/sprints")
	projectManagerSprint.Use(middlewares.RequireRoleIs(models.ProjectManager))

	projectManagerSprint.Post("/", h.CreateSprint)
	projectManagerSprint.Get("/", h.FindSprints)
	projectManagerSprint.Get("/:sprintId", h.GetSprint)
	projectManagerSprint.Put("/:sprintId", h.UpdateSprint)
	projectManagerSprint.Delete("/:sprintId", h.DeleteSprint)
}
