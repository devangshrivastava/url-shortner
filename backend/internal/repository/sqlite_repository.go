package repository

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite"

	"url-shortener/internal/model"
)

type SQLiteURLRepository struct {
	db *sql.DB
}

type URLRepository interface {
	Save(ctx context.Context, userID int64, code string, longURL string, expiresAt string) error
	Get(ctx context.Context, code string) (model.URL, error)
	SaveClick(ctx context.Context, click model.Click) error
	GetAnalytics(ctx context.Context, code string) (model.Analytics, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.URL, error)
	Update(ctx context.Context, code string, longURL *string, expiresAt *string) (bool, error)
	Delete(ctx context.Context, code string) (bool, error)
}

func NewSQLiteURLRepository(dbPath string) (*SQLiteURLRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteURLRepository{db: db}
	if err := repo.createTables(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteURLRepository) createTables() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			long_url TEXT NOT NULL,
			expires_at TEXT,
			user_id INTEGER,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS clicks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			code TEXT NOT NULL,
			clicked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			ip TEXT,
			user_agent TEXT,
			referer TEXT
		);
	`)
	return err
}

func (r *SQLiteURLRepository) Save(
	ctx context.Context,
	userID int64,
	code string,
	longURL string,
	expiresAt string,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO urls (code, long_url, expires_at, user_id) VALUES (?, ?, NULLIF(?, ''), ?)`,
		code,
		longURL,
		expiresAt,
		userID,
	)
	return err
}

func (r *SQLiteURLRepository) Get(ctx context.Context, code string) (model.URL, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT code, long_url, expires_at, user_id, created_at
		FROM urls
		WHERE code = ?
	`, code)

	return scanSQLiteURL(row)
}

func (r *SQLiteURLRepository) SaveClick(ctx context.Context, click model.Click) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO clicks (code, ip, user_agent, referer) VALUES (?, ?, ?, ?)`,
		click.Code,
		click.IP,
		click.UserAgent,
		click.Referer,
	)
	return err
}

func (r *SQLiteURLRepository) GetAnalytics(
	ctx context.Context,
	code string,
) (model.Analytics, error) {
	var totalClicks int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM clicks WHERE code = ?`, code).Scan(&totalClicks); err != nil {
		return model.Analytics{}, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT clicked_at, ip, user_agent, referer
		FROM clicks
		WHERE code = ?
		ORDER BY clicked_at DESC
		LIMIT 10
	`, code)
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

func (r *SQLiteURLRepository) ListByUserID(
	ctx context.Context,
	userID int64,
) ([]model.URL, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, long_url, expires_at, user_id, created_at
		FROM urls
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []model.URL{}
	for rows.Next() {
		url, err := scanSQLiteURL(rows)
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

func (r *SQLiteURLRepository) Update(
	ctx context.Context,
	code string,
	longURL *string,
	expiresAt *string,
) (bool, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE urls
		SET
			long_url = COALESCE(?, long_url),
			expires_at = CASE WHEN ? IS NULL THEN expires_at ELSE NULLIF(?, '') END
		WHERE code = ?
	`, longURL, expiresAt, expiresAt, code)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

func (r *SQLiteURLRepository) Delete(ctx context.Context, code string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM clicks WHERE code = ?`, code); err != nil {
		return false, err
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM urls WHERE code = ?`, code)
	if err != nil {
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected == 1, nil
}

type sqliteRowScanner interface {
	Scan(dest ...any) error
}

func scanSQLiteURL(row sqliteRowScanner) (model.URL, error) {
	var url model.URL
	var expiresAt sql.NullString
	var userID sql.NullInt64
	var createdAt time.Time

	err := row.Scan(&url.Code, &url.LongURL, &expiresAt, &userID, &createdAt)
	if err != nil {
		return model.URL{}, err
	}

	if expiresAt.Valid {
		url.ExpiresAt = expiresAt.String
	}
	if userID.Valid {
		url.UserID = &userID.Int64
	}
	url.CreatedAt = createdAt

	return url, nil
}

var _ URLRepository = (*SQLiteURLRepository)(nil)
