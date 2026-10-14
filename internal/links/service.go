package links

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"net/url"
)

const (
	slugLength                = 7
	maxSlugGenerationAttempts = 5
	alphabet                  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrInvalidURL              = errors.New("invalid URL")
	ErrNotFound                = errors.New("link not found")
	ErrSlugConflict            = errors.New("slug already exists")
	ErrSlugGenerationExhausted = errors.New("could not generate unique slug")
)

type Service struct {
	store        *Store
	generateSlug func(int) (string, error)
}

func NewService(store *Store) *Service {
	return &Service{
		store:        store,
		generateSlug: generateSlug,
	}
}

func (s *Service) Create(ctx context.Context, rawURL string) (Link, error) {
	if !isValidURL(rawURL) {
		return Link{}, ErrInvalidURL
	}

	for range maxSlugGenerationAttempts {
		slug, err := s.generateSlug(slugLength)
		if err != nil {
			return Link{}, err
		}

		link, err := s.store.Create(ctx, slug, rawURL)
		if err == nil {
			return link, nil
		}

		if errors.Is(err, ErrSlugConflict) {
			continue
		}

		return Link{}, err
	}

	return Link{}, ErrSlugGenerationExhausted
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
