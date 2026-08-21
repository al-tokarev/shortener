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
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

// Handler содержит зависимости для обработки HTTP запросов.
type Handler struct {
	service    *urlservices.Service
	dispatcher *observer.Dispatcher
	logger     *zap.SugaredLogger
}

// NewHandler создает новый экземпляр Handler.
// Принимает сервис для работы с URL, диспетчер событий и логгер.
func NewHandler(service *urlservices.Service, dispatcher *observer.Dispatcher, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service:    service,
		dispatcher: dispatcher,
		logger:     logger.With(zap.String("component", "handler")),
	}
}

// GetShortenedUrl обрабатывает POST запрос на создание короткой ссылки.
// Ожидает длинную ссылку в теле запроса с Content-Type: text/plain.
// Возвращает укороченную ссылку с кодом 201 Created.
// Возможные ошибки: 400 Bad Request, 401 Unauthorized, 409 Conflict.
func (handler *Handler) GetShortenedUrl(w http.ResponseWriter, r *http.Request) {
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

	createdUrl, err := handler.service.SetUrl(string(body), userID)
	if err != nil {
		handler.logger.Debug("Err by add url", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(config.Options.AddrResp + "/" + createdUrl.ShortUrl))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.dispatcher.Dispatch(events.AuditEvent{
		Ts:     time.Now().Unix(),
		Action: events.Shorten,
		UserID: userID,
		URL:    createdUrl.OriginalUrl,
	})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Options.AddrResp + "/" + createdUrl.ShortUrl))
}

// GetJsonShortenedUrl обрабатывает POST запрос на создание короткой ссылки.
// Ожидает JSON с полем "url" в теле запроса.
// Возвращает JSON с укороченной ссылкой и кодом 201 Created.
func (handler *Handler) GetJsonShortenedUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		handler.logger.Debug("Err content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.logger.Info("Start decoding")
	var request model.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		handler.logger.Debug("Err request decode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createdUrl, err := handler.service.SetUrl(request.Url, userID)
	if err != nil && !errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
		handler.logger.Debug("Err by add url", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := model.Response{
		Result: config.Options.AddrResp + "/" + createdUrl.ShortUrl,
	}

	enc := json.NewEncoder(w)
	if errors.Is(err, urlrepository.ErrOriginalURLAlreadyExists) {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
		handler.dispatcher.Dispatch(events.AuditEvent{
			Ts:     time.Now().Unix(),
			Action: events.Shorten,
			UserID: userID,
			URL:    createdUrl.OriginalUrl,
		})
	}

	if err := enc.Encode(response); err != nil {
		handler.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// GetJsonShortenedBatch обрабатывает POST запрос на пакетное создание коротких ссылок.
// Ожидает массив JSON объектов с полями "correlation_id" и "original_url".
// Возвращает массив JSON объектов с укороченными ссылками.
func (handler *Handler) GetJsonShortenedBatch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		handler.logger.Debug("Err content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.logger.Info("Start decoding")
	request := []model.RequestBatchUrl{}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		handler.logger.Debug("Err request decode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	createdBatchUrls, err := handler.service.SetBatch(&request, userID)
	if err != nil {
		handler.logger.Debug("Err by add batch", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response := []model.ResponseBatchUrl{}
	for _, batchUrl := range *createdBatchUrls {
		response = append(response, model.ResponseBatchUrl{
			CorrelationId: batchUrl.CorrelationId,
			ShortUrl:      config.Options.AddrResp + "/" + batchUrl.Url.ShortUrl,
		})
	}

	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(response); err != nil {
		handler.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// GetUserURLs возвращает все URL, созданные пользователем.
// Возвращает JSON массив с оригинальными и укороченными ссылками.
func (handler *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := handler.service.GetUserURLs(userID)
	if err != nil {
		handler.logger.Warn("Failed to get user URLs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if len(*urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := []model.ResponseUserUrl{}
	for _, url := range *urls {
		response = append(response, model.ResponseUserUrl{
			OriginalUrl: url.OriginalUrl,
			ShortUrl:    config.Options.AddrResp + "/" + url.ShortUrl,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusOK)
	if err := enc.Encode(response); err != nil {
		handler.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// DeleteUserURLs помечает URL как удаленные.
// Ожидает JSON массив коротких идентификаторов.
// Возвращает 202 Accepted при успешном выполнении.
func (handler *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
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
		handler.logger.Warn("Failed to decode delete request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(shortIDs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	handler.service.DeleteUserURLs(shortIDs, userID)

	w.WriteHeader(http.StatusAccepted)
}

// RedirectFullUrl обрабатывает GET запрос на переход по короткой ссылке.
// Извлекает короткий идентификатор из URL и редиректит на оригинальный URL.
func (handler *Handler) RedirectFullUrl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fullUrl, err := handler.service.GetFullUrl(id)
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
		Ts:     time.Now().Unix(),
		Action: events.Follow,
		URL:    fullUrl,
	}
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if ok {
		event.UserID = userID
	}
	handler.dispatcher.Dispatch(event)

	http.Redirect(w, r, fullUrl, http.StatusTemporaryRedirect)
}

// PingHandler проверяет доступность базы данных.
// Возвращает "Pong" и код 200 OK при успешной проверке.
func (handler *Handler) PingHandler(w http.ResponseWriter, r *http.Request) {
	err := handler.service.PingDb()

	if err != nil {
		handler.logger.Warn("Error sql connection", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	handler.logger.Info("Pong. SQL connection is success")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
