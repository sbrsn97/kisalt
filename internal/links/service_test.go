package links

import "testing"

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
