// Пакет model содержит основные модели, в т.ч. модели запросов и ответов сервера
package model

// generate:reset
// URL представляет собой модель короткой ссылки.
// Содержит информацию об оригинальном URL, коротком идентификаторе и владельце.
type URL struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// generate:reset
type Example struct {
	x, y int
	g    bool
	mapa map[string]int
}

// generate:reset
// URLBatch представляет собой связку короткой ссылки с идентификатором корреляции.
// Используется при пакетном создании коротких ссылок.
type URLBatch struct {
	CorrelationID string
	URL           *URL
}

// Request представляет собой запрос на создание короткой ссылки.
type Request struct {
	URL string `json:"url"`
}

// Response представляет собой ответ с укороченной ссылкой.
type Response struct {
	Result string `json:"result"`
}

// RequestBatchURL представляет собой запрос на пакетное создание коротких ссылок.
type RequestBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseBatchURL представляет собой ответ с укороченной ссылкой для пакетного запроса.
type ResponseBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// ResponseUserURL представляет собой ответ со ссылками пользователя.
type ResponseUserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
