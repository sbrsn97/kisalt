package app

import (
	"net/http"

	"github.com/sbrsn97/kisalt/internal/links"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	linkService := links.NewService()
	linkHandler := links.NewHandler(linkService)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write([]byte("OK")); err != nil {
			return
		}
	})

	mux.HandleFunc("POST /links", linkHandler.Create)
	mux.HandleFunc("GET /{slug}", linkHandler.Redirect)

	return mux
}
