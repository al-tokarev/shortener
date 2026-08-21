package urlhandlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Example_GetShortenedUrl демонстрирует запрос на создание короткой ссылки.
func ExampleHandler_GetShortenedUrl() {
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		strings.NewReader("https://example.com"),
	)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", "user123")

	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("Path: %s\n", req.URL.Path)
	fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))

	// Output:
	// Method: POST
	// Path: /
	// Content-Type: text/plain
}

// Example_GetJsonShortenedUrl демонстрирует запрос на создание короткой ссылки через JSON.
func ExampleHandler_GetJsonShortenedUrl() {
	jsonBody := `{"url": "https://example.com/very/long/url"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/shorten",
		strings.NewReader(jsonBody),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "user123")

	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("Path: %s\n", req.URL.Path)
	fmt.Printf("Content-Type: %s\n", req.Header.Get("Content-Type"))
	fmt.Printf("Body: %s\n", jsonBody)

	// Output:
	// Method: POST
	// Path: /api/shorten
	// Content-Type: application/json
	// Body: {"url": "https://example.com"}
}

// Example_RedirectFullUrl демонстрирует переход по короткой ссылке.
func ExampleHandler_RedirectFullUrl() {
	req := httptest.NewRequest(
		http.MethodGet,
		"/abc12345",
		nil,
	)

	fmt.Printf("Method: %s\n", req.Method)
	fmt.Printf("Path: %s\n", req.URL.Path)
	fmt.Printf("Short ID: %s\n", "abc12345")

	// Output:
	// Method: GET
	// Path: /abc12345
	// Short ID: abc12345
}
