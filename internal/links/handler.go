package links

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
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
	var req createLinkRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	link, err := h.service.Create(req.URL)
	if err != nil {
		if errors.Is(err, ErrInvalidURL) {
			http.Error(w, "invalid URL", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	resp := createLinkResponse{
		Slug: link.Slug,
		URL:  link.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", "/"+link.Slug)
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		return
	}
}
