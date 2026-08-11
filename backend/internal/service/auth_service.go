package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

var (
	ErrInvalidEmail       = errors.New("email is required")
	ErrPasswordTooShort   = errors.New("password must be at least 8 characters")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type AuthService struct {
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) Signup(
	ctx context.Context,
	email string,
	password string,
) (int64, error) {
	if email == "" {
		return 0, ErrInvalidEmail
	}

	if len(password) < 8 {
		return 0, ErrPasswordTooShort
	}

	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return 0, ErrEmailAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return 0, err
	}

	id, err := s.users.Create(ctx, email, string(passwordHash))
	if err != nil {
		if isUniqueViolation(err) {
			return 0, ErrEmailAlreadyExists
		}
		return 0, err
	}

	return id, nil
}

func (s *AuthService) Login(
	ctx context.Context,
	email string,
	password string,
) (model.User, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, ErrInvalidCredentials
		}
		return model.User{}, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return model.User{}, ErrInvalidCredentials
	}

	user.PasswordHash = ""
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
