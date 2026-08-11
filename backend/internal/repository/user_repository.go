package repository

import (
	"context"

	"url-shortener/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, email string, passwordHash string) (int64, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
	GetByID(ctx context.Context, id int64) (model.User, error)
}
