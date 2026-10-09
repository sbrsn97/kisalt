package links

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirect(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	link, err := service.Create("https://example.com/some/path")
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

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

	if location := rec.Header().Get("Location"); location != link.URL {
		t.Fatalf(
			"expected Location %q, got %q",
			link.URL,
			location,
		)
	}
}

func TestRedirectNotFound(t *testing.T) {
	service := NewService()
	handler := NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{slug}", handler.Redirect)

	req := httptest.NewRequest(
		http.MethodGet,
		"/doesnotexist",
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
