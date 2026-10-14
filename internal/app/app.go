package app

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sbrsn97/kisalt/internal/links"
)

func New(db *pgxpool.Pool) *http.ServeMux {
	mux := http.NewServeMux()

	linkStore := links.NewStore(db)
	linkService := links.NewService(linkStore)
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
