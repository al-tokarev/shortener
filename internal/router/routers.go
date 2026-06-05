package router

import (
	"net/http"

	"github.com/al-tokarev/shortener/internal/compress"
	"github.com/al-tokarev/shortener/internal/handler/urlhandlers"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func NewRouter(handler *urlhandlers.Handler, log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()
	r.Use(logger.WithLogging(log))
	r.Use(compress.GzipMiddleware(log))

	r.Post("/", handler.GetShortenedUrl)
	r.Post("/api/shorten", handler.GetJsonShortenedUrl)
	r.Post("/api/shorten/batch", handler.GetJsonShortenedBatch)
	r.Get("/{id}", handler.RedirectFullUrl)
	r.Get("/ping", handler.PingHandler)

	return r
}
