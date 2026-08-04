package repository

import "sync"

type URLRepository interface {
	Save(code string, longURL string)
	Get(code string) (string, bool)
}

type MemoryURLRepository struct {
	mu   sync.RWMutex
	urls map[string]string
}

func NewMemoryURLRepository() *MemoryURLRepository {
	return &MemoryURLRepository{
		urls: make(map[string]string),
	}
}

func (r *MemoryURLRepository) Save(code string, longURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.urls[code] = longURL
}

func (r *MemoryURLRepository) Get(code string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	longURL, ok := r.urls[code]
	return longURL, ok
}