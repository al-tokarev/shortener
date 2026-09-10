// Модуль urlhandlers содержит обработчики для работы с ссылками
package urlhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/al-tokarev/shortener/internal/auth"
	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/model"
	"github.com/al-tokarev/shortener/internal/observer"
	"github.com/al-tokarev/shortener/internal/observer/events"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Handler содержит зависимости для обработки HTTP запросов.
type URLHandler struct {
	service    urlservices.URLServiceInterface
	dispatcher observer.DispatcherInterface
	logger     *zap.SugaredLogger
}

// NewHandler создает новый экземпляр Handler.
// Принимает сервис для работы с URL, диспетчер событий и логгер.
func NewHandler(service urlservices.URLServiceInterface, dispatcher *observer.Dispatcher, logger *zap.SugaredLogger) *URLHandler {
	return &URLHandler{
		service:    service,
		dispatcher: dispatcher,
		logger:     logger.With(zap.String("component", "handler")),
	}
}

// GetShortenedURL обрабатывает POST запрос на создание короткой ссылки.
// Ожидает длинную ссылку в теле запроса с Content-Type: text/plain.
// Возвращает укороченную ссылку с кодом 201 Created.
// Возможные ошибки: 400 Bad Request, 401 Unauthorized, 409 Conflict.
func (h *URLHandler) GetShortenedURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if r.Header.Get("Content-Type") != "text/plain" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Body is empty"))
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createdURL, err := h.service.SetURL(string(body), userID)
	if err != nil {
		h.logger.Debug("Err by add url", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(config.Options.AddrResp + "/" + createdURL.ShortURL))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.dispatcher.Dispatch(events.AuditEvent{
		TS:     time.Now().Unix(),
		Action: events.Shorten,
		UserID: userID,
		URL:    createdURL.OriginalURL,
	})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Options.AddrResp + "/" + createdURL.ShortURL))
}

// GetJSONShortenedURL обрабатывает POST запрос на создание короткой ссылки.
// Ожидает JSON с полем "url" в теле запроса.
// Возвращает JSON с укороченной ссылкой и кодом 201 Created.
func (h *URLHandler) GetJSONShortenedURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Debug("Err content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Info("Start decoding")
	var request model.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		h.logger.Debug("Err request decode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createdURL, err := h.service.SetURL(request.URL, userID)
	if err != nil && !errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
		h.logger.Debug("Err by add url", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := model.Response{
		Result: config.Options.AddrResp + "/" + createdURL.ShortURL,
	}

	enc := json.NewEncoder(w)
	if errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
		h.dispatcher.Dispatch(events.AuditEvent{
			TS:     time.Now().Unix(),
			Action: events.Shorten,
			UserID: userID,
			URL:    createdURL.OriginalURL,
		})
	}

	if err := enc.Encode(response); err != nil {
		h.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// GetJSONShortenedBatch обрабатывает POST запрос на пакетное создание коротких ссылок.
// Ожидает массив JSON объектов с полями "correlation_id" и "original_url".
// Возвращает массив JSON объектов с укороченными ссылками.
func (h *URLHandler) GetJSONShortenedBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		h.logger.Debug("Err content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.logger.Info("Start decoding")
	request := []model.RequestBatchURL{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		h.logger.Debug("Err request decode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createdBatchURLs, err := h.service.SetBatch(&request, userID)
	if err != nil {
		h.logger.Debug("Err by add batch", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := []model.ResponseBatchURL{}
	for _, batchURL := range *createdBatchURLs {
		response = append(response, model.ResponseBatchURL{
			CorrelationID: batchURL.CorrelationID,
			ShortURL:      config.Options.AddrResp + "/" + batchURL.URL.ShortURL,
		})
	}

	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(response); err != nil {
		h.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// GetUserURLs возвращает все URL, созданные пользователем.
// Возвращает JSON массив с оригинальными и укороченными ссылками.
func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.service.GetUserURLs(userID)
	if err != nil {
		h.logger.Warn("Failed to get user URLs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(*urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := []model.ResponseUserURL{}
	for _, url := range *urls {
		response = append(response, model.ResponseUserURL{
			OriginalURL: url.OriginalURL,
			ShortURL:    config.Options.AddrResp + "/" + url.ShortURL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusOK)
	if err := enc.Encode(response); err != nil {
		h.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// DeleteUserURLs помечает URL как удаленные.
// Ожидает JSON массив коротких идентификаторов.
// Возвращает 202 Accepted при успешном выполнении.
func (h *URLHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		h.logger.Warn("Failed to decode delete request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(shortIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.service.DeleteUserURLs(shortIDs, userID)

	w.WriteHeader(http.StatusAccepted)
}

// RedirectFullURL обрабатывает GET запрос на переход по короткой ссылке.
// Извлекает короткий идентификатор из URL и редиректит на оригинальный URL.
func (h *URLHandler) RedirectFullURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fullURL, err := h.service.GetFullURL(id)
	if err != nil {
		if errors.Is(err, urlrepository.ErrURLNotFound) {
			http.Error(w, "URL is not found", http.StatusNotFound)
		} else if errors.Is(err, urlrepository.ErrURLDeleted) {
			http.Error(w, "URL is deleted", http.StatusGone)
		} else {
			http.Error(w, "Error by find URL", http.StatusBadRequest)
		}
		return
	}

	event := events.AuditEvent{
		TS:     time.Now().Unix(),
		Action: events.Follow,
		URL:    fullURL,
	}
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if ok {
		event.UserID = userID
	}
	h.dispatcher.Dispatch(event)

	http.Redirect(w, r, fullURL, http.StatusTemporaryRedirect)
}

// PingHandler проверяет доступность базы данных.
// Возвращает "Pong" и код 200 OK при успешной проверке.
func (h *URLHandler) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := h.service.PingDB()

	if err != nil {
		h.logger.Warn("Error sql connection", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.logger.Info("Pong. SQL connection is success")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
