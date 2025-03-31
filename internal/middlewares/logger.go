package middlewares

import (
	"log/slog"

	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func NewLoggingMiddleware(baseLogger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := uuid.NewString()
		reqLogger := baseLogger.With(
			"request_id", requestID,
			"http_method", c.Method(),
			"http_path", c.Path(),
		)

		ctx := c.UserContext()

		ctx = utils.ContextWithLogger(ctx, reqLogger)

		c.SetUserContext(ctx)

		err := c.Next()

		return err
	}
}
