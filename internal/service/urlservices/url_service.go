package urlservices

import (
	"errors"
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

func (service *Service) SetUrl(original string) (*urlrepository.Url, error) {
	// service.logger.Infow("Create new url", "short", short, "original", original)

	creator, err := service.repository.NewUrlCreator()
	if err != nil {
		service.logger.Warn("Error by create creator", err)
		return nil, err
	}
	defer creator.Close()

	const maxAttempts = 10
	currentAttempt := 1
	var url urlrepository.Url
	var errorCreate error
	for currentAttempt < maxAttempts {
		url = urlrepository.Url{
			Uuid:        service.repository.GetLastId() + 1,
			ShortUrl:    service.GenerateShort(),
			OriginalUrl: original,
		}
		errorCreate = creator.Add(&url)
		if errors.Is(errorCreate, urlrepository.ErrShortURLAlreadyExists) {
			service.logger.Debugw("Duplicate url by create", "Attempt", currentAttempt)
			currentAttempt++
			continue
		}
		break
	}
	if errorCreate != nil {
		return nil, errorCreate
	}

	service.logger.Infow("New url", "Original", url.OriginalUrl, "Short", url.ShortUrl)
	return &url, nil
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

func (service *Service) GenerateShort() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytesId = make([]byte, 8)
	for i := range bytesId {
		bytesId[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytesId)
}
