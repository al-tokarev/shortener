package urlhandlers

import (
	"bytes"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/observer"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository/mocks"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/golang/mock/gomock"
)

func setupBenchmarkHandler(b *testing.B) (*Handler, func()) {
	testLogger, err := logger.NewLogger()
	if err != nil {
		b.Fatal(err)
	}

	config.Options.AddrResp = "http://localhost:8080"

	tmpFile, err := os.CreateTemp("", "test_storage_*.json")
	if err != nil {
		b.Fatal(err)
	}
	config.Options.StoragePath = tmpFile.Name()

	// Настраиваем моки
	ctrl := gomock.NewController(b)
	mockRepo := mocks.NewMockRepositoryInterface(ctrl)

	mockRepo.EXPECT().Save(gomock.Any()).Return(nil).AnyTimes()
	mockRepo.EXPECT().GetLastId().Return(0).AnyTimes()
	mockRepo.EXPECT().GetOriginalByShort(gomock.Any()).Return("http://yandex.ru", nil).AnyTimes()
	mockRepo.EXPECT().InitializeStorage().Return(nil).AnyTimes()
	mockRepo.EXPECT().Ping().Return(nil).AnyTimes()

	service := urlservices.NewService(mockRepo, testLogger)
	dispatcher := observer.NewDispatcher()
	handler := NewHandler(service, dispatcher, testLogger)

	cleanup := func() {
		ctrl.Finish()
		os.Remove(tmpFile.Name())
	}

	return handler, cleanup
}

func BenchmarkGetShortenedUrl(b *testing.B) {
	handler, cleanup := setupBenchmarkHandler(b)
	defer cleanup()

	body := []byte("https://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()

		handler.GetShortenedUrl(w, req)
	}
}

func BenchmarkGetJsonShortenedUrl(b *testing.B) {
	handler, cleanup := setupBenchmarkHandler(b)
	defer cleanup()

	request := `{"url": "https://example.com"}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader([]byte(request)))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.GetJsonShortenedUrl(w, req)
	}
}

func BenchmarkRedirectFullUrl(b *testing.B) {
	handler, cleanup := setupBenchmarkHandler(b)
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/abc123", nil)
		w := httptest.NewRecorder()

		handler.RedirectFullUrl(w, req)
	}
}
