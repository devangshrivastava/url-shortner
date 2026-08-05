package repository

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

type SQLiteURLRepository struct {
	db *sql.DB
}

func NewSQLiteURLRepository(dbPath string) (*SQLiteURLRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	repo := &SQLiteURLRepository{db: db}

	err = repo.createTable()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *SQLiteURLRepository) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		code TEXT PRIMARY KEY,
		long_url TEXT NOT NULL
	);
	`

	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteURLRepository) Save(code string, longURL string) {
	query := `INSERT INTO urls (code, long_url) VALUES (?, ?)`

	_, err := r.db.Exec(query, code, longURL)
	if err != nil {
		panic(err)
	}
}

func (r *SQLiteURLRepository) Get(code string) (string, bool) {
	query := `SELECT long_url FROM urls WHERE code = ?`

	var longURL string
	err := r.db.QueryRow(query, code).Scan(&longURL)

	if err == sql.ErrNoRows {
		return "", false
	}

	if err != nil {
		panic(err)
	}

	return longURL, true
}