package service

import (
	"log"

	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
	"lqkhoi-go-http-api/pkg/utils"
	"lqkhoi-go-http-api/pkg/custom_structs"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user *dto.CreateUserRequest) error
	FindByID(id int) (*models.User, error)
	Login(rq dto.LoginRequest) (string, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user models.User) error
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

func (s *userService) CreateUser(userRequest *dto.CreateUserRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	userRequest.Password = string(hashedPassword)
	user := userRequest.MapToUser()
	return s.userRepository.Create(user)
}

func (s *userService) FindByID(id int) (*models.User, error) {
	return s.userRepository.FindByID(id)
}

func (s *userService) Login(rq dto.LoginRequest) (string, error) {
	var user *models.User
	var err error

	user, err = s.userRepository.FindByEmail(rq.Credential)
	if err != nil {
		log.Printf("Internal Database Error looking up email %q: %v", rq.Credential, err)
		return "", custom_structs.ErrDatabaseFail
	}

	if user == nil {
		user, err = s.userRepository.FindByUsername(rq.Credential)
		if err != nil {
			log.Printf("Internal Database Error looking up username %q: %v", rq.Credential, err)
			return "", custom_structs.ErrDatabaseFail
		}
	}

	if user == nil {
		log.Printf("No user found with credential: %q", rq.Credential)
		return "", custom_structs.ErrUsernameOrEmailNotExist
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(rq.Password))
	if err == nil {
		log.Printf("User %q provide corrected password.", rq.Credential)
		token, err := utils.GenerateToken(user.ID, rq.Credential, user.Role)
		if err != nil {
			log.Printf("Can not sign token for user %q.", rq.Credential)
			return "", custom_structs.ErrTokenCanNotBeSigned
		}
		return token, nil
	} else if err == bcrypt.ErrMismatchedHashAndPassword {
		log.Printf("Incorrect Password Login Attempt for credential: %q", rq.Credential)
		return "", custom_structs.ErrPasswordIncorrect
	} else {
		log.Printf("Error comparing password for credential %q: %v", rq.Credential, err)
		return "", custom_structs.ErrInternalServer
	}

}

func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.userRepository.List()
}

func (s *userService) UpdateUser(user models.User) error{
	return s.UpdateUser(user)
}

func (s *userService) DeleteUser(id int) error {
	return s.userRepository.Delete(id)
}


