package dto

import (
	"lqkhoi-go-http-api/internal/models"
)

// CreateUserRequest represents the request body for creating a new user.
type CreateUserRequest struct {
	// Email is the user's email address.
	// @example "john.doe@example.com"
	Email     string          `json:"email" validate:"required,email"`
	// Password is the user's password.
	// @example "securepassword123"
	Password  string          `json:"password" validate:"required,min=8"`
	// FirstName is the user's first name.
	// @example "John"
	FirstName string          `json:"first_name" validate:"required,min=2,max=100"`
	// LastName is the user's last name.
	// @example "Doe"
	LastName  string          `json:"last_name" validate:"required,min=2,max=100"`
	// Role is the user's role in the system.
	// @example "TEAM_MEMBER"
	Role      models.UserRole `json:"role" validate:"omitempty,oneof=TEAM_MEMBER PROJECT_MANAGER ADMIN"`
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
	// @example "john.doe@example.com"
	Email    string `json:"email" validate:"required,email"`
	// Password is the user's password.
	// @example "securepassword123"
	Password string `json:"password" validate:"required,min=8"`
}

// UserResponse represents the response body for user details.
type UserResponse struct {
	// ID is the unique identifier of the user.
	// @example 42
	ID                 int    `json:"id"`
	// Email is the user's email address.
	// @example "john.doe@example.com"
	Email              string `json:"email"`
	// Role is the user's role in the system.
	// @example "TEAM_MEMBER"
	Role               string `json:"role"`
	// FirstName is the user's first name.
	// @example "John"
	FirstName          string `json:"first_name"`
	// LastName is the user's last name.
	// @example "Doe"
	LastName           string `json:"last_name"`
	// CurrentProjectID is the optional ID of the user's current project.
	// @example 1
	CurrentProjectID   int    `json:"current_project_id,omitempty"`
	// CurrentProjectName is the optional name of the user's current project.
	// @example "Website Redesign"
	CurrentProjectName string `json:"current_project_name,omitempty"`
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
	for index, user := range users {
		urs = append(urs, UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Role:      string(user.Role),
			FirstName: user.FirstName,
			LastName:  user.LastName,
		})
		if user.CurrentProjectID != nil {
			urs[index].CurrentProjectID = *user.CurrentProjectID
		}
		if user.CurrentProject != nil {
			urs[index].CurrentProjectName = user.CurrentProject.Name
		}
	}
	return urs
}

// UpdateUserRequest represents the request body for updating an existing user.
type UpdateUserRequest struct {
	// FirstName is the optional new first name of the user.
	// @example "Johnny"
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	// LastName is the optional new last name of the user.
	// @example "Smith"
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
}