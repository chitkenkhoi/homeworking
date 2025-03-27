package handler

import (
	"errors"
	"fmt"
	"log"
	"log/slog"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	userService service.UserService
}

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

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUserHandler(c *fiber.Ctx) error {
	input := &dto.CreateUserRequest{}
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", err.Error()))
	}
	errors := utils.ValidateStruct(*input) // Pass the struct value
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errors))
	}
	log.Printf("Validation successful for input: %+v\n", *input)

	if user, err := h.userService.CreateUser(input); err != nil {
		log.Printf("Failed to create user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to create user", err.Error()))
	} else {
		return c.Status(fiber.StatusCreated).JSON(
			createSuccessResponse("user is created", user))
	}
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	input := &dto.LoginRequest{}
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", err))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}
	log.Printf("Validation successful for input: %+v\n", *input)

	if token, err := h.userService.Login(*input); err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) || errors.Is(err, structs.ErrInternalServer) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", err))
		} else if errors.Is(err, structs.ErrEmailNotExist) || errors.Is(err, structs.ErrPasswordIncorrect) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("User's credential is not correct", err.Error()))
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("Bad request", err.Error()))
		}
	} else {
		return c.Status(fiber.StatusAccepted).JSON(
			createSuccessResponse("user login successfully", struct {
				Token string `json:"token"`
			}{
				Token: token,
			}))
	}
}

func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	claimsData := c.Locals("user_claims")
	userClaims, ok := claimsData.(*structs.Claims)
	if !ok || userClaims == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Internal Server Error: Invalid claims format", nil))
	}
	user, err := h.userService.FindByID(userClaims.UserID)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse(
				"Internal database error", err.Error()))
		} else {
			return c.Status(fiber.StatusNotFound).JSON(createErrorResponse(
				"User not found", err.Error()))
		}
	} else {
		return c.Status(fiber.StatusFound).JSON(createSuccessResponse("Found user", user))
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("userId")
	if err != nil {
		slog.Error("invalid user id", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid user id", nil))
	}
	user, err := h.userService.FindByID(id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(
				createErrorResponse("user is not found",
					fmt.Errorf("user with id %v does not exist", id)))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("internal server error", nil))
		}

	} else {
		return c.Status(fiber.StatusFound).JSON(
			createSuccessResponse("found user", user))
	}
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	if users, err := h.userService.GetAllUsers(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Internal database error", err.Error()))
	} else {
		return c.Status(fiber.StatusOK).JSON(
			createSuccessResponse("Found all users", users))
	}
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("userId")
	if err != nil {
		slog.Error("invalid user id", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid user id", nil))
	}
	err = h.userService.DeleteUser(id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("user is not found",
					fmt.Errorf("user with id %v does not exist", id)))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("internal server error", nil))
		}
	} else {
		return c.Status(fiber.StatusOK).JSON(
			createSuccessResponse("user has been deleted", nil))
	}
}
