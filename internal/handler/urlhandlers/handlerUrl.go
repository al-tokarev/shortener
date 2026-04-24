package urlhandlers

import (
	"io"
	"net/http"

	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
)

func GetShortenedUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain")

	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Body is empty", 400)
		return
	}

	w.WriteHeader(201)
	w.Write([]byte("EwHXdJfB"))
}

func GetFullUrl(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if _, ok := urlservices.StorageURL[id]; !ok {
		http.Error(w, "URL is not found", 400)
		return
	}

	http.Redirect(w, r, urlservices.StorageURL[id], 307)
}
