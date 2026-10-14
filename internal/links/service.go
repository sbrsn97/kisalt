package links

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
)

const (
	slugLength = 7
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrInvalidURL = errors.New("invalid URL")
	ErrNotFound   = errors.New("link not found")
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{
		store: store,
	}
}

func (s *Service) Create(ctx context.Context, rawURL string) (Link, error) {
	if !isValidURL(rawURL) {
		return Link{}, ErrInvalidURL
	}

	slug, err := generateSlug(slugLength)
	if err != nil {
		return Link{}, err
	}

	link, err := s.store.Create(ctx, slug, rawURL)
	if err != nil {
		return Link{}, err
	}

	return link, nil
}

func (s *Service) Get(ctx context.Context, slug string) (Link, error) {
	return s.store.GetBySlug(ctx, slug)
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
