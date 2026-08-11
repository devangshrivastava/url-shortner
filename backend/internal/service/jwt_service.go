package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenExpiry = 24 * time.Hour

var (
	ErrJWTSecretMissing = errors.New("JWT_SECRET is not set")
	ErrInvalidToken     = errors.New("invalid token")
)

type JWTService struct {
	secret []byte
}

type tokenClaims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTServiceFromEnv() (*JWTService, error) {
	return NewJWTService(os.Getenv("JWT_SECRET"))
}

func NewJWTService(secret string) (*JWTService, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrJWTSecretMissing
	}

	return &JWTService{secret: []byte(secret)}, nil
}

func (s *JWTService) GenerateToken(userID int64) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseAndValidateToken(tokenString string) (int64, error) {
	claims := &tokenClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return s.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid || claims.ExpiresAt == nil || claims.UserID <= 0 {
		return 0, ErrInvalidToken
	}

	return claims.UserID, nil
}
