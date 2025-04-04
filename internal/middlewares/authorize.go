package middlewares

import (
	"fmt"
	"strconv"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"github.com/gofiber/fiber/v2"
)

func RequireRoleIs(role models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsData := c.Locals("user_claims")
		userClaims, ok := claimsData.(*structs.Claims)
		if !ok || userClaims == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error: Invalid claims format",
			})
		}

		if userClaims.Role != role {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": fmt.Sprintf("Forbidden: %v access required", role),
			})
		}

		return c.Next()
	}
}

func RequireOwnerOrAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsData := c.Locals("user_claims")
		userClaims, ok := claimsData.(*structs.Claims)
		if !ok || userClaims == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error: Invalid claims format",
			})
		}

		if userClaims.Role == models.Admin {
			return c.Next()
		}

		if strconv.Itoa(userClaims.UserID) != c.Params("userId") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Owner or admin access required",
			})
		}

		return c.Next()
	}
}

func RequireOwnerOrProjectManager() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsData := c.Locals("user_claims")
		userClaims, ok := claimsData.(*structs.Claims)
		if !ok || userClaims == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error: Invalid claims format",
			})
		}

		if userClaims.Role == models.ProjectManager {
			return c.Next()
		}

		if strconv.Itoa(userClaims.UserID) != c.Params("userId") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Owner or project manager access required",
			})
		}

		return c.Next()
	}
}

func RequireOwnerIfTeamMember() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsData := c.Locals("user_claims")
		userClaims, ok := claimsData.(*structs.Claims)
		if !ok || userClaims == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error: Invalid claims format",
			})
		}

		if userClaims.Role == models.TeamMember {
			if strconv.Itoa(userClaims.UserID) != c.Params("userId") {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"error": "Forbidden: Owner access required",
				})
			}
		}

		return c.Next()
	}
}
