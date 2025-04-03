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

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUserHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "UserHandler",
		"handler", "CreateUserHandler",
	)

	logger.Debug("Parsing JSON input")

	input := &dto.CreateUserRequest{}
	if err := c.BodyParser(input); err != nil {

		logger.Error("Can not parse JSON", "error", err.Error())

		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}

	logger.Debug("Validating user's input", "input", input)

	errs := utils.ValidateStruct(*input) // Pass the struct value
	if errs != nil {
		logger.Error("Validation failed", "error", errs)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	user := input.MapToUser()

	logger.Debug("Successfully mapping data to user model", "user", user)

	if user, err := h.userService.CreateUser(ctx, user); err != nil {
		logger.Error("Failed to create user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(
			createErrorResponse("Failed to create user", err.Error()))
	} else {
		output := dto.MapToUserDto(user)

		logger.Debug("Successfully map to response", "response", output)

		return c.Status(fiber.StatusCreated).JSON(
			createSuccessResponse("user is created", output))
	}
}

func (h *UserHandler) Login(c *fiber.Ctx) error {
	input := &dto.LoginRequest{}
	if err := c.BodyParser(input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	log.Printf("Validation successful for input: %+v\n", *input)

	ctx := c.UserContext()
	if token, err := h.userService.Login(ctx, *input); err != nil {
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

	ctx := c.UserContext()
	user, err := h.userService.FindByID(ctx, userClaims.UserID)

	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(createErrorResponse(
				"Internal database error", err.Error()))
		} else {
			return c.Status(fiber.StatusNotFound).JSON(createErrorResponse(
				"User not found", err.Error()))
		}
	} else {
		output := dto.MapToUserDto(user)
		return c.Status(fiber.StatusFound).JSON(createSuccessResponse("Found user", output))
	}
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("userId")
	if err != nil {
		slog.Error("invalid user id", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid user id", nil))
	}

	ctx := c.UserContext()
	user, err := h.userService.FindByID(ctx, id)

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
		output := dto.MapToUserDto(user)
		return c.Status(fiber.StatusFound).JSON(
			createSuccessResponse("found user", output))
	}
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := utils.LoggerFromContext(ctx).With(
		"component", "UserHandler",
		"handler", "UpdateUser",
	)
	id, err := verifyIdParamInt(c, logger, "userId")
	if err != nil {
		logger.Error("Invalid project id")
		return err
	}

	input := &dto.UpdateUserRequest{}
	if err = c.BodyParser(input); err != nil {
		logger.Error("Cannot parse JSON", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Cannot parse JSON", nil))
	}

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "errors", errs)
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse("Validation failed", nil))
	}

	logger.Debug("Validation finish successfully for input", "input", *input)
	updatedUser, err := h.userService.UpdateUser(ctx, id, input)
	if err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal database fail", nil),
			)
		}
		return c.Status(fiber.StatusBadRequest).JSON(createErrorResponse(
			"Validation error", err.Error(),
		))
	}
	output := dto.MapToUserDto(updatedUser)
	return c.Status(fiber.StatusAccepted).JSON(createSuccessResponse(
		"Student has been updated", output,
	))
}

func (h *UserHandler) GetUsers(c *fiber.Ctx) error {
	ctx := c.UserContext()
	if users, err := h.userService.GetAllUsers(ctx); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Internal database error", err.Error()))
	} else {
		output := dto.MapToUserDtoSlice(users)
		return c.Status(fiber.StatusOK).JSON(
			createSliceSuccessResponseGeneric("Found all users", output))
	}
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	id, err := c.ParamsInt("userId")
	if err != nil {
		slog.Error("invalid user id", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("invalid user id", nil))
	}

	ctx := c.UserContext()
	err = h.userService.DeleteUser(ctx, id)
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
