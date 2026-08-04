package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"url-shortener/internal/repository"
)

var ErrURLNotFound = errors.New("url not found")
var ErrInvalidURL = errors.New("invalid url")

type URLService struct {
	repo repository.URLRepository
}

func NewURLService(repo repository.URLRepository) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) Shorten(longURL string) (string, error) {
	if longURL == "" {
		return "", ErrInvalidURL
	}

	code := generateCode()

	s.repo.Save(code, longURL)

	return code, nil
}

func (s *URLService) Resolve(code string) (string, error) {
	longURL, ok := s.repo.Get(code)
	if !ok {
		return "", ErrURLNotFound
	}

	return longURL, nil
}

func generateCode() string {
	b := make([]byte, 6)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(b)[:8]
}