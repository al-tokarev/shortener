package router

import (
	"net/http"
	"time"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/go-chi/chi"
)

func GoRouter() error {
	r := chi.NewRouter()

	r.Post("/", urlhandlers.GetShortenedUrl)
	r.Get("/{id}", urlhandlers.RedirectFullUrl)

	server := &http.Server{
		Addr:              config.Options.AddrServe,
		Handler:           r,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return server.ListenAndServe()
}
