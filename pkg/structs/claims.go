package structs

import (
	"lqkhoi-go-http-api/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   int `json:"user_id"`
	Credential string `json:"credential"`
	Role models.UserRole `json:"role"`
	jwt.RegisteredClaims
}