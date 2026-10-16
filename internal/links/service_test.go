package links

import (
	"context"
	"errors"
	"testing"
)

type fakeStore struct {
	createFunc    func(ctx context.Context, slug, originalURL string) (Link, error)
	getBySlugFunc func(ctx context.Context, slug string) (Link, error)
}

func (f *fakeStore) Create(
	ctx context.Context,
	slug string,
	originalURL string,
) (Link, error) {
	return f.createFunc(ctx, slug, originalURL)
}

func (f *fakeStore) GetBySlug(
	ctx context.Context,
	slug string,
) (Link, error) {
	return f.getBySlugFunc(ctx, slug)
}

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
	ctx := context.Background()

	slugs := []string{
		"conflict",
		"success",
	}

	generated := 0
	createCalls := 0

	store := &fakeStore{
		createFunc: func(
			ctx context.Context,
			slug string,
			originalURL string,
		) (Link, error) {
			createCalls++

			if slug == "conflict" {
				return Link{}, ErrSlugConflict
			}

			return Link{
				Slug:        slug,
				OriginalURL: originalURL,
			}, nil
		},
	}

	service := NewService(store)

	service.generateSlug = func(length int) (string, error) {
		slug := slugs[generated]
		generated++
		return slug, nil
	}

	link, err := service.Create(
		ctx,
		"https://example.com/new",
	)
	if err != nil {
		t.Fatalf("create link: %v", err)
	}

	if link.Slug != "success" {
		t.Fatalf(
			"expected slug %q, got %q",
			"success",
			link.Slug,
		)
	}

	if generated != 2 {
		t.Fatalf(
			"expected generator to be called 2 times, got %d",
			generated,
		)
	}

	if createCalls != 2 {
		t.Fatalf(
			"expected store Create to be called 2 times, got %d",
			createCalls,
		)
	}
}

func TestCreateStopsAfterSlugGenerationAttemptsExhausted(t *testing.T) {
	ctx := context.Background()

	store := &fakeStore{
		createFunc: func(
			ctx context.Context,
			slug string,
			originalURL string,
		) (Link, error) {
			return Link{}, ErrSlugConflict
		},
	}

	service := NewService(store)

	generatorCalls := 0

	service.generateSlug = func(length int) (string, error) {
		generatorCalls++
		return "conflict", nil
	}

	_, err := service.Create(
		ctx,
		"https://example.com/new",
	)

	if !errors.Is(err, ErrSlugGenerationExhausted) {
		t.Fatalf(
			"expected ErrSlugGenerationExhausted, got %v",
			err,
		)
	}

	if generatorCalls != maxSlugGenerationAttempts {
		t.Fatalf(
			"expected %d attempts, got %d",
			maxSlugGenerationAttempts,
			generatorCalls,
		)
	}
}

func TestCreateDoesNotRetryNonConflictError(t *testing.T) {
	ctx := context.Background()

	storeErr := errors.New("database unavailable")

	createCalls := 0

	store := &fakeStore{
		createFunc: func(
			ctx context.Context,
			slug string,
			originalURL string,
		) (Link, error) {
			createCalls++
			return Link{}, storeErr
		},
	}

	service := NewService(store)

	generatorCalls := 0

	service.generateSlug = func(length int) (string, error) {
		generatorCalls++
		return "abcdefg", nil
	}

	_, err := service.Create(
		ctx,
		"https://example.com",
	)

	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}

	if createCalls != 1 {
		t.Fatalf(
			"expected 1 store call, got %d",
			createCalls,
		)
	}

	if generatorCalls != 1 {
		t.Fatalf(
			"expected 1 generator call, got %d",
			generatorCalls,
		)
	}
}

func TestGetReturnsLinkFromStore(t *testing.T) {
	ctx := context.Background()

	expected := Link{
		ID:          42,
		Slug:        "abc1234",
		OriginalURL: "https://example.com",
	}

	store := &fakeStore{
		getBySlugFunc: func(
			ctx context.Context,
			slug string,
		) (Link, error) {
			if slug != expected.Slug {
				t.Fatalf(
					"expected slug %q, got %q",
					expected.Slug,
					slug,
				)
			}

			return expected, nil
		},
	}

	service := NewService(store)

	got, err := service.Get(ctx, expected.Slug)
	if err != nil {
		t.Fatalf("get link: %v", err)
	}

	if got != expected {
		t.Fatalf(
			"expected %+v, got %+v",
			expected,
			got,
		)
	}
}
