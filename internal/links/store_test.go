package links

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestStoreCreateAndGetBySlug(t *testing.T) {
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

	slug := "test-integration"

	t.Cleanup(func() {
		_, _ = db.Exec(
			context.Background(),
			"DELETE FROM links WHERE slug = $1",
			slug,
		)
	})

	created, err := store.Create(
		ctx,
		slug,
		"https://example.com/integration-test",
	)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	got, err := store.GetBySlug(ctx, slug)
	if err != nil {
		t.Fatalf("get link: %v", err)
	}

	if got.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, got.ID)
	}

	if got.Slug != slug {
		t.Fatalf("expected slug %q, got %q", slug, got.Slug)
	}

	if got.OriginalURL != "https://example.com/integration-test" {
		t.Fatalf(
			"expected URL %q, got %q",
			"https://example.com/integration-test",
			got.OriginalURL,
		)
	}
}

func TestStoreGetBySlugNotFound(t *testing.T) {
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

	_, err = store.GetBySlug(ctx, "does-not-exist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
