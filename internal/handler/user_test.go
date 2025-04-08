package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	servicemocks "lqkhoi-go-http-api/internal/service/mocks"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupUserHandlerTest(t *testing.T) (*gomock.Controller, *servicemocks.MockUserService, *UserHandler) {
	ctrl := gomock.NewController(t)
	mockUserService := servicemocks.NewMockUserService(ctrl)
	userHandler := NewUserHandler(mockUserService)
	return ctrl, mockUserService, userHandler
}

func performRequest(t *testing.T, app *fiber.App, method, path string, body io.Reader, headers map[string]string) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	return resp
}

func setupTestAppWithLogger(handler *UserHandler) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			slog.Error("Unhandled error in test app", "error", err, "status", code)
			return ctx.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	app.Use(func(c *fiber.Ctx) error {
		c.SetUserContext(utils.ContextWithLogger(c.UserContext(), logger))
		return c.Next()
	})
	return app
}

// --- Test Cases ---

func TestUserHandler_CreateUserHandler(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Post("/users", handler.CreateUserHandler)

	validInput := dto.CreateUserRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Role:      models.TeamMember,
	}
	validInputJson, _ := json.Marshal(validInput)

	createdUser := &models.User{
		ID:        1,
		Email:     validInput.Email,
		Password:  "hashedpassword",
		FirstName: validInput.FirstName,
		LastName:  validInput.LastName,
		Role:      models.UserRole(validInput.Role),
	}

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) (*models.User, error) {
				assert.Equal(t, validInput.Email, user.Email)
				assert.Equal(t, validInput.Password, user.Password)
				return createdUser, nil
			}).Times(1)

		resp := performRequest(t, app, "POST", "/users", bytes.NewReader(validInputJson), nil)

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "user is created", body["message"])
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(createdUser.ID), data["id"])
		assert.Equal(t, createdUser.Email, data["email"])
		assert.NotContains(t, data, "password")
	})

	t.Run("Bad JSON", func(t *testing.T) {
		resp := performRequest(t, app, "POST", "/users", strings.NewReader("{invalid json"), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation Error", func(t *testing.T) {
		invalidInput := dto.CreateUserRequest{Email: "not-an-email"}
		invalidInputJson, _ := json.Marshal(invalidInput)

		resp := performRequest(t, app, "POST", "/users", bytes.NewReader(invalidInputJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Service Error", func(t *testing.T) {
		serviceErr := errors.New("database constraint error")
		mockUserService.EXPECT().
			CreateUser(gomock.Any(), gomock.Any()).
			Return(nil, serviceErr).
			Times(1)

		resp := performRequest(t, app, "POST", "/users", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestUserHandler_Login(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Post("/login", handler.Login)

	validInput := dto.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	validInputJson, _ := json.Marshal(validInput)
	expectedToken := "mock_jwt_token"

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			Login(gomock.Any(), validInput).
			Return(expectedToken, nil).
			Times(1)

		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "user login successfully", body["message"])
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, expectedToken, data["token"])
	})

	t.Run("Bad JSON", func(t *testing.T) {
		resp := performRequest(t, app, "POST", "/login", strings.NewReader("bad json"), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation Error", func(t *testing.T) {
		invalidInput := dto.LoginRequest{Email: "invalid"}
		invalidJson, _ := json.Marshal(invalidInput)
		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(invalidJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Credentials Incorrect (Email)", func(t *testing.T) {
		mockUserService.EXPECT().
			Login(gomock.Any(), validInput).
			Return("", structs.ErrEmailNotExist).
			Times(1)

		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "User's credential is not correct", body["message"])
	})

	t.Run("Credentials Incorrect (Password)", func(t *testing.T) {
		mockUserService.EXPECT().
			Login(gomock.Any(), validInput).
			Return("", structs.ErrPasswordIncorrect).
			Times(1)

		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "User's credential is not correct", body["message"])
	})

	t.Run("Internal Server Error (DB)", func(t *testing.T) {
		mockUserService.EXPECT().
			Login(gomock.Any(), validInput).
			Return("", structs.ErrDatabaseFail).
			Times(1)

		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Internal server error", body["message"])
	})

	t.Run("Other Bad Request Error", func(t *testing.T) {
		otherErr := errors.New("some token signing issue maybe")
		mockUserService.EXPECT().
			Login(gomock.Any(), validInput).
			Return("", otherErr).
			Times(1)

		resp := performRequest(t, app, "POST", "/login", bytes.NewReader(validInputJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Bad request", body["message"])
	})
}

func TestUserHandler_GetMe(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)

	userID := 10
	userClaims := &structs.Claims{UserID: userID, Role: "admin"}
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_claims", userClaims)
		return c.Next()
	})
	app.Get("/users/me", handler.GetMe)

	foundUser := &models.User{
		ID:        userID,
		Email:     "me@example.com",
		FirstName: "Current",
		LastName:  "User",
		Role:      models.Admin,
	}

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			FindByID(gomock.Any(), userID).
			Return(foundUser, nil).
			Times(1)

		resp := performRequest(t, app, "GET", "/users/me", nil, nil)
		require.Equal(t, http.StatusFound, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Found user", body["message"])
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(foundUser.ID), data["id"])
		assert.Equal(t, foundUser.Email, data["email"])
	})

	t.Run("Not Found", func(t *testing.T) {
		mockUserService.EXPECT().
			FindByID(gomock.Any(), userID).
			Return(nil, structs.ErrUserNotExist).
			Times(1)

		resp := performRequest(t, app, "GET", "/users/me", nil, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Database Error", func(t *testing.T) {
		mockUserService.EXPECT().
			FindByID(gomock.Any(), userID).
			Return(nil, structs.ErrDatabaseFail).
			Times(1)

		resp := performRequest(t, app, "GET", "/users/me", nil, nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

}

func TestUserHandler_GetUser(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Get("/users/:userId", handler.GetUser)
	targetUserID := 5
	foundUser := &models.User{
		ID:        targetUserID,
		Email:     "target@example.com",
		FirstName: "Target",
		LastName:  "User",
		Role:      models.TeamMember,
	}
	urlPath := fmt.Sprintf("/users/%d", targetUserID)

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			FindByID(gomock.Any(), targetUserID).
			Return(foundUser, nil).
			Times(1)

		resp := performRequest(t, app, "GET", urlPath, nil, nil)
		require.Equal(t, http.StatusFound, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "found user", body["message"])
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(foundUser.ID), data["id"])
		assert.Equal(t, foundUser.Email, data["email"])
	})


	t.Run("Not Found", func(t *testing.T) {
		mockUserService.EXPECT().
			FindByID(gomock.Any(), targetUserID).
			Return(nil, structs.ErrUserNotExist).
			Times(1)

		resp := performRequest(t, app, "GET", urlPath, nil, nil)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		dbErr := errors.New("some other db issue")
		mockUserService.EXPECT().
			FindByID(gomock.Any(), targetUserID).
			Return(nil, dbErr).
			Times(1)

		resp := performRequest(t, app, "GET", urlPath, nil, nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Put("/users/:userId", handler.UpdateUser)

	targetUserID := 15
	firstName := "UpdatedFirst"
	lastName := "UpdatedLast"
	updateInput := dto.UpdateUserRequest{
		FirstName: &firstName,
		LastName:  &lastName,
	}
	updateInputJson, _ := json.Marshal(updateInput)

	updatedUser := &models.User{
		ID:        targetUserID,
		Email:     "original@example.com",
		FirstName: firstName,
		LastName:  lastName,
		Role:      models.TeamMember,
	}
	urlPath := fmt.Sprintf("/users/%d", targetUserID)

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			UpdateUser(gomock.Any(), targetUserID, &updateInput).
			Return(updatedUser, nil).
			Times(1)

		resp := performRequest(t, app, "PUT", urlPath, bytes.NewReader(updateInputJson), nil)
		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Student has been updated", body["message"])
		data, ok := body["data"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, float64(updatedUser.ID), data["id"])
		assert.Equal(t, updatedUser.FirstName, data["first_name"])
		assert.Equal(t, updatedUser.LastName, data["last_name"])
	})

	t.Run("Bad JSON Body", func(t *testing.T) {
		resp := performRequest(t, app, "PUT", urlPath, strings.NewReader("bad json"), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Validation Error in Body", func(t *testing.T) {
	})

	t.Run("Service Validation Error (e.g., Not Found)", func(t *testing.T) {
		serviceErr := structs.ErrUserNotExist
		mockUserService.EXPECT().
			UpdateUser(gomock.Any(), targetUserID, &updateInput).
			Return(nil, serviceErr).
			Times(1)

		resp := performRequest(t, app, "PUT", urlPath, bytes.NewReader(updateInputJson), nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Validation error", body["message"])
		assert.Contains(t, body["details"], serviceErr.Error())
	})

	t.Run("Service Database Error", func(t *testing.T) {
		mockUserService.EXPECT().
			UpdateUser(gomock.Any(), targetUserID, &updateInput).
			Return(nil, structs.ErrDatabaseFail).
			Times(1)

		resp := performRequest(t, app, "PUT", urlPath, bytes.NewReader(updateInputJson), nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Internal database fail", body["message"])
	})
}

func TestUserHandler_GetUsers(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Get("/users", handler.GetUsers)

	foundUsers := []*models.User{
		{ID: 1, Email: "user1@example.com", FirstName: "U", LastName: "One", Role: models.Admin},
		{ID: 2, Email: "user2@example.com", FirstName: "U", LastName: "Two", Role: models.TeamMember},
	}

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			GetAllUsers(gomock.Any()).
			Return(foundUsers, nil).
			Times(1)

		resp := performRequest(t, app, "GET", "/users", nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Found all users", body["message"])
		data, ok := body["data"].([]any)
		require.True(t, ok)
		require.Len(t, data, 2)

		user1, _ := data[0].(map[string]any)
		user2, _ := data[1].(map[string]any)
		assert.Equal(t, float64(1), user1["id"])
		assert.Equal(t, "user1@example.com", user1["email"])
		assert.Equal(t, float64(2), user2["id"])
		assert.Equal(t, "user2@example.com", user2["email"])
	})

	t.Run("Service Error", func(t *testing.T) {
		serviceErr := errors.New("failed to list users")
		mockUserService.EXPECT().
			GetAllUsers(gomock.Any()).
			Return(nil, serviceErr).
			Times(1)

		resp := performRequest(t, app, "GET", "/users", nil, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "Internal database error", body["message"])
		assert.Contains(t, body["details"], serviceErr.Error())
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	ctrl, mockUserService, handler := setupUserHandlerTest(t)
	defer ctrl.Finish()

	app := setupTestAppWithLogger(handler)
	app.Delete("/users/:userId", handler.DeleteUser)

	targetUserID := 99
	urlPath := fmt.Sprintf("/users/%d", targetUserID)

	t.Run("Success", func(t *testing.T) {
		mockUserService.EXPECT().
			DeleteUser(gomock.Any(), targetUserID).
			Return(nil).
			Times(1)

		resp := performRequest(t, app, "DELETE", urlPath, nil, nil)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "user has been deleted", body["message"])
		assert.Nil(t, body["data"])
	})

	t.Run("Invalid User ID Param", func(t *testing.T) {
		resp := performRequest(t, app, "DELETE", "/users/xyz", nil, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockUserService.EXPECT().
			DeleteUser(gomock.Any(), targetUserID).
			Return(structs.ErrUserNotExist).
			Times(1)

		resp := performRequest(t, app, "DELETE", urlPath, nil, nil)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "user is not found", body["message"])
	})

	t.Run("Internal Server Error", func(t *testing.T) {
		dbErr := errors.New("db connection lost")
		mockUserService.EXPECT().
			DeleteUser(gomock.Any(), targetUserID).
			Return(dbErr).
			Times(1)

		resp := performRequest(t, app, "DELETE", urlPath, nil, nil)
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
		var body map[string]any
		err := json.NewDecoder(resp.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, "internal server error", body["message"])
	})
}