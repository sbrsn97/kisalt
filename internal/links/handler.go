package links

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxCreateLinkBodyBytes = 4 * 1024

type Handler struct {
	service linkService
}

type linkService interface {
	Create(ctx context.Context, rawURL string) (Link, error)
	Get(ctx context.Context, slug string) (Link, error)
}

func NewHandler(service linkService) *Handler {
	return &Handler{
		service: service,
	}
}

type createLinkRequest struct {
	URL string `json:"url"`
}

type createLinkResponse struct {
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxCreateLinkBodyBytes)

	var req createLinkRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(
			w,
			"request body must contain a single JSON object",
			http.StatusBadRequest,
		)
		return
	}

	link, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	response := createLinkResponse{
		Slug: link.Slug,
		URL:  link.OriginalURL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/"+link.Slug)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	link, err := h.service.Get(r.Context(), slug)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "link not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}
