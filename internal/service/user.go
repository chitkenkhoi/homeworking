package service

import (
	"errors"
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
	CreateUser(user *models.User) (*models.User, error)
	FindByID(id int) (*models.User, error)
	Login(rq dto.LoginRequest) (string, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
}

type userService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userService{
		userRepository: userRepository,
	}
}

func (s *userService) CreateUser(user *models.User) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		if err == bcrypt.ErrPasswordTooLong {
			return nil, structs.ErrPasswordTooLong
		}
		return nil, err
	}
	user.Password = string(hashedPassword)
	return s.userRepository.Create(user)
}

func (s *userService) FindByID(id int) (*models.User, error) {
	user, err := s.userRepository.FindByID(id)
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

func (s *userService) Login(rq dto.LoginRequest) (string, error) {
	var user *models.User
	var err error

	user, err = s.userRepository.FindByEmail(rq.Email)
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

func (s *userService) GetAllUsers() ([]models.User, error) {
	users, err := s.userRepository.List()
	if err != nil {
		slog.Error("Internal database fail", "error", err)
		return nil, structs.ErrDatabaseFail
	}
	slog.Info("Found a list of user", "users", users)
	return users, nil
}

func (s *userService) UpdateUser(user *models.User) error {
	return s.userRepository.Update(user)
}

func (s *userService) DeleteUser(id int) error {
	err := s.userRepository.Delete(id)
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
