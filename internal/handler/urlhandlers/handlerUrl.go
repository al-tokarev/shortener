package urlhandlers

import (
	"io"
	"net/http"

	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
)

func GetShortenedUrl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if len(body) == 0 {
		http.Error(w, "Body is empty", 400)
		return
	}

	urlservices.StorageURL["EwHXdJfB"] = string(body)

	w.WriteHeader(201)
	w.Write([]byte("http://localhost:8080/EwHXdJfB"))
}

func RedirectFullUrl(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	fullUrl, ok := urlservices.StorageURL[id]
	if !ok {
		http.Error(w, "URL is not found", 400)
		return
	}

	http.Redirect(w, r, fullUrl, 307)
}
