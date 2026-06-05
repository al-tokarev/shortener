package urlrepository

import "github.com/al-tokarev/shortener/internal/model"

type RepositoryInterface interface {
	InitializeStorage() error
	Save(url *model.Url) error
	SaveBatch(urls *[]model.Url) error
	GetOriginalByShort(short string) string
	GetByOriginal(original string) (*model.Url, error)
	GetLastId() int
	Ping() error
}
