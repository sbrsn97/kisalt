package links

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) Create(ctx context.Context, slug, originalURL string) (Link, error) {
	const query = `
		INSERT INTO links (slug, original_url)
		VALUES ($1, $2)
		RETURNING id, slug, original_url, created_at
	`

	var link Link

	err := s.db.QueryRow(
		ctx,
		query,
		slug,
		originalURL,
	).Scan(
		&link.ID,
		&link.Slug,
		&link.OriginalURL,
		&link.CreatedAt,
	)

	if err != nil {
		return Link{}, err
	}

	return link, nil
}

func (s *Store) GetBySlug(ctx context.Context, slug string) (Link, error) {
	const query = `
		SELECT id, slug, original_url, created_at
		FROM links
		WHERE slug = $1
	`

	var link Link

	err := s.db.QueryRow(
		ctx,
		query,
		slug,
	).Scan(
		&link.ID,
		&link.Slug,
		&link.OriginalURL,
		&link.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Link{}, ErrNotFound
		}

		return Link{}, err
	}

	return link, nil
}
