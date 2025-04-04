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

func SetupUserRoutes(prefixApp fiber.Router, h *handler.UserHandler, lm fiber.Handler) {
	log := prefixApp.Group("/")
	log.Use(lm)
	
	log.Post("/users", h.CreateUserHandler)
	log.Post("/login", h.Login)

	authenticated := log.Group("/")
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
