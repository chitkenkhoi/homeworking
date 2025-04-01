package routes

import (
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/middlewares"
	"lqkhoi-go-http-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func SetupProjectRoutes(app *fiber.App, h *handler.ProjectHandler) {
	authenticated := app.Group("/")
	authenticated.Use(middlewares.AuthMiddleware)

	projectManagerOnly := authenticated.Group("/projects")
	projectManagerOnly.Use(middlewares.RequireRoleIs(models.ProjectManager))

	projectManagerOnly.Post("/", h.CreateProjectHandler)
	projectManagerOnly.Post("/:projectId", h.AddTeamMembers)
	projectManagerOnly.Get("/", h.ListProjectsHanlder)
	projectManagerOnly.Get("/:projectId", h.GetProject)
	projectManagerOnly.Put("/:projectId", h.UpdateProject)
	projectManagerOnly.Delete("/:projectId", h.DeleteProject)

}
