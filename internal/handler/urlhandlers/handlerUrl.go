package urlhandlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/model"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func GetShortenedUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
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

	urlservices.SetUrl("EwHXdJfB", string(body))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(config.Options.AddrResp + "/EwHXdJfB"))
}

func GetJsonShortenedUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if r.Header.Get("Content-Type") != "application/json" {
		logger.Sugar.Debug("Err content-type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Sugar.Info("Start decoding")
	var request model.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&request); err != nil {
		logger.Sugar.Debug("Err request decode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlservices.SetUrl("EwHXdJfB", request.Url)

	response := model.Response{
		Result: config.Options.AddrResp + "/EwHXdJfB",
	}

	enc := json.NewEncoder(w)
	w.WriteHeader(http.StatusCreated)
	if err := enc.Encode(response); err != nil {
		logger.Sugar.Debug("Err response encode", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func RedirectFullUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "id")

	fullUrl, ok := urlservices.GetFullUrl(id)
	if !ok {
		http.Error(w, "URL is not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, fullUrl, http.StatusTemporaryRedirect)
}
