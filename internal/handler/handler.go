package handler

import "github.com/al-tokarev/shortener/internal/handler/urlhandlers"

type Handler struct {
	URLHandler *urlhandlers.URLHandler
}

func NewHandler(URLHandler *urlhandlers.URLHandler) *Handler {
	return &Handler{
		URLHandler: URLHandler,
	}
}
