package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/model"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Create(
	ctx context.Context,
	email string,
	passwordHash string,
) (int64, error) {
	const query = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, email, passwordHash).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *PostgresUserRepository) GetByEmail(
	ctx context.Context,
	email string,
) (model.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	return r.getUser(ctx, query, email)
}

func (r *PostgresUserRepository) GetByID(
	ctx context.Context,
	id int64,
) (model.User, error) {
	const query = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	return r.getUser(ctx, query, id)
}

func (r *PostgresUserRepository) getUser(
	ctx context.Context,
	query string,
	value any,
) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, query, value).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}
