package urlhandlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/require"
)

func TestGetShortenedUrl(t *testing.T) {
	type want struct {
		contentType string
		body        string
		statusCode  int
	}
	tests := []struct {
		name        string
		httpMethod  string
		contentType string
		body        string
		want        want
	}{
		{
			name:        "Correct request",
			httpMethod:  "POST",
			contentType: "text/plain",
			body:        "https://google.com",
			want: want{
				contentType: "text/plain",
				body:        "http://localhost:8080/EwHXdJfB",
				statusCode:  201,
			},
		},
		{
			name:        "Incorrect request",
			httpMethod:  "PUT",
			contentType: "text/plain",
			body:        "https://google.com",
			want: want{
				contentType: "text/plain",
				body:        "",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
		{
			name:        "Incorrect content-type",
			httpMethod:  "POST",
			contentType: "application/json",
			body:        "https://google.com",
			want: want{
				contentType: "text/plain",
				body:        "",
				statusCode:  400,
			},
		},
		{
			name:        "Empty body",
			httpMethod:  "POST",
			contentType: "text/plain",
			body:        "",
			want: want{
				contentType: "text/plain",
				body:        "Body is empty",
				statusCode:  400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody := strings.NewReader(test.body)

			request := httptest.NewRequest(test.httpMethod, "/", reqBody)
			request.Header.Set("Content-Type", test.contentType)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetShortenedUrl)
			h(w, request)

			result := w.Result()

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, test.want.body, string(body))
		})
	}
}

func TestRedirectFullUrl(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/{id}", RedirectFullUrl)

	type want struct {
		statusCode int
	}
	tests := []struct {
		name       string
		httpMethod string
		want       want
		id         string
	}{
		{
			name:       "Correct request",
			httpMethod: "GET",
			want: want{
				statusCode: 307,
			},
			id: "EwHXdJfB",
		},
		{
			name:       "Incorrect http method",
			httpMethod: "PUT",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			id: "EwHXdJfB",
		},
		{
			name:       "Not found id",
			httpMethod: "GET",
			want: want{
				statusCode: 400,
			},
			id: "abcdef",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			urlservices.StorageURL["EwHXdJfB"] = "http://example.com"

			request := httptest.NewRequest(test.httpMethod, "/"+test.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			require.Equal(t, test.want.statusCode, result.StatusCode)
		})
	}
}
