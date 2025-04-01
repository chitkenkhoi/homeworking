package handler

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

func verifyIdParamInt(c *fiber.Ctx, baseLogger *slog.Logger, param string) (int, error) {
	logger := baseLogger.With(
		"method", "verifyIdParamInt",
		"param", param,
	)

	id, err := c.ParamsInt(param)
	if err != nil || id <= 0 {
		logger.Error("Error parsing ID",
			"ID", c.Params(param),
			"error", err)
		return 0, c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Invalid ID format", nil))
	}

	logger.Debug("Valid ID parameter", "ID", id)

	return id, nil
}
