package handler

import (
	"github.com/gofiber/fiber/v2"
)

func createErrorResponse(msg string, err any) fiber.Map {
	return fiber.Map{
		"error":   msg,
		"details": err,
	}
}

func createSuccessResponse(msg string, data any) fiber.Map {
	return fiber.Map{
		"message": msg,
		"data":    data,
	}
}

func createSliceSuccessResponseGeneric[T any](msg string, data []T) fiber.Map {
	return fiber.Map{"message": msg, "data": data, "count": len(data)}
}
