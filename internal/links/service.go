package links

import (
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
	"sync"
)

const (
	slugLength = 7
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var ErrInvalidURL = errors.New("invalid URL")

type Service struct {
	mu    sync.RWMutex
	links map[string]Link
}

func NewService() *Service {
	return &Service{
		links: make(map[string]Link),
	}
}

func (s *Service) Create(rawURL string) (Link, error) {
	if !isValidURL(rawURL) {
		return Link{}, ErrInvalidURL
	}

	for {
		slug, err := generateSlug(slugLength)
		if err != nil {
			return Link{}, err
		}

		s.mu.Lock()

		if _, exists := s.links[slug]; exists {
			s.mu.Unlock()
			continue
		}

		link := Link{
			Slug: slug,
			URL:  rawURL,
		}

		s.links[slug] = link
		s.mu.Unlock()

		return link, nil
	}
}

func isValidURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}

	return parsed.Scheme == "http" || parsed.Scheme == "https"
}

func generateSlug(length int) (string, error) {
	result := make([]byte, length)

	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}

		result[i] = alphabet[n.Int64()]
	}

	return string(result), nil
}
