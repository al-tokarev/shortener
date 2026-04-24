package router

import (
	"net/http"

	"github.com/al-tokarev/shortener.git/internal/handler/handlerurl"
)

func GoRouter() error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", handlerurl.GetShortenedUrl)
	mux.HandleFunc("GET /{id}", handlerurl.GetFullUrl)

	return http.ListenAndServe(":8080", mux)
}
