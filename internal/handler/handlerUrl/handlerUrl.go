package handlerurl

import (
	"io"
	"net/http"

	urlservice "github.com/al-tokarev/shortener/internal/service/urlService"
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

	if _, ok := urlservice.StorageURL[id]; !ok {
		http.Error(w, "URL is not found", 400)
		return
	}

	http.Redirect(w, r, urlservice.StorageURL[id], 307)
}
