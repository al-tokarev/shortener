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
	"github.com/al-tokarev/shortener/internal/repository/urlrepository/mocks"
	"github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*httptest.Server, *mocks.MockRepositoryInterface, func()) {
	testLogger, err := logger.NewLogger()
	require.NoError(t, err)

	config.Options.AddrResp = "http://localhost:8080"

	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	require.NoError(t, err)
	config.Options.StoragePath = tmpFile.Name()

	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockRepositoryInterface(ctrl)

	mockRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetLastId().Return(0).AnyTimes()
	mockRepo.EXPECT().GetOriginalByShort(gomock.Any()).Return("http://yandex.ru").AnyTimes()
	mockRepo.EXPECT().InitializeStorage().Return(nil).AnyTimes()
	mockRepo.EXPECT().Ping().Return(nil).AnyTimes()

	service := urlservices.NewService(mockRepo, testLogger)

	handler := NewHandler(service, testLogger)

	r := chi.NewRouter()
	r.Use(compress.GzipMiddleware(testLogger))
	r.Post("/", handler.GetShortenedUrl)
	r.Post("/api/shorten", handler.GetJsonShortenedUrl)
	r.Get("/{id}", handler.RedirectFullUrl)

	srv := httptest.NewServer(r)

	cleanup := func() {
		srv.Close()
		ctrl.Finish()
		os.Remove(tmpFile.Name())
	}

	return srv, mockRepo, cleanup
}

func TestShortenUrl(t *testing.T) {
	srv, _, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("shorten_text", func(t *testing.T) {
		requestBody := "https://google.com"

		req, err := http.NewRequest("POST", srv.URL+"/", bytes.NewBufferString(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		shortURL := strings.TrimSpace(string(body))
		assert.Contains(t, shortURL, config.Options.AddrResp+"/")

		id := strings.TrimPrefix(shortURL, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)
	})

	t.Run("shorten_json", func(t *testing.T) {
		type shortenRequest struct {
			URL string `json:"url"`
		}
		type shortenResponse struct {
			Result string `json:"result"`
		}

		reqBody := shortenRequest{URL: "https://practicum.yandex.ru"}
		jsonBody, err := json.Marshal(reqBody)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", srv.URL+"/api/shorten", bytes.NewBuffer(jsonBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response shortenResponse
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response.Result, config.Options.AddrResp+"/")

		id := strings.TrimPrefix(response.Result, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)
	})

	t.Run("incorrect_method", func(t *testing.T) {
		req, err := http.NewRequest("PUT", srv.URL+"/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	t.Run("incorrect_content_type", func(t *testing.T) {
		req, err := http.NewRequest("POST", srv.URL+"/", bytes.NewBufferString("https://google.com"))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty_body", func(t *testing.T) {
		req, err := http.NewRequest("POST", srv.URL+"/", nil)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestGzipCompression(t *testing.T) {
	srv, _, cleanup := setupTestServer(t)
	defer cleanup()

	requestBody := `{"url":"https://practicum.yandex.ru"}`

	t.Run("sends_gzip", func(t *testing.T) {
		// Сжимаем тело запроса
		buf := bytes.NewBuffer(nil)
		zw := gzip.NewWriter(buf)
		_, err := zw.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zw.Close()
		require.NoError(t, err)

		req, err := http.NewRequest("POST", srv.URL+"/api/shorten", buf)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var response model.Response
		err = json.Unmarshal(body, &response)
		require.NoError(t, err)

		assert.Contains(t, response.Result, config.Options.AddrResp+"/")

		id := strings.TrimPrefix(response.Result, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		req, err := http.NewRequest("POST", srv.URL+"/api/shorten", bytes.NewBufferString(requestBody))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)
		require.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)
		defer zr.Close()

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		var response model.Response
		err = json.Unmarshal(b, &response)
		require.NoError(t, err)

		assert.Contains(t, response.Result, config.Options.AddrResp+"/")

		id := strings.TrimPrefix(response.Result, config.Options.AddrResp+"/")
		require.NotEmpty(t, id)
		require.Len(t, id, 8)
	})
}

func TestRedirectUrl(t *testing.T) {
	srv, _, cleanup := setupTestServer(t)
	defer cleanup()

	originalURL := "http://yandex.ru"

	type shortenRequest struct {
		URL string `json:"url"`
	}
	type shortenResponse struct {
		Result string `json:"result"`
	}

	reqBody := shortenRequest{URL: originalURL}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", srv.URL+"/api/shorten", bytes.NewBuffer(jsonBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var response shortenResponse
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)
	resp.Body.Close()

	shortURL := response.Result
	shortID := strings.TrimPrefix(shortURL, config.Options.AddrResp+"/")
	require.NotEmpty(t, shortID)

	t.Run("expand_success", func(t *testing.T) {
		req, err := http.NewRequest("GET", srv.URL+"/"+shortID, nil)
		require.NoError(t, err)

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
		require.Equal(t, originalURL, resp.Header.Get("Location"))
	})

	t.Run("incorrect_method", func(t *testing.T) {
		req, err := http.NewRequest("PUT", srv.URL+"/"+shortID, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})

	// t.Run("not_found", func(t *testing.T) {
	// 	mockRepo.EXPECT().GetOriginalByShort(gomock.Any()).Return("").AnyTimes()

	// 	req, err := http.NewRequest("GET", srv.URL+"/nonexistent", nil)
	// 	require.NoError(t, err)

	// 	resp, err := http.DefaultClient.Do(req)
	// 	require.NoError(t, err)
	// 	defer resp.Body.Close()

	// 	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	// })
}
