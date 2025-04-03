package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"

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

func (s *userService) Login(ctx context.Context, rq dto.LoginRequest) (string, error) {
	var user *models.User
	var err error

	user, err = s.userRepository.FindByEmail(ctx, rq.Email)
	if err != nil {
		log.Printf("Internal Database Error looking up email %q: %v", rq.Email, err)
		return "", structs.ErrDatabaseFail
	}

	if user == nil {
		log.Printf("No user found with email: %q", rq.Email)
		return "", structs.ErrEmailNotExist
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rq.Password))
	if err == nil {
		log.Printf("User %q provide corrected password.", rq.Email)
		token, err := utils.GenerateToken(user.ID, rq.Email, user.Role)
		if err != nil {
			log.Printf("Can not sign token for user %q.", rq.Email)
			return "", structs.ErrTokenCanNotBeSigned
		}
		return token, nil
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		log.Printf("Incorrect Password Login Attempt for email: %q", rq.Email)
		return "", structs.ErrPasswordIncorrect
	} else {
		log.Printf("Error comparing password for email %q: %v", rq.Email, err)
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
	err := s.userRepository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, structs.ErrUserNotExist) {
			slog.Error("Can not find user with", "id", id)
			return err
		} else {
			return structs.ErrDatabaseFail
		}
	}
	slog.Debug("Deleted user with", "id", id)
	return nil
}
