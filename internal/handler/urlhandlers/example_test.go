package urlhandlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/al-tokarev/shortener/internal/auth"
	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/observer"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	urlservices "github.com/al-tokarev/shortener/internal/service/urlservices"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func setupExampleHandler() *URLHandler {
	testLogger := zap.NewNop().Sugar()
	config.Options.AddrResp = "http://localhost:8080"

	repo := urlrepository.NewLocalRepository(testLogger)
	repo.InitializeStorage()
	service := urlservices.NewService(repo, testLogger)
	dispatcher := observer.NewDispatcher(testLogger)

	return NewHandler(service, dispatcher, testLogger)
}

// withChiParam добавляет параметр маршрута chi в контекст.
func withChiParam(req *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx)
	return req.WithContext(ctx)
}

// ExampleURLHandler_GetShortenedUrl демонстрирует создание короткой ссылки.
func ExampleURLHandler_GetShortenedUrl() {
	handler := setupExampleHandler()

	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://example.com/very/long/url"),
	)
	req.Header.Set("Content-Type", "text/plain")
	req = req.WithContext(auth.SetUserIDToContext(req.Context(), "user123"))

	w := httptest.NewRecorder()
	handler.GetShortenedUrl(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	fmt.Println("HasPrefix:", strings.HasPrefix(w.Body.String(), "http://localhost:8080/"))

	// Output:
	// Status: 201
	// Content-Type: text/plain
	// HasPrefix: true
}

// ExampleURLHandler_GetJsonShortenedUrl демонстрирует создание короткой ссылки через JSON.
func ExampleURLHandler_GetJsonShortenedUrl() {
	handler := setupExampleHandler()

	jsonBody := `{"url": "https://example.com"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(auth.SetUserIDToContext(req.Context(), "user456"))

	w := httptest.NewRecorder()
	handler.GetJsonShortenedUrl(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Content-Type:", w.Header().Get("Content-Type"))
	fmt.Println("HasResult:", strings.Contains(w.Body.String(), "result"))

	// Output:
	// Status: 201
	// Content-Type: application/json
	// HasResult: true
}

// ExampleURLHandler_RedirectFullUrl демонстрирует переход по короткой ссылке.
func ExampleURLHandler_RedirectFullUrl() {
	handler := setupExampleHandler()

	// Создаем короткую ссылку
	createReq := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://example.com/redirect/target"),
	)
	createReq.Header.Set("Content-Type", "text/plain")
	createReq = createReq.WithContext(auth.SetUserIDToContext(createReq.Context(), "user789"))

	createW := httptest.NewRecorder()
	handler.GetShortenedUrl(createW, createReq)

	shortURL := strings.TrimPrefix(createW.Body.String(), config.Options.AddrResp+"/")

	req := httptest.NewRequest(http.MethodGet, "/"+shortURL, nil)
	req = withChiParam(req, "id", shortURL)

	w := httptest.NewRecorder()
	handler.RedirectFullUrl(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Location:", w.Header().Get("Location"))

	// Output:
	// Status: 307
	// Location: https://example.com/redirect/target
}
