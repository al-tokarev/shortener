package urlhandlers

import (
	"io"
	"net/http"

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
		w.WriteHeader(400)
		return
	}

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(body) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Body is empty"))
		return
	}

	urlservices.StorageURL["EwHXdJfB"] = string(body)

	w.WriteHeader(201)
	w.Write([]byte("http://localhost:8080/EwHXdJfB"))
}

func RedirectFullUrl(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	id := chi.URLParam(r, "id")

	fullUrl, ok := urlservices.StorageURL[id]
	if !ok {
		http.Error(w, "URL is not found", 400)
		return
	}

	http.Redirect(w, r, fullUrl, 307)
}
