package urlhandlers

import (
	"io"
	"net/http"

	"github.com/al-tokarev/shortener/internal/config"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
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
