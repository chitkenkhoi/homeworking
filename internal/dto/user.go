package dto

import (
	"lqkhoi-go-http-api/internal/models"
)

type CreateUserRequest struct {
	Email     string          `json:"email" validate:"required,email"`
	Password  string          `json:"password" validate:"required,min=8"`
	FirstName string          `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string          `json:"last_name" validate:"required,min=2,max=100"`
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

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UserResponse struct {
	ID                 int    `json:"id"`
	Email              string `json:"email"`
	Role               string `json:"role"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	CurrentProjectID   int    `json:"current_project_id,omitempty"`
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

type UpdateUserRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2,max=100"`
}
