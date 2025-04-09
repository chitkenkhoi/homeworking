package dto

import (
	"lqkhoi-go-http-api/internal/models"
)

// CreateUserRequest represents the request body for creating a new user.
type CreateUserRequest struct {
	// Email is the user's email address.
	Email     string          `json:"email" validate:"required,email" example:"john.doe@example.com"`
	// Password is the user's password.
	Password  string          `json:"password" validate:"required,min=8" example:"securepassword123"`
	// FirstName is the user's first name.
	FirstName string          `json:"first_name" validate:"required,min=2,max=100" example:"John"`
	// LastName is the user's last name.
	LastName  string          `json:"last_name" validate:"required,min=2,max=100" example:"Doe"`
	// Role is the user's role in the system.
	Role      models.UserRole `json:"role" validate:"omitempty,oneof=TEAM_MEMBER PROJECT_MANAGER ADMIN" example:"TEAM_MEMBER"`
}

func (cur *CreateUserRequest) MapToUser() *models.User {
	return &models.User{
		Email:     cur.Email,
		Password:  cur.Password,
		FirstName: cur.FirstName,
		LastName:  cur.LastName,
		Role:      cur.Role,
	}
}

// LoginRequest represents the request body for user login.
type LoginRequest struct {
	// Email is the user's email address.
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	// Password is the user's password.
	Password string `json:"password" validate:"required,min=8" example:"securepassword123"`
}

// UserResponse represents the response body for user details.
type UserResponse struct {
	// ID is the unique identifier of the user.
	ID                 int    `json:"id" example:"42"`
	// Email is the user's email address.
	Email              string `json:"email" example:"john.doe@example.com"`
	// Role is the user's role in the system.
	Role               string `json:"role" example:"TEAM_MEMBER"`
	// FirstName is the user's first name.
	FirstName          string `json:"first_name" example:"John"`
	// LastName is the user's last name.
	LastName           string `json:"last_name" example:"Doe"`
	// CurrentProjectID is the optional ID of the user's current project.
	CurrentProjectID   int    `json:"current_project_id,omitempty" example:"1"`
	// CurrentProjectName is the optional name of the user's current project.
	CurrentProjectName string `json:"current_project_name,omitempty" example:"Website Redesign"`
}

func MapToUserDto(user *models.User) *UserResponse {
	ur := &UserResponse{}
	ur.ID = user.ID
	ur.Email = user.Email
	ur.Role = string(user.Role)
	ur.FirstName = user.FirstName
	ur.LastName = user.LastName
	if user.CurrentProjectID != nil {
		ur.CurrentProjectID = *user.CurrentProjectID
	}
	if user.CurrentProject != nil {
		ur.CurrentProjectName = user.CurrentProject.Name
	}
	return ur
}

func MapToUserDtoSlice(users []*models.User) []UserResponse {
	urs := make([]UserResponse, 0, len(users))
	for _, user := range users { 
		ur := MapToUserDto(user)
		urs = append(urs, *ur)
	}
	return urs
}

// UpdateUserRequest represents the request body for updating an existing user.
type UpdateUserRequest struct {
	// FirstName is the optional new first name of the user.
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100" example:"Johnny"`
	// LastName is the optional new last name of the user.
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100" example:"Smith"`
}
