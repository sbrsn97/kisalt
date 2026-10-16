package links

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

type fakeService struct {
	createFunc func(ctx context.Context, rawURL string) (Link, error)
	getFunc    func(ctx context.Context, slug string) (Link, error)
}

func (f *fakeService) Create(
	ctx context.Context,
	rawURL string,
) (Link, error) {
	return f.createFunc(ctx, rawURL)
}

func (f *fakeService) Get(
	ctx context.Context,
	slug string,
) (Link, error) {
	return f.getFunc(ctx, slug)
}

func TestCreateLink(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			if rawURL != "https://example.com" {
				t.Fatalf(
					"expected URL %q, got %q",
					"https://example.com",
					rawURL,
				)
			}

			return Link{
				Slug:        "abc1234",
				OriginalURL: rawURL,
			}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"url":"https://example.com"}`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if location := rec.Header().Get("Location"); location != "/abc1234" {
		t.Fatalf(
			"expected Location %q, got %q",
			"/abc1234",
			location,
		)
	}
}

func TestCreateLinkRejectsMalformedJSON(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			t.Fatal("service should not be called")
			return Link{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"url":`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateLinkRejectsUnknownFields(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			t.Fatal("service should not be called")
			return Link{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(
			`{"url":"https://example.com","unexpected":"value"}`,
		),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateLinkRejectsMultipleJSONObjects(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			t.Fatal("service should not be called")
			return Link{}, nil
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(
			`{"url":"https://example.com"}{"url":"https://google.com"}`,
		),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateLinkRejectsOversizedBody(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			t.Fatal("service should not be called")
			return Link{}, nil
		},
	}

	handler := NewHandler(service)

	body := `{"url":"https://example.com/` +
		strings.Repeat("a", maxCreateLinkBodyBytes) +
		`"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateLinkRejectsInvalidURL(t *testing.T) {
	service := &fakeService{
		createFunc: func(
			ctx context.Context,
			rawURL string,
		) (Link, error) {
			return Link{}, ErrInvalidURL
		},
	}

	handler := NewHandler(service)

	req := httptest.NewRequest(
		http.MethodPost,
		"/links",
		strings.NewReader(`{"url":"not-a-url"}`),
	)

	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
