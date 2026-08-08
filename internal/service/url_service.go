package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"url-shortener/internal/model"
	"url-shortener/internal/repository"
)

var ErrURLNotFound = errors.New("url not found")
var ErrInvalidURL = errors.New("invalid url")
var ErrURLExpired = errors.New("url expired")

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) Shorten(longURL string, customCode string, expiresAt string) (string, error) {
	if longURL == "" {
		return "", ErrInvalidURL
	}

	if expiresAt != "" {
		_, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return "", errors.New("invalid expiry format, use RFC3339 like 2026-12-31T23:59:59Z")
		}
	}

	var code string

	if customCode != "" {
		code = customCode

		_, _, exists := s.repo.Get(code)
		if exists {
			return "", errors.New("custom code already exists")
		}
	} else {
		code = generateCode()
	}

	s.repo.Save(code, longURL, expiresAt)

	return code, nil
}

func (s *URLService) Resolve(code string) (string, error) {
	longURL, expiresAt, ok := s.repo.Get(code)
	if !ok {
		return "", ErrURLNotFound
	}

	if expiresAt != "" {
		expiryTime, err := time.Parse(time.RFC3339, expiresAt)
		if err != nil {
			return "", errors.New("invalid expiry saved in database")
		}

		if time.Now().After(expiryTime) {
			return "", ErrURLExpired
		}
	}

	return longURL, nil
}

func (s *URLService) TrackClick(click model.Click) {
	s.repo.SaveClick(click)
}

func (s *URLService) Analytics(code string) model.Analytics {
	return s.repo.GetAnalytics(code)
}

func generateCode() string {
	b := make([]byte, 6)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}