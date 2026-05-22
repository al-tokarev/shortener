package urlhandlers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/al-tokarev/shortener/internal/compress"
	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/model"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShortenedUrl(t *testing.T) {
	logger.Initialize()
	if config.Options.AddrResp == "" {
		config.Options.AddrResp = "http://localhost:8080"
	}

	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

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
				statusCode:  http.StatusCreated,
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
				statusCode:  http.StatusBadRequest,
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
				statusCode:  http.StatusBadRequest,
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

			require.Equal(t, len(test.want.body), len(string(body)))
		})
	}
}

func TestGetJsonShortenedUrl(t *testing.T) {
	logger.Initialize()
	if config.Options.AddrResp == "" {
		config.Options.AddrResp = "http://localhost:8080"
	}
	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	config.Options.StoragePath = tmpFile.Name()

	type want struct {
		contentType string
		body        string
		statusCode  int
	}
	tests := []struct {
		name        string
		httpMethod  string
		contentType string
		body        interface{}
		want        want
	}{
		{
			name:        "Correct request",
			httpMethod:  "POST",
			contentType: "application/json",
			body: model.Request{
				Url: "https://practicum.yandex.ru",
			},
			want: want{
				contentType: "application/json",
				body:        `{"result":"http://localhost:8080/EwHXdJfB"}` + "\n",
				statusCode:  http.StatusCreated,
			},
		},
		{
			name:        "Incorrect http method",
			httpMethod:  "PUT",
			contentType: "application/json",
			body: model.Request{
				Url: "https://practicum.yandex.ru",
			},
			want: want{
				contentType: "application/json",
				body:        "",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
		{
			name:        "Incorrect content-type",
			httpMethod:  "POST",
			contentType: "text/plain",
			body: model.Request{
				Url: "https://practicum.yandex.ru",
			},
			want: want{
				contentType: "application/json",
				body:        "",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name:        "Empty body",
			httpMethod:  "POST",
			contentType: "application/json",
			body:        "",
			want: want{
				contentType: "application/json",
				body:        "",
				statusCode:  http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(test.body)
			require.NoError(t, err)

			request := httptest.NewRequest(test.httpMethod, "/api/shorten", bytes.NewBuffer(jsonBody))
			request.Header.Set("Content-Type", test.contentType)

			w := httptest.NewRecorder()
			h := http.HandlerFunc(GetJsonShortenedUrl)
			h(w, request)

			result := w.Result()
			defer result.Body.Close()

			require.Equal(t, test.want.statusCode, result.StatusCode)
			require.Equal(t, test.want.contentType, result.Header.Get("Content-Type"))

			body, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			require.Equal(t, len(test.want.body), len(string(body)))
		})
	}
}

func TestCompression(t *testing.T) {
	logger.Initialize()
	if config.Options.AddrResp == "" {
		config.Options.AddrResp = "http://localhost:8080"
	}

	requestBody := `{
        "url":"https://practicum.yandex.ru"
    }`

	// ожидаемое содержимое тела ответа при успешном запросе
	// successBody := `{
	// 	"result":"http://localhost:8080/EwHXdJfB"
	// }`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", buf)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "gzip")
		request.Header.Set("Accept-Encoding", "")

		w := httptest.NewRecorder()
		h := compress.GzipMiddleware(http.HandlerFunc(GetJsonShortenedUrl))
		h.ServeHTTP(w, request)

		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusCreated, result.StatusCode)

		body, err := io.ReadAll(result.Body)
		require.NoError(t, err)

		// Парсим ответ
		var response model.Response
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		// Проверяем структуру ответа
		assert.Contains(t, response.Result, config.Options.AddrResp+"/")

		// Извлекаем ID и проверяем, что URL сохранился
		id := strings.TrimPrefix(response.Result, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)

		savedURL, exists := urlservices.GetFullUrl(id)
		require.True(t, exists)
		require.Equal(t, "https://practicum.yandex.ru", savedURL)
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		reqBody := strings.NewReader(requestBody)

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", reqBody)
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Content-Encoding", "")
		request.Header.Set("Accept-Encoding", "gzip")

		w := httptest.NewRecorder()
		h := compress.GzipMiddleware(http.HandlerFunc(GetJsonShortenedUrl))
		h.ServeHTTP(w, request)

		result := w.Result()
		defer result.Body.Close()

		require.Equal(t, http.StatusCreated, result.StatusCode)

		gr, err := gzip.NewReader(result.Body)
		require.NoError(t, err)

		body, err := io.ReadAll(gr)
		require.NoError(t, err)

		// Парсим ответ
		var response model.Response
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		// Проверяем структуру ответа
		assert.Contains(t, response.Result, config.Options.AddrResp+"/")

		// Извлекаем ID и проверяем, что URL сохранился
		id := strings.TrimPrefix(response.Result, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)

		savedURL, exists := urlservices.GetFullUrl(id)
		require.True(t, exists)
		require.Equal(t, "https://practicum.yandex.ru", savedURL)
	})
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
				statusCode: http.StatusTemporaryRedirect,
			},
			id: urlservices.GenerateShort(),
		},
		{
			name:       "Incorrect http method",
			httpMethod: "PUT",
			want: want{
				statusCode: http.StatusMethodNotAllowed,
			},
			id: urlservices.GenerateShort(),
		},
		{
			name:       "Not found id",
			httpMethod: "GET",
			want: want{
				statusCode: http.StatusBadRequest,
			},
			id: "abc",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if len(test.id) == 8 {
				urlservices.SetUrl(test.id, "http://example.com")
			}

			request := httptest.NewRequest(test.httpMethod, "/"+test.id, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, request)

			result := w.Result()

			require.Equal(t, test.want.statusCode, result.StatusCode)
		})
	}
}
