package auth

import (
	"time"

	"github.com/avtorsky/gphrmart/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(user *models.User, signingKey []byte, expireDuration time.Duration) *jwt.Token {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   user.ID,
		Username: user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token
}
