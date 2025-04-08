package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	repomocks "lqkhoi-go-http-api/internal/repository/mocks"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupUserServiceTest(t *testing.T) (context.Context, *gomock.Controller, *repomocks.MockUserRepository, UserService) {
	ctrl := gomock.NewController(t)
	mockUserRepo := repomocks.NewMockUserRepository(ctrl)
	userService := NewUserService(mockUserRepo)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := utils.ContextWithLogger(context.Background(), logger)

	return ctx, ctrl, mockUserRepo, userService
}

func TestUserService_CreateUser(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	plainPassword := "password123"
	userInput := &models.User{
		Email:     "test@example.com",
		Password:  plainPassword,
		FirstName: "Test",
		LastName:  "User",
		Role:      models.TeamMember,
	}

	expectedUserAfterCreate := *userInput
	expectedUserAfterCreate.ID = 1

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) (*models.User, error) {
				assert.Equal(t, userInput.Email, user.Email)
				assert.NotEqual(t, plainPassword, user.Password)
				err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(plainPassword))
				require.NoError(t, err, "Password passed to repo should be a valid hash of the plain password")

				createdUser := *user
				createdUser.ID = expectedUserAfterCreate.ID
				return &createdUser, nil
			}).Times(1)

		createdUser, err := service.CreateUser(ctx, userInput)

		require.NoError(t, err)
		require.NotNil(t, createdUser)
		assert.Equal(t, expectedUserAfterCreate.ID, createdUser.ID)
		assert.Equal(t, expectedUserAfterCreate.Email, createdUser.Email)
		assert.NotEqual(t, plainPassword, createdUser.Password)
	})

	t.Run("Failure - Repository Error", func(t *testing.T) {
		repoErr := structs.ErrDataViolateConstraint
		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil, repoErr).
			Times(1)

		createdUser, err := service.CreateUser(ctx, userInput)

		require.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, errors.Is(err, repoErr))
	})

	t.Run("Failure - Password Too Long (Bcrypt Error)", func(t *testing.T) {
		longPasswordUser := &models.User{
			Password: strings.Repeat("a", 73),
			Email:    "long@pass.com",
		}

		createdUser, err := service.CreateUser(ctx, longPasswordUser)

		require.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, errors.Is(err, structs.ErrPasswordTooLong))
	})
}

func TestUserService_FindByID(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	userID := 1
	expectedUser := &models.User{ID: userID, Email: "found@example.com"}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(expectedUser, nil).
			Times(1)

		user, err := service.FindByID(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(nil, structs.ErrUserNotExist).
			Times(1)

		user, err := service.FindByID(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, structs.ErrUserNotExist))
	})

	t.Run("Failure - Database Error", func(t *testing.T) {
		dbErr := errors.New("some database error")
		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(nil, dbErr).
			Times(1)

		user, err := service.FindByID(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, structs.ErrDatabaseFail))
	})
}

func TestUserService_FindValidTeamMembersForAssignment(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	userIDs := []int{1, 2, 3, 4, 5}
	projID := 10

	user1 := &models.User{ID: 1, Role: models.TeamMember, CurrentProjectID: nil}
	user2 := &models.User{ID: 2, Role: models.ProjectManager, CurrentProjectID: nil}
	user3 := &models.User{ID: 3, Role: models.TeamMember, CurrentProjectID: &projID}
	user4 := &models.User{ID: 4, Role: models.TeamMember, CurrentProjectID: nil}

	repoResultUsers := []*models.User{user1, user2, user3, user4}

	t.Run("Success - All Valid", func(t *testing.T) {
		validIDs := []int{10, 11}
		validUser1 := &models.User{ID: 10, Role: models.TeamMember, CurrentProjectID: nil}
		validUser2 := &models.User{ID: 11, Role: models.TeamMember, CurrentProjectID: nil}

		mockUserRepo.EXPECT().
			FindByIDs(ctx, validIDs).
			Return([]*models.User{validUser1, validUser2}, nil).
			Times(1)

		resultIDs, err := service.FindValidTeamMembersForAssignment(ctx, validIDs)

		require.NoError(t, err)
		assert.ElementsMatch(t, validIDs, resultIDs)
	})

	t.Run("Success - Empty Input", func(t *testing.T) {
		resultIDs, err := service.FindValidTeamMembersForAssignment(ctx, []int{})
		require.NoError(t, err)
		assert.Empty(t, resultIDs)
	})

	t.Run("Failure - Repository Error", func(t *testing.T) {
		dbErr := errors.New("repo find error")
		mockUserRepo.EXPECT().
			FindByIDs(ctx, userIDs).
			Return(nil, dbErr).
			Times(1)

		resultIDs, err := service.FindValidTeamMembersForAssignment(ctx, userIDs)

		require.Error(t, err)
		assert.Nil(t, resultIDs)
		assert.ErrorContains(t, err, "failed to retrieve user data")
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("Partial Success - Some Invalid Users", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindByIDs(ctx, userIDs).
			Return(repoResultUsers, nil).
			Times(1)

		resultIDs, err := service.FindValidTeamMembersForAssignment(ctx, userIDs)

		require.Error(t, err)
		assert.ErrorContains(t, err, "validation failed for some users")
		assert.ErrorContains(t, err, "user 5 not found")
		assert.ErrorContains(t, err, fmt.Sprintf("user %d has incorrect role '%s'", user2.ID, user2.Role))
		assert.ErrorContains(t, err, fmt.Sprintf("user %d is already assigned to project %d", user3.ID, *user3.CurrentProjectID))

		expectedValidIDs := []int{1, 4}
		assert.ElementsMatch(t, expectedValidIDs, resultIDs)
	})
}

func TestUserService_AssignUsersToProject(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	projectID := 50
	userIDs := []int{101, 102}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			AssignUsersToProject(ctx, projectID, userIDs).
			Return(nil).
			Times(1)

		err := service.AssignUsersToProject(ctx, projectID, userIDs)
		require.NoError(t, err)
	})

	t.Run("Success - Empty User IDs", func(t *testing.T) {
		err := service.AssignUsersToProject(ctx, projectID, []int{})
		require.NoError(t, err)
	})

	t.Run("Failure - Repository Error", func(t *testing.T) {
		dbErr := errors.New("assignment failed")
		mockUserRepo.EXPECT().
			AssignUsersToProject(ctx, projectID, userIDs).
			Return(dbErr).
			Times(1)

		err := service.AssignUsersToProject(ctx, projectID, userIDs)

		require.Error(t, err)
		assert.True(t, errors.Is(err, structs.ErrDatabaseFail))
	})
}

func TestUserService_Login(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	email := "login@example.com"
	plainPassword := "correctPassword"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)

	loginReq := dto.LoginRequest{
		Email:    email,
		Password: plainPassword,
	}

	dbUser := &models.User{
		ID:        25,
		Email:     email,
		Password:  string(hashedPassword),
		Role:      models.Admin,
		FirstName: "Login",
		LastName:  "User",
	}


	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindByEmail(ctx, email).
			Return(dbUser, nil).
			Times(1)

		token, err := service.Login(ctx, loginReq)

		require.NoError(t, err)
		assert.NotEmpty(t, token, "Token should not be empty on successful login")
	})

	t.Run("Failure - Email Not Found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			FindByEmail(ctx, email).
			Return(nil, structs.ErrEmailNotExist).
			Times(1)

		token, err := service.Login(ctx, loginReq)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, structs.ErrEmailNotExist)
		assert.ErrorContains(t, err, "fail to find email")
	})

	t.Run("Failure - Incorrect Password", func(t *testing.T) {
		wrongPasswordReq := dto.LoginRequest{
			Email:    email,
			Password: "wrongPassword",
		}
		mockUserRepo.EXPECT().
			FindByEmail(ctx, email).
			Return(dbUser, nil).
			Times(1)

		token, err := service.Login(ctx, wrongPasswordReq)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, structs.ErrPasswordIncorrect)
	})

	t.Run("Failure - Repository Error on FindByEmail", func(t *testing.T) {
		dbErr := errors.New("db lookup failed")
		mockUserRepo.EXPECT().
			FindByEmail(ctx, email).
			Return(nil, dbErr).
			Times(1)

		token, err := service.Login(ctx, loginReq)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, structs.ErrDatabaseFail)
	})

	t.Run("Failure - Bcrypt Internal Error (Less Common)", func(t *testing.T) {
		malformedUser := &models.User{
			ID:       26,
			Email:    email,
			Password: "not-a-valid-bcrypt-hash",
			Role:     models.TeamMember,
		}
		mockUserRepo.EXPECT().
			FindByEmail(ctx, email).
			Return(malformedUser, nil).
			Times(1)

		token, err := service.Login(ctx, loginReq)

		require.Error(t, err)
		assert.Empty(t, token)
		assert.ErrorIs(t, err, structs.ErrInternalServer)
		assert.NotErrorIs(t, err, structs.ErrPasswordIncorrect)
	})

}

func TestUserService_GetAllUsers(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	expectedUsers := []*models.User{
		{ID: 1, Email: "user1@example.com"},
		{ID: 2, Email: "user2@example.com"},
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			List(ctx).
			Return(expectedUsers, nil).
			Times(1)

		users, err := service.GetAllUsers(ctx)

		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
	})

	t.Run("Failure - Repository Error", func(t *testing.T) {
		dbErr := errors.New("list failed")
		mockUserRepo.EXPECT().
			List(ctx).
			Return(nil, dbErr).
			Times(1)

		users, err := service.GetAllUsers(ctx)

		require.Error(t, err)
		assert.Nil(t, users)
		assert.ErrorIs(t, err, structs.ErrDatabaseFail)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	userID := 5
	firstName := "NewFirst"
	lastName := "NewLast"

	updateReq := &dto.UpdateUserRequest{
		FirstName: &firstName,
		LastName:  &lastName,
	}

	expectedUpdateMap := map[string]any{
		"first_name": firstName,
		"last_name":  lastName,
	}

	userAfterUpdate := &models.User{
		ID:        userID,
		Email:     "original@example.com",
		FirstName: firstName,
		LastName:  lastName,
		Role:      models.TeamMember,
	}

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Update(ctx, userID, expectedUpdateMap).
			Return(nil).
			Times(1)

		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(userAfterUpdate, nil).
			Times(1)

		updatedUser, err := service.UpdateUser(ctx, userID, updateReq)

		require.NoError(t, err)
		require.NotNil(t, updatedUser)
		assert.Equal(t, userAfterUpdate, updatedUser)
	})

	t.Run("Success - No Fields to Update", func(t *testing.T) {
		emptyReq := &dto.UpdateUserRequest{}
		currentUser := &models.User{ID: userID, Email: "current@example.com"}

		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(currentUser, nil).
			Times(1)

		updatedUser, err := service.UpdateUser(ctx, userID, emptyReq)

		require.NoError(t, err)
		require.NotNil(t, updatedUser)
		assert.Equal(t, currentUser, updatedUser)
	})

	t.Run("Failure - Repository Update Error", func(t *testing.T) {
		dbErr := errors.New("update failed")
		mockUserRepo.EXPECT().
			Update(ctx, userID, expectedUpdateMap).
			Return(dbErr).
			Times(1)

		updatedUser, err := service.UpdateUser(ctx, userID, updateReq)

		require.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.ErrorIs(t, err, structs.ErrDatabaseFail)
	})

	t.Run("Failure - Repository Update Error (Not Found)", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Update(ctx, userID, expectedUpdateMap).
			Return(structs.ErrUserNotExist).
			Times(1)

		updatedUser, err := service.UpdateUser(ctx, userID, updateReq)

		require.Error(t, err)
		assert.Nil(t, updatedUser)
		assert.ErrorIs(t, err, structs.ErrUserNotExist)
		assert.ErrorContains(t, err, "repository failed to update user")
	})

	t.Run("Failure - Repository FindByID Error After Successful Update", func(t *testing.T) {
		findErr := errors.New("find failed after update")

		mockUserRepo.EXPECT().
			Update(ctx, userID, expectedUpdateMap).
			Return(nil).
			Times(1)

		mockUserRepo.EXPECT().
			FindByID(ctx, userID).
			Return(nil, findErr).
			Times(1)

		updatedUser, err := service.UpdateUser(ctx, userID, updateReq)

		require.NoError(t, err, "SUT currently ignores FindByID error after update")
		assert.Nil(t, updatedUser, "SUT currently returns nil if FindByID fails after update")
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx, ctrl, mockUserRepo, service := setupUserServiceTest(t)
	defer ctrl.Finish()

	userID := 99

	t.Run("Success", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Delete(ctx, userID).
			Return(nil).
			Times(1)

		err := service.DeleteUser(ctx, userID)
		require.NoError(t, err)
	})

	t.Run("Failure - Not Found", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Delete(ctx, userID).
			Return(structs.ErrUserNotExist).
			Times(1)

		err := service.DeleteUser(ctx, userID)
		require.Error(t, err)
		assert.ErrorIs(t, err, structs.ErrUserNotExist)
	})

	t.Run("Failure - Database Error", func(t *testing.T) {
		dbErr := errors.New("delete failed")
		mockUserRepo.EXPECT().
			Delete(ctx, userID).
			Return(dbErr).
			Times(1)

		err := service.DeleteUser(ctx, userID)
		require.Error(t, err)
		assert.ErrorIs(t, err, structs.ErrDatabaseFail)
	})
}