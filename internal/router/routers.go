package router

import (
	"net/http"

	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
)

func GoRouter() error {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /", urlhandlers.GetShortenedUrl)
	mux.HandleFunc("GET /{id}", urlhandlers.RedirectFullUrl)

	return http.ListenAndServe(":8080", mux)
}
