package routes

import (
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/middlewares"
	"lqkhoi-go-http-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func SetupTaskRoutes(app *fiber.App, h *handler.TaskHandler, lm fiber.Handler) {
	log := app.Group("/")
	log.Use(lm)
	
	authenticated := log.Group("/")
	authenticated.Use(middlewares.AuthMiddleware)
	authenticated.Get("/tasks/:taskId", h.GetTask)

	OwnerOrProjectManager := authenticated.Group("/")
	OwnerOrProjectManager.Get("/users/:userId/tasks", h.FindTasksByUserID)

	ProjectManagerOnly := authenticated.Group("/")
	ProjectManagerOnly.Use(middlewares.RequireRoleIs(models.ProjectManager))

	ProjectManagerOnly.Get("/projects/:projectId/tasks", h.FindTasksByProjectID)
	ProjectManagerOnly.Post("/tasks",h.CreateTask)
	ProjectManagerOnly.Get("/tasks", h.FindTasks)
	ProjectManagerOnly.Put("/tasks/:taskId", h.UpdateTask)
	ProjectManagerOnly.Delete("/tasks/:taskId", h.DeleteTask)
	ProjectManagerOnly.Post("/tasks/:taskId/assign", h.AssignTaskToUser)
}
