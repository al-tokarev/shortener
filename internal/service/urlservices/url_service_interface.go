package urlservices

import "github.com/al-tokarev/shortener/internal/model"

type URLServiceInterface interface {
	SetURL(original string, userID string) (*model.URL, error)
	GetUserURLs(userID string) (*[]model.URL, error)
	SetBatch(batchURLs *[]model.RequestBatchURL, userID string) (*[]model.URLBatch, error)
	GetFullURL(short string) (string, error)
	GenerateShort() string
	DeleteUserURLs(shortIDs []string, userID string)
	PingDB() error
}
