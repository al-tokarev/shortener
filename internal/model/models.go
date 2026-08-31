// Пакет model содержит основные модели, в т.ч. модели запросов и ответов сервера
package model

// generate:reset
// Url представляет собой модель короткой ссылки.
// Содержит информацию об оригинальном URL, коротком идентификаторе и владельце.
type Url struct {
	Uuid        int    `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// generate:reset
// UrlBatch представляет собой связку короткой ссылки с идентификатором корреляции.
// Используется при пакетном создании коротких ссылок.
type UrlBatch struct {
	CorrelationId string
	Url           *Url
}

// Request представляет собой запрос на создание короткой ссылки.
type Request struct {
	Url string `json:"url"`
}

// Response представляет собой ответ с укороченной ссылкой.
type Response struct {
	Result string `json:"result"`
}

// RequestBatchUrl представляет собой запрос на пакетное создание коротких ссылок.
type RequestBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}

// ResponseBatchUrl представляет собой ответ с укороченной ссылкой для пакетного запроса.
type ResponseBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}

// ResponseUserUrl представляет собой ответ со ссылками пользователя.
type ResponseUserUrl struct {
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}
