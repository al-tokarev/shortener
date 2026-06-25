package urlrepository

import "github.com/al-tokarev/shortener/internal/model"

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/al-tokarev/shortener/internal/repository/urlrepository RepositoryInterface

type RepositoryInterface interface {
	InitializeStorage() error
	Save(url *model.Url) error
	SaveBatch(urls *[]model.Url) error
	BatchDelete(shortIDs []string, userID string) error
	GetOriginalByShort(short string) (string, error)
	GetByOriginal(original string) (*model.Url, error)
	GetUserURLs(userID string) (*[]model.Url, error)
	GetLastId() int
	Ping() error
}
