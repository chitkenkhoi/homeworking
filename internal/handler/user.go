package handler

import (
	"errors"
	"fmt"
	"log/slog"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/service"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUserHandler creates a new user
// @Summary Create a new user
// @Description Creates a new user with the provided details
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} dto.UserSuccessResponse "User created successfully"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or validation failure"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users [post]
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

	errs := utils.ValidateStruct(*input)
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

// Login handles user login
// @Summary User login
// @Description Authenticates a user and returns a token
// @Tags Users
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login credentials"
// @Success 202 {object} dto.TokenResponse "Login successful"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid credentials or input"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "UserHandler",
		"handler", "Login",
	)

	input := &dto.LoginRequest{}
	if err := c.BodyParser(input); err != nil {
		logger.Error("Can not parse JSON", "error", err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Cannot parse JSON", nil))
	}

	logger.Debug("Validating user's input", "input", input)

	errs := utils.ValidateStruct(*input)
	if errs != nil {
		logger.Error("Validation failed", "error", errs)
		return c.Status(fiber.StatusBadRequest).JSON(
			createErrorResponse("Validation failed", errs))
	}

	if token, err := h.userService.Login(ctx, *input); err != nil {
		if errors.Is(err, structs.ErrDatabaseFail) || errors.Is(err, structs.ErrInternalServer) {
			return c.Status(fiber.StatusInternalServerError).JSON(
				createErrorResponse("Internal server error", err))
		} else if errors.Is(err, structs.ErrEmailNotExist) || errors.Is(err, structs.ErrPasswordIncorrect) {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("User's credential is not correct", nil))
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(
				createErrorResponse("Bad request", err.Error()))
		}
	} else {
		return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
			"token": token,
		})
	}
}

// GetMe retrieves the authenticated user's details
// @Summary Get current user
// @Description Retrieves details of the authenticated user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 302 {object} dto.UserSuccessResponse "User found"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /me [get]
func (h *UserHandler) GetMe(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "UserHandler",
		"handler", "GetMe",
	)
	claimsData := c.Locals("user_claims")
	userClaims, _ := claimsData.(*structs.Claims)

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
		logger.Debug("Response is prepared", "response", output)
		return c.Status(fiber.StatusFound).JSON(createSuccessResponse("Found user", output))
	}
}

// GetUser retrieves a user by ID
// @Summary Get user by ID
// @Description Retrieves a user based on the provided ID
// @Tags Users
// @Produce json
// @Param userId path int true "User ID"
// @Success 302 {object} dto.UserSuccessResponse "User found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid ID"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{userId} [get]
func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	baseLogger := utils.LoggerFromContext(ctx)

	logger := baseLogger.With(
		"component", "UserHandler",
		"handler", "GetUser",
	)

	id, err := verifyIdParamInt(c, logger, "userId")
	if err != nil {
		return err
	}

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
		logger.Debug("Response is prepared", "response", output)
		return c.Status(fiber.StatusFound).JSON(
			createSuccessResponse("found user", output))
	}
}

// UpdateUser updates a user's details
// @Summary Update a user
// @Description Updates an existing user's details
// @Tags Users
// @Accept json
// @Produce json
// @Param userId path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 202 {object} dto.UserSuccessResponse "User updated"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid input or ID"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{userId} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx := c.UserContext()
	logger := utils.LoggerFromContext(ctx).With(
		"component", "UserHandler",
		"handler", "UpdateUser",
	)
	id, err := verifyIdParamInt(c, logger, "userId")
	if err != nil {
		logger.Error("Invalid user id")
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

// GetUsers retrieves all users
// @Summary Get all users
// @Description Retrieves a list of all users
// @Tags Users
// @Produce json
// @Success 200 {object} dto.UserSliceSuccessResponse "Users found"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Database error"
// @Router /users [get]
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

// DeleteUser deletes a user by ID
// @Summary Delete a user
// @Description Deletes a user based on the provided ID
// @Tags Users
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} dto.GenericSuccessResponse "User deleted"
// @Failure 400 {object} dto.ErrorResponse "Bad request - Invalid ID or user not found"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /users/{userId} [delete]
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
			createSuccessResponse[any]("user has been deleted", nil))
	}
}
