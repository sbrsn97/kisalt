package links

import (
	"errors"
	"testing"
)

func TestCreateLink(t *testing.T) {
	service := NewService()

	link, err := service.Create("https://example.com/some/long/path")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if link.URL != "https://example.com/some/long/path" {
		t.Errorf("expected URL %q, got %q", "https://example.com/some/long/path", link.URL)
	}

	if len(link.Slug) != slugLength {
		t.Errorf("expected slug length %d, got %d", slugLength, len(link.Slug))
	}
}

func TestCreateLinkRejectInvalidURL(t *testing.T) {
	service := NewService()

	_, err := service.Create("not-a-url")
	if err != ErrInvalidURL {
		t.Fatalf("expected ErrInvalidURL, got %v", err)
	}
}

func TestGetLink(t *testing.T) {
	service := NewService()

	created, err := service.Create("https://example.com/")
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	got, err := service.Get(created.Slug)
	if err != nil {
		t.Fatalf("get link: %v", err)
	}

	if got != created {
		t.Fatalf("expected link %+v, got %+v", created, got)
	}
}

func TestGetLinkNotFound(t *testing.T) {
	service := NewService()

	_, err := service.Get("doesnotexist")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
