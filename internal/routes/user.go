package routes

import (
	"lqkhoi-go-http-api/internal/handler"
	"lqkhoi-go-http-api/internal/middlewares"
	"lqkhoi-go-http-api/internal/models"

	"github.com/gofiber/fiber/v2"
)

func sampleHanlder(c *fiber.Ctx) error {
	return nil
}

func SetupUserRoutes(app *fiber.App, h *handler.UserHandler) {
	app.Post("/users", h.CreateUserHandler)
	app.Post("/login", h.Login)

	authenticated := app.Group("/")
	authenticated.Use(middlewares.AuthMiddleware)

	authenticated.Get("/me", h.GetMe)

	ownerOrAdmin := authenticated.Group("/users/:userId")
	ownerOrAdmin.Use(middlewares.RequireOwnerOrAdmin())

	ownerOrAdmin.Get("/", h.GetUser)
	ownerOrAdmin.Put("/", sampleHanlder)
	ownerOrAdmin.Delete("/", h.DeleteUser)

	adminOnly := authenticated.Group("/users")
	adminOnly.Use(middlewares.RequireRoleIs(models.Admin))

	adminOnly.Get("/", h.GetUsers)

}
