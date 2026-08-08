package repository

import (
	"database/sql"
	"url-shortener/internal/model"
	// "sync"
	_ "modernc.org/sqlite"
)

type SQLiteURLRepository struct {
	db *sql.DB
}

type URLRepository interface {
	Save(code string, longURL string, expiresAt string)
	Get(code string) (string, string, bool)
	SaveClick(click model.Click)
	GetAnalytics(code string) model.Analytics
}

func NewSQLiteURLRepository(dbPath string) (*SQLiteURLRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteURLRepository{db: db}

	err = repo.createTables()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteURLRepository) createTables() error {
	urlsTable := `
		CREATE TABLE IF NOT EXISTS urls (
			code TEXT PRIMARY KEY,
			long_url TEXT NOT NULL,
			expires_at TEXT
		);
		`

	clicksTable := `
	CREATE TABLE IF NOT EXISTS clicks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT NOT NULL,
		clicked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ip TEXT,
		user_agent TEXT,
		referer TEXT
	);
	`

	if _, err := r.db.Exec(urlsTable); err != nil {
		return err
	}

	if _, err := r.db.Exec(clicksTable); err != nil {
		return err
	}

	return nil
}

func (r *SQLiteURLRepository) Save(code string, longURL string, expiresAt string) {
	query := `INSERT INTO urls (code, long_url, expires_at) VALUES (?, ?, ?)`

	_, err := r.db.Exec(query, code, longURL, expiresAt)
	if err != nil {
		panic(err)
	}
}

func (r *SQLiteURLRepository) Get(code string) (string, string, bool) {
	query := `SELECT long_url, expires_at FROM urls WHERE code = ?`

	var longURL string
	var expiresAt sql.NullString

	err := r.db.QueryRow(query, code).Scan(&longURL, &expiresAt)

	if err == sql.ErrNoRows {
		return "", "", false
	}

	if err != nil {
		panic(err)
	}

	return longURL, expiresAt.String, true
}

func (r *SQLiteURLRepository) SaveClick(click model.Click) {
	query := `
	INSERT INTO clicks (code, ip, user_agent, referer)
	VALUES (?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, click.Code, click.IP, click.UserAgent, click.Referer)
	if err != nil {
		panic(err)
	}
}

func (r *SQLiteURLRepository) GetAnalytics(code string) model.Analytics {
	totalQuery := `SELECT COUNT(*) FROM clicks WHERE code = ?`

	var totalClicks int
	err := r.db.QueryRow(totalQuery, code).Scan(&totalClicks)
	if err != nil {
		panic(err)
	}

	recentQuery := `
	SELECT clicked_at, ip, user_agent, referer
	FROM clicks
	WHERE code = ?
	ORDER BY clicked_at DESC
	LIMIT 10
	`

	rows, err := r.db.Query(recentQuery, code)
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

	return model.Analytics{
		Code:         code,
		TotalClicks: totalClicks,
		RecentClicks: recentClicks,
	}
}