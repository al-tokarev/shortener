package router

import (
	"net/http"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/go-chi/chi"
)

func GoRouter() error {
	r := chi.NewRouter()

	r.Post("/", urlhandlers.GetShortenedUrl)
	r.Get("/{id}", urlhandlers.RedirectFullUrl)

	return http.ListenAndServe(config.Options.AddrServe, r)
}
