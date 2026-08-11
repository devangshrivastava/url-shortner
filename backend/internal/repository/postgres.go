package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/model"
)

type PostgresURLRepository struct {
	db *pgxpool.Pool
}

func InitializePostgresSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			long_url TEXT NOT NULL,
			expires_at TEXT,
			user_id BIGINT REFERENCES users(id),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS clicks (
			id BIGSERIAL PRIMARY KEY,
			code TEXT NOT NULL REFERENCES urls(code) ON DELETE CASCADE,
			clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			ip TEXT,
			user_agent TEXT,
			referer TEXT
		);
	`)
	return err
}

func NewPostgresURLRepository(db *pgxpool.Pool) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func (r *PostgresURLRepository) Save(
	ctx context.Context,
	userID int64,
	code string,
	longURL string,
	expiresAt string,
) error {
	const query = `
		INSERT INTO urls (code, long_url, expires_at, user_id)
		VALUES ($1, $2, NULLIF($3, ''), $4)
	`

	_, err := r.db.Exec(ctx, query, code, longURL, expiresAt, userID)
	return err
}

func (r *PostgresURLRepository) Get(ctx context.Context, code string) (model.URL, error) {
	const query = `
		SELECT code, long_url, expires_at, user_id, created_at
		FROM urls
		WHERE code = $1
	`

	return r.getURL(ctx, query, code)
}

func (r *PostgresURLRepository) SaveClick(ctx context.Context, click model.Click) error {
	const query = `
		INSERT INTO clicks (code, ip, user_agent, referer)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(ctx, query, click.Code, click.IP, click.UserAgent, click.Referer)
	return err
}

func (r *PostgresURLRepository) GetAnalytics(
	ctx context.Context,
	code string,
) (model.Analytics, error) {
	const totalQuery = `SELECT COUNT(*) FROM clicks WHERE code = $1`

	var totalClicks int
	if err := r.db.QueryRow(ctx, totalQuery, code).Scan(&totalClicks); err != nil {
		return model.Analytics{}, err
	}

	const recentQuery = `
		SELECT
			to_char(clicked_at, 'YYYY-MM-DD"T"HH24:MI:SS.USOF'),
			ip,
			user_agent,
			referer
		FROM clicks
		WHERE code = $1
		ORDER BY clicked_at DESC
		LIMIT 10
	`

	rows, err := r.db.Query(ctx, recentQuery, code)
	if err != nil {
		return model.Analytics{}, err
	}
	defer rows.Close()

	recentClicks := []model.ClickInfo{}
	for rows.Next() {
		var click model.ClickInfo
		if err := rows.Scan(&click.ClickedAt, &click.IP, &click.UserAgent, &click.Referer); err != nil {
			return model.Analytics{}, err
		}
		recentClicks = append(recentClicks, click)
	}
	if err := rows.Err(); err != nil {
		return model.Analytics{}, err
	}

	return model.Analytics{
		Code:         code,
		TotalClicks:  totalClicks,
		RecentClicks: recentClicks,
	}, nil
}

func (r *PostgresURLRepository) ListByUserID(
	ctx context.Context,
	userID int64,
) ([]model.URL, error) {
	const query = `
		SELECT code, long_url, expires_at, user_id, created_at
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []model.URL{}
	for rows.Next() {
		url, err := scanURL(rows)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (r *PostgresURLRepository) Update(
	ctx context.Context,
	code string,
	longURL *string,
	expiresAt *string,
) (bool, error) {
	const query = `
		UPDATE urls
		SET
			long_url = COALESCE($2, long_url),
			expires_at = CASE
				WHEN $3::text IS NULL THEN expires_at
				ELSE NULLIF($3::text, '')
			END
		WHERE code = $1
	`

	commandTag, err := r.db.Exec(ctx, query, code, longURL, expiresAt)
	if err != nil {
		return false, err
	}

	return commandTag.RowsAffected() == 1, nil
}

func (r *PostgresURLRepository) Delete(ctx context.Context, code string) (bool, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM clicks WHERE code = $1`, code); err != nil {
		return false, err
	}

	commandTag, err := tx.Exec(ctx, `DELETE FROM urls WHERE code = $1`, code)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return commandTag.RowsAffected() == 1, nil
}

func (r *PostgresURLRepository) getURL(
	ctx context.Context,
	query string,
	value any,
) (model.URL, error) {
	row := r.db.QueryRow(ctx, query, value)
	return scanURL(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanURL(row rowScanner) (model.URL, error) {
	var url model.URL
	var expiresAt pgtype.Text
	var userID pgtype.Int8

	err := row.Scan(
		&url.Code,
		&url.LongURL,
		&expiresAt,
		&userID,
		&url.CreatedAt,
	)
	if err != nil {
		return model.URL{}, err
	}

	if expiresAt.Valid {
		url.ExpiresAt = expiresAt.String
	}
	if userID.Valid {
		url.UserID = &userID.Int64
	}

	return url, nil
}

var _ URLRepository = (*PostgresURLRepository)(nil)
