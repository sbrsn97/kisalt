package main

import (
	"log"
	"net/http"

	"github.com/sbrsn97/kisalt/internal/app"
)

func main() {
	handler := app.New()

	log.Println("server listening on :8080")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
