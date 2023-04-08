package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Token struct {
	Access  string
}

func GenerateNewToken(id string) (*Token, error) {
	secret := os.Getenv("JWT_SECRET_KEY")
	claims := jwt.MapClaims{}

	claims["id"] = id
	claims["created"] = time.Now().Unix()
	claims["expires"] = time.Now().Add(time.Hour * 72).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, err
	}

	return &Token{
		Access: t,
	}, nil
}