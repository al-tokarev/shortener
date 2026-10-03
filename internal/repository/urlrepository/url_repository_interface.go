package urlrepository

import (
	"context"

	"github.com/al-tokarev/shortener/internal/model"
)

//go:generate mockgen -destination=mocks/mock_repository.go -package=mocks github.com/al-tokarev/shortener/internal/repository/urlrepository RepositoryInterface

type RepositoryInterface interface {
	InitializeStorage() error
	Save(url *model.URL) error
	SaveBatch(urls *[]model.URL) error
	BatchDelete(shortIDs []string, userID string) error
	GetOriginalByShort(short string) (string, error)
	GetByOriginal(original string) (*model.URL, error)
	GetUserURLs(userID string) (*[]model.URL, error)
	GetStat(ctx context.Context) (int, int, error)
	GetLastID() int
	Ping() error
}
