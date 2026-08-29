package urlservices

import "github.com/al-tokarev/shortener/internal/model"

type URLServiceInterface interface {
	SetUrl(original string, userID string) (*model.Url, error)
	GetUserURLs(userID string) (*[]model.Url, error)
	SetBatch(batchUrls *[]model.RequestBatchUrl, userID string) (*[]model.UrlBatch, error)
	GetFullUrl(short string) (string, error)
	GenerateShort() string
	DeleteUserURLs(shortIDs []string, userID string)
	PingDb() error
}
