package links

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRedirect(t *testing.T) {
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
	service := NewService(store)
	handler := NewHandler(service)

	link, err := service.Create(
		ctx,
		"https://example.com/some/path",
	)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	t.Cleanup(func() {
		_, _ = db.Exec(
			context.Background(),
			"DELETE FROM links WHERE id = $1",
			link.ID,
		)
	})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{slug}", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/"+link.Slug,
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusFound,
			rec.Code,
		)
	}

	if location := rec.Header().Get("Location"); location != link.OriginalURL {
		t.Fatalf(
			"expected Location %q, got %q",
			link.OriginalURL,
			location,
		)
	}
}

func TestRedirectNotFound(t *testing.T) {
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
	service := NewService(store)
	handler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{slug}", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/does-not-exist",
		nil,
	)

	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
