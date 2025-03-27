package utils

import (
	"time"

	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/pkg/structs"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID int, credential string, role models.UserRole) (string, error) {
	claims := &structs.Claims{
		UserID:     userID,
		Credential: credential,
		Role:       role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "LeQuangKhoi",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(GetenvStringValue("JWT_SECRET", "randomkey")))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}
