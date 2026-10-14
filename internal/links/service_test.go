package links

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "http URL",
			url:  "http://example.com",
			want: true,
		},
		{
			name: "https URL",
			url:  "https://example.com",
			want: true,
		},
		{
			name: "missing scheme",
			url:  "example.com",
			want: false,
		},
		{
			name: "invalid URL",
			url:  "not-a-url",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidURL(tt.url)

			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestCreateRetriesAfterSlugConflict(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	conflictingSlug := "conflict"
	successfulSlug := "success"

	t.Cleanup(func() {
		_, _ = db.Exec(
			context.Background(),
			"DELETE FROM links WHERE slug = ANY($1)",
			[]string{conflictingSlug, successfulSlug},
		)
	})

	_, err = store.Create(
		ctx,
		conflictingSlug,
		"https://example.com/existing",
	)
	if err != nil {
		t.Fatalf("seed conflicting link: %v", err)
	}

	service := NewService(store)

	slugs := []string{
		conflictingSlug,
		successfulSlug,
	}

	call := 0

	service.generateSlug = func(length int) (string, error) {
		slug := slugs[call]
		call++
		return slug, nil
	}

	link, err := service.Create(
		ctx,
		"https://example.com/new",
	)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	if link.Slug != successfulSlug {
		t.Fatalf(
			"expected slug %q, got %q",
			successfulSlug,
			link.Slug,
		)
	}

	if call != 2 {
		t.Fatalf(
			"expected generator to be called 2 times, got %d",
			call,
		)
	}
}

func TestCreateStopsAfterSlugGenerationAttemptsExhausted(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set")
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create database pool: %v", err)
	}
	defer db.Close()

	store := NewStore(db)

	slug := "always-conflict"

	t.Cleanup(func() {
		_, _ = db.Exec(
			context.Background(),
			"DELETE FROM links WHERE slug = $1",
			slug,
		)
	})

	_, err = store.Create(
		ctx,
		slug,
		"https://example.com/existing",
	)
	if err != nil {
		t.Fatalf("seed conflicting link: %v", err)
	}

	service := NewService(store)

	calls := 0

	service.generateSlug = func(length int) (string, error) {
		calls++
		return slug, nil
	}

	_, err = service.Create(
		ctx,
		"https://example.com/new",
	)

	if !errors.Is(err, ErrSlugGenerationExhausted) {
		t.Fatalf(
			"expected ErrSlugGenerationExhausted, got %v",
			err,
		)
	}

	if calls != maxSlugGenerationAttempts {
		t.Fatalf(
			"expected %d attempts, got %d",
			maxSlugGenerationAttempts,
			calls,
		)
	}
}
