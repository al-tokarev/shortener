package model

type Url struct {
	Uuid        int    `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

type UrlBatch struct {
	CorrelationId string
	Url           *Url
}

type Request struct {
	Url string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type RequestBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	OriginalUrl   string `json:"original_url"`
}

type ResponseBatchUrl struct {
	CorrelationId string `json:"correlation_id"`
	ShortUrl      string `json:"short_url"`
}
