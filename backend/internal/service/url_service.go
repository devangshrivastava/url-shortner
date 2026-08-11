package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrInvalidURL    = errors.New("invalid url")
	ErrInvalidExpiry = errors.New("invalid expiry format, use RFC3339 like 2026-12-31T23:59:59Z")
	ErrURLExpired    = errors.New("url expired")
	ErrURLCodeExists = errors.New("custom code already exists")
	ErrForbidden     = errors.New("forbidden")
)

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(
	ctx context.Context,
	userID int64,
	longURL string,
	customCode string,
	expiresAt string,
) (string, error) {
	if longURL == "" {
		return "", ErrInvalidURL
	}
	if err := validateExpiry(expiresAt); err != nil {
		return "", err
	}

	code := customCode
	if code == "" {
		var err error
		code, err = generateCode()
		if err != nil {
			return "", err
		}
	} else {
		_, err := s.repo.Get(ctx, code)
		if err == nil {
			return "", ErrURLCodeExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", err
		}
	}

	if err := s.repo.Save(ctx, userID, code, longURL, expiresAt); err != nil {
		if isURLUniqueViolation(err) {
			return "", ErrURLCodeExists
		}
		return "", err
	}

	return code, nil
}

func (s *URLService) Resolve(ctx context.Context, code string) (string, error) {
	url, err := s.repo.Get(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", err
	}

	if err := validateURLExpiry(url.ExpiresAt); err != nil {
		return "", err
	}

	return url.LongURL, nil
}

func (s *URLService) TrackClick(ctx context.Context, click model.Click) error {
	return s.repo.SaveClick(ctx, click)
}

func (s *URLService) GetAnalytics(
	ctx context.Context,
	userID int64,
	code string,
) (model.Analytics, error) {
	if _, err := s.getOwnedURL(ctx, userID, code); err != nil {
		return model.Analytics{}, err
	}

	return s.repo.GetAnalytics(ctx, code)
}

func (s *URLService) ListURLs(
	ctx context.Context,
	userID int64,
) ([]model.URL, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *URLService) UpdateURL(
	ctx context.Context,
	userID int64,
	code string,
	longURL *string,
	expiresAt *string,
) (model.URL, error) {
	if longURL == nil && expiresAt == nil {
		return model.URL{}, ErrInvalidURL
	}
	if longURL != nil && *longURL == "" {
		return model.URL{}, ErrInvalidURL
	}
	if expiresAt != nil {
		if err := validateExpiry(*expiresAt); err != nil {
			return model.URL{}, err
		}
	}

	if _, err := s.getOwnedURL(ctx, userID, code); err != nil {
		return model.URL{}, err
	}

	updated, err := s.repo.Update(ctx, code, longURL, expiresAt)
	if err != nil {
		return model.URL{}, err
	}
	if !updated {
		return model.URL{}, ErrURLNotFound
	}

	url, err := s.repo.Get(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, ErrURLNotFound
		}
		return model.URL{}, err
	}

	return url, nil
}

func (s *URLService) DeleteURL(
	ctx context.Context,
	userID int64,
	code string,
) error {
	if _, err := s.getOwnedURL(ctx, userID, code); err != nil {
		return err
	}

	deleted, err := s.repo.Delete(ctx, code)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrURLNotFound
	}

	return nil
}

func (s *URLService) getOwnedURL(
	ctx context.Context,
	userID int64,
	code string,
) (model.URL, error) {
	url, err := s.repo.Get(ctx, code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.URL{}, ErrURLNotFound
		}
		return model.URL{}, err
	}

	if url.UserID == nil || *url.UserID != userID {
		return model.URL{}, ErrForbidden
	}

	return url, nil
}

func validateExpiry(expiresAt string) error {
	if expiresAt == "" {
		return nil
	}

	if _, err := time.Parse(time.RFC3339, expiresAt); err != nil {
		return ErrInvalidExpiry
	}

	return nil
}

func validateURLExpiry(expiresAt string) error {
	if expiresAt == "" {
		return nil
	}

	expiryTime, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return ErrInvalidExpiry
	}
	if time.Now().After(expiryTime) {
		return ErrURLExpired
	}

	return nil
}

func generateCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8], nil
}

func isURLUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
