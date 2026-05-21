package router

import (
	"net/http"
	"time"

	"github.com/al-tokarev/shortener/internal/compress"
	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/go-chi/chi"
)

func GoRouter() error {
	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(compress.GzipMiddleware)

	r.Post("/", urlhandlers.GetShortenedUrl)
	r.Post("/api/shorten", urlhandlers.GetJsonShortenedUrl)
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

	logger.Sugar.Infow("Server is starting", "addr", server.Addr)
	return server.ListenAndServe()
}
