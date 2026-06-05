package urlhandlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/model"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

type Handler struct {
	service *urlservices.Service
	logger  *zap.SugaredLogger
}

func NewHandler(service *urlservices.Service, logger *zap.SugaredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.With(zap.String("component", "handler")),
	}
}

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

	createdUrl, err := handler.service.SetUrl(string(body))
	if err != nil {
		handler.logger.Debug("Err by add url", zap.Error(err))
		if errors.Is(err, urlrepository.ErrShortURLAlreadyExists) {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Options.AddrResp + "/" + createdUrl.ShortUrl))
}

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

	createdUrl, err := handler.service.SetUrl(request.Url)
	if err != nil {
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
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(response); err != nil {
		handler.logger.Warn("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

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

	createdBatchUrls, err := handler.service.SetBatch(&request)
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
			ShortUrl:      batchUrl.Url.ShortUrl,
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

func (handler *Handler) RedirectFullUrl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	fullUrl, ok := handler.service.GetFullUrl(id)
	if !ok {
		http.Error(w, "URL is not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, fullUrl, http.StatusTemporaryRedirect)
}

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
