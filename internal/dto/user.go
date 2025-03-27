package dto

import(
	"lqkhoi-go-http-api/internal/models"
)
type CreateUserRequest struct {
	Email     string   `json:"email" validate:"required,email"`
	Password  string   `json:"password" validate:"required,min=8"`
	FirstName string   `json:"first_name" validate:"required,min=2,max=100"` 
	LastName  string   `json:"last_name" validate:"required,min=2,max=100"`  
	Role      models.UserRole `json:"role" validate:"omitempty,oneof=TEAM_MEMBER PROJECT_MANAGER ADMIN"`
}

func (cur *CreateUserRequest)MapToUser()*models.User{
	return &models.User{
		Email: cur.Email,
		Password: cur.Password,
		FirstName: cur.FirstName,
		LastName: cur.LastName,
		Role: cur.Role,
	}
}

type LoginRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}


