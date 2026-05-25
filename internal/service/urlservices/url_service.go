package urlservices

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"go.uber.org/zap"
)

type Service struct {
	repository *urlrepository.Repository
	logger     *zap.SugaredLogger
}

func NewService(repository *urlrepository.Repository, logger *zap.SugaredLogger) *Service {
	return &Service{
		repository: repository,
		logger:     logger.With(zap.String("component", "service")),
	}
}

func (service *Service) SetUrl(short string, original string) error {
	service.logger.Infow("Create new url", "short", short, "original", original)

	creator, err := service.repository.NewUrlCreator()
	if err != nil {
		service.logger.Warn("Error by create creator", err)
		return err
	}
	defer creator.Close()

	url := urlrepository.Url{
		Uuid:        service.repository.GetLastId() + 1,
		ShortUrl:    short,
		OriginalUrl: original,
	}

	if err := creator.Add(&url); err != nil {
		return fmt.Errorf("add url to storage: %w", err)
	}
	service.logger.Infow("New url", "Original", url.OriginalUrl, "Short", url.ShortUrl)
	return nil
}

func (service *Service) GetFullUrl(short string) (string, bool) {
	ok := false

	url := service.repository.GetOriginalByShort(short)
	if url != "" {
		service.logger.Infow("URL is finded")
		ok = true
	}

	return url, ok
}

func (service *Service) GenerateShort() (string, error) {
	const maxAttempts = 10
	for range maxAttempts {
		id := service.generateRandom()
		if _, ok := service.GetFullUrl(id); !ok {
			service.logger.Infow("Success generate short", "Id", id)
			return id, nil
		}
	}
	return "", errors.New("failed to generate unique short URL")
}

func (service *Service) generateRandom() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytesId = make([]byte, 8)
	for i := range bytesId {
		bytesId[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytesId)
}
