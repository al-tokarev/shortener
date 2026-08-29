// пакет router определяет список маршрутов приложения
package router

import (
	"net/http"

	"github.com/al-tokarev/shortener/internal/auth"
	"github.com/al-tokarev/shortener/internal/compress"
	"github.com/al-tokarev/shortener/internal/handler"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// NewRouter возвращает хендлер по требуему эндпоинту
func NewRouter(handler *handler.Handler, log *zap.SugaredLogger) http.Handler {
	r := chi.NewRouter()
	r.Use(logger.WithLogging(log))
	r.Use(compress.GzipMiddleware(log))
	r.Use(auth.AuthMiddleware)

	r.Post("/", handler.URLHandler.GetShortenedUrl)
	r.Post("/api/shorten", handler.URLHandler.GetJsonShortenedUrl)
	r.Post("/api/shorten/batch", handler.URLHandler.GetJsonShortenedBatch)
	r.Get("/api/user/urls", handler.URLHandler.GetUserURLs)
	r.Get("/{id}", handler.URLHandler.RedirectFullUrl)
	r.Get("/ping", handler.URLHandler.PingHandler)
	r.Delete("/api/user/urls", handler.URLHandler.DeleteUserURLs)

	return r
}
