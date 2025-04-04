package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"strings"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/pkg/structs"
	"lqkhoi-go-http-api/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	FindByID(ctx context.Context, id int) (*models.User, error)
	FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error)
	AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) error
	Login(ctx context.Context, rq dto.LoginRequest) (string, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	UpdateUser(ctx context.Context, userID int,
		data *dto.UpdateUserRequest) (*models.User, error)
	DeleteUser(ctx context.Context, id int) error
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		if err == bcrypt.ErrPasswordTooLong {
			return nil, structs.ErrPasswordTooLong
		}
		return nil, err
	}
	user.Password = string(hashedPassword)
	return s.userRepository.Create(ctx, user)
}

func (s *userService) FindByID(ctx context.Context, id int) (*models.User, error) {
	user, err := s.userRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			slog.Error("User does not exist", "id", id)
			return nil, err
		}
		return nil, structs.ErrDatabaseFail
	}
	slog.Info("Find user with id", "id", id, "data", user)
	return user, nil
}

func (s *userService) FindValidTeamMembersForAssignment(ctx context.Context, userIDs []int) ([]int, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserService",
		"method", "FindValidTeamMembersForAssignment",
		"userIDs", userIDs,
	)
	logger.Debug("Finding and validating team members for assignment")

	if len(userIDs) == 0 {
		logger.Debug("No user IDs provided for validation")
		return []int{}, nil
	}

	users, err := s.userRepository.FindByIDs(ctx, userIDs)
	if err != nil {
		logger.Error("Failed to fetch users from repository", "error", err)
		return nil, fmt.Errorf("failed to retrieve user data: %w", err)
	}

	logger.Debug("Successfully fetched users from repository", "found_users_count", len(users))

	foundIDs := make(map[int]struct{}, len(users))
	for _, u := range users {
		foundIDs[u.ID] = struct{}{}
	}

	invalidUserMessages := make([]string, 0, 2*len(userIDs))
	validUserIDs := make([]int, 0, len(users))

	for _, reqID := range userIDs {
		if _, ok := foundIDs[reqID]; !ok {
			msg := fmt.Sprintf("user %d not found", reqID)
			logger.Warn("Requested user not found", "user_id", reqID)
			invalidUserMessages = append(invalidUserMessages, msg)
		}
	}

	for _, user := range users {
		userLogger := logger.With("user_id", user.ID)
		isValid := true

		if user.Role != models.TeamMember {
			msg := fmt.Sprintf("user %d has incorrect role '%s' (required: '%s')", user.ID, user.Role, models.TeamMember)
			userLogger.Warn("Invalid role for assignment", "current_role", user.Role, "required_role", models.TeamMember)
			invalidUserMessages = append(invalidUserMessages, msg)
			isValid = false
		}

		if user.CurrentProjectID != nil {
			msg := fmt.Sprintf("user %d is already assigned to project %d", user.ID, *user.CurrentProjectID)
			userLogger.Warn("User already assigned to a project", "project_id", *user.CurrentProjectID)
			invalidUserMessages = append(invalidUserMessages, msg)
			isValid = false
		}

		if isValid {
			userLogger.Debug("User is eligible for assignment")
			validUserIDs = append(validUserIDs, user.ID)
		}
	}

	if len(invalidUserMessages) > 0 {
		joinedMessages := strings.Join(invalidUserMessages, "; ")
		logger.Warn("Some users failed validation for assignment", "fail_count", len(invalidUserMessages), "valid_count", len(validUserIDs), "errors", joinedMessages)
		return validUserIDs, fmt.Errorf("validation failed for some users: %s", joinedMessages)
	}

	logger.Info("All requested users validated successfully for assignment", "valid_count", len(validUserIDs))
	return validUserIDs, nil
}

func (s *userService) AssignUsersToProject(ctx context.Context, projectID int, userIDs []int) error {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserService",
		"method", "AssignUsersToProject",
		"project_id", projectID,
		"userIDs", userIDs,
	)
	logger.Debug("Starting user assignment to project")

	if len(userIDs) == 0 {
		logger.Debug("No user IDs provided for assignment")
		return nil
	}

	err := s.userRepository.AssignUsersToProject(ctx, projectID, userIDs)
	if err != nil {
		logger.Error("Failed to assign users to project in repository", "error", err)
		return structs.ErrDatabaseFail
	}

	logger.Info("Successfully assigned users to project", "valid_count", len(userIDs))
	return nil
}

func (s *userService) Login(ctx context.Context, rq dto.LoginRequest) (string, error) {
	baseLogger := utils.LoggerFromContext(ctx)
	logger := baseLogger.With(
		"component", "UserService",
		"method", "Login",
	)
	var user *models.User
	var err error

	user, err = s.userRepository.FindByEmail(ctx, rq.Email)
	if err != nil {
		logger.Error("Internal daatabase error looking up email", "email", rq.Email, "error", err.Error())
		return "", structs.ErrDatabaseFail
	}

	if user == nil {
		logger.Warn("No user found", "email", rq.Email)
		return "", structs.ErrEmailNotExist
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rq.Password))
	if err == nil {
		logger.Info("User provide corrected password","email",rq.Email)
		token, err := utils.GenerateToken(user.ID, rq.Email, user.Role)
		if err != nil {
			logger.Error("Can not sign token for user","email",rq.Email)
			return "", structs.ErrTokenCanNotBeSigned
		}
		return token, nil
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		logger.Error("Incorrect Password Login Attempt for email","email", rq.Email)
		logger.Debug("Incorrect Password Login Attempt for email","email", rq.Email,"provided_password",rq.Password)
		return "", structs.ErrPasswordIncorrect
	} else {
		logger.Error("Error comparing password for email","email", rq.Email,"error",err.Error())
		return "", structs.ErrInternalServer
	}

}

func (s *userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	users, err := s.userRepository.List(ctx)
	if err != nil {
		slog.Error("Internal database fail", "error", err)
		return nil, structs.ErrDatabaseFail
	}
	slog.Info("Found a list of user", "users", users)
	return users, nil
}

func (s *userService) UpdateUser(ctx context.Context, userID int,
	data *dto.UpdateUserRequest) (*models.User, error) {
	logger := utils.LoggerFromContext(ctx).With(
		"component", "UserService",
		"handler", "UpdateUser",
		"user_id", userID,
	)

	logger.Debug("Starting to update user")
	updateMap := make(map[string]any)

	if data.FirstName != nil {
		updateMap["first_name"] = *data.FirstName
	}
	if data.LastName != nil {
		updateMap["last_name"] = *data.LastName
	}
	if len(updateMap) == 0 {
		logger.Info("No fields to update, returning current user")
		return s.userRepository.FindByID(ctx, userID)
	}

	logger.Debug("Attempting project update operation", "input", updateMap)

	if err := s.userRepository.Update(ctx, userID, updateMap); err != nil {
		logger.Error("Failed to update user in repository", "error", err)
		if errors.Is(err, structs.ErrUserNotExist) {
			return nil, fmt.Errorf("repository failed to update user: %w", err)
		}
		return nil, structs.ErrDatabaseFail
	}

	logger.Info("Succesfully updated")

	updatedProject, _ := s.userRepository.FindByID(ctx, userID)
	return updatedProject, nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
	logger := utils.LoggerFromContext(ctx).With(
		"component", "UserService",
		"handler", "DeleteUser",
		"user_id", id,
	)
	err := s.userRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			logger.Warn("User not found for deletion", "user_id", id)
			return err
		} else {
			logger.Error("Failed to delete user from repository", "error", err)
			return structs.ErrDatabaseFail
		}
	}
	logger.Info("Successfully deleted user", "user_id", id)
	return nil
}
