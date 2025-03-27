package middlewares

import (
	"strconv"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"github.com/gofiber/fiber/v2"
)

func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claimsData := c.Locals("user_claims")
		// if claimsData == nil {
		//     return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		//         "error": "Unauthorized: Missing claims",
		//     })
		// }
		userClaims, ok := claimsData.(*structs.Claims)
		if !ok || userClaims == nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error: Invalid claims format",
			})
		}

		if userClaims.Role != models.Admin {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: Admin access required",
			})
		}

		return c.Next()
	}
}

func RequireOwner() fiber.Handler {
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
