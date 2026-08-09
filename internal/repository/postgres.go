package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"url-shortener/internal/model"
)

type PostgresURLRepository struct {
	db *pgxpool.Pool
}

func InitializePostgresSchema(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			long_url TEXT NOT NULL,
			expires_at TEXT
		);

		CREATE TABLE IF NOT EXISTS clicks (
			id BIGSERIAL PRIMARY KEY,
			code TEXT NOT NULL,
			clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			ip TEXT,
			user_agent TEXT,
			referer TEXT
		);
	`)
	return err
}

func NewPostgresURLRepository(db *pgxpool.Pool) *PostgresURLRepository {
	return &PostgresURLRepository{
		db: db,
	}
}

func (r *PostgresURLRepository) Save(
	code string,
	longURL string,
	expiresAt string,
) {
	query := `
		INSERT INTO urls (code, long_url, expires_at)
		VALUES ($1, $2, NULLIF($3, ''))
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		code,
		longURL,
		expiresAt,
	)

	if err != nil {
		panic(err)
	}
}

func (r *PostgresURLRepository) Get(
	code string,
) (string, string, bool) {

	query := `
		SELECT long_url, expires_at
		FROM urls
		WHERE code = $1
	`

	var longURL string
	var expiresAt pgtype.Text

	err := r.db.QueryRow(
		context.Background(),
		query,
		code,
	).Scan(
		&longURL,
		&expiresAt,
	)

	if err == pgx.ErrNoRows {
		return "", "", false
	}

	if err != nil {
		panic(err)
	}

	if !expiresAt.Valid {
		return longURL, "", true
	}

	return longURL, expiresAt.String, true
}

func (r *PostgresURLRepository) SaveClick(click model.Click) {
	query := `
		INSERT INTO clicks (
			code,
			ip,
			user_agent,
			referer
		)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		context.Background(),
		query,
		click.Code,
		click.IP,
		click.UserAgent,
		click.Referer,
	)

	if err != nil {
		panic(err)
	}
}

func (r *PostgresURLRepository) GetAnalytics(
	code string,
) model.Analytics {

	totalQuery := `
		SELECT COUNT(*)
		FROM clicks
		WHERE code = $1
	`

	var totalClicks int

	err := r.db.QueryRow(
		context.Background(),
		totalQuery,
		code,
	).Scan(&totalClicks)

	if err != nil {
		panic(err)
	}

	recentQuery := `
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

	rows, err := r.db.Query(
		context.Background(),
		recentQuery,
		code,
	)

	if err != nil {
		panic(err)
	}

	defer rows.Close()

	recentClicks := []model.ClickInfo{}

	for rows.Next() {
		var click model.ClickInfo

		err := rows.Scan(
			&click.ClickedAt,
			&click.IP,
			&click.UserAgent,
			&click.Referer,
		)

		if err != nil {
			panic(err)
		}

		recentClicks = append(recentClicks, click)
	}

	if err := rows.Err(); err != nil {
		panic(err)
	}

	return model.Analytics{
		Code:         code,
		TotalClicks:  totalClicks,
		RecentClicks: recentClicks,
	}
}
