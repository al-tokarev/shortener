package urlservices

import (
	"errors"
	"math/rand"

	"github.com/al-tokarev/shortener/internal/model"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"go.uber.org/zap"
)

type Service struct {
	repository urlrepository.RepositoryInterface
	logger     *zap.SugaredLogger
}

func NewService(repository urlrepository.RepositoryInterface, logger *zap.SugaredLogger) *Service {
	return &Service{
		repository: repository,
		logger:     logger.With(zap.String("component", "service")),
	}
}

func (service *Service) SetUrl(original string, userID string) (*model.Url, error) {
	const maxAttempts = 10
	currentAttempt := 1
	var url model.Url
	var errorCreate error
	for currentAttempt < maxAttempts {
		url = model.Url{
			Uuid:        service.repository.GetLastId() + 1,
			ShortUrl:    service.GenerateShort(),
			OriginalUrl: original,
			UserID:      userID,
		}
		errorCreate = service.repository.Save(&url)
		if errors.Is(errorCreate, urlrepository.ErrShortURLAlreadyExists) {
			service.logger.Debugw("Duplicate url by create", "Attempt", currentAttempt)
			currentAttempt++
			continue
		}
		if errors.Is(errorCreate, urlrepository.ErrOriginalURLAlreadyExists) {
			service.logger.Info("Duplicate original url by create", "Attempt")
			url, err := service.repository.GetByOriginal(original)
			if err != nil {
				return nil, err
			}
			return url, errorCreate
		}
		break
	}
	if errorCreate != nil {
		return nil, errorCreate
	}

	service.logger.Infow("New url", "Original", url.OriginalUrl, "Short", url.ShortUrl)
	return &url, nil
}

func (service *Service) GetUserURLs(userID string) (*[]model.Url, error) {
	service.logger.Infow("Getting user URLs", "user_id", userID)
	return service.repository.GetUserURLs(userID)
}

func (service *Service) SetBatch(batchUrls *[]model.RequestBatchUrl, userID string) (*[]model.UrlBatch, error) {
	urls := []model.Url{}
	urlsBatch := []model.UrlBatch{}

	lastId := service.repository.GetLastId() + 1
	for _, bUrl := range *batchUrls {
		url := model.Url{
			Uuid:        lastId,
			ShortUrl:    service.GenerateShort(),
			OriginalUrl: bUrl.OriginalUrl,
			UserID:      userID,
		}
		urls = append(urls, url)
		urlsBatch = append(urlsBatch, model.UrlBatch{
			CorrelationId: bUrl.CorrelationId,
			Url:           &url,
		})

		lastId++
	}

	errorCreate := service.repository.SaveBatch(&urls)
	if errorCreate != nil {
		return nil, errorCreate
	}

	return &urlsBatch, nil
}

func (service *Service) GetFullUrl(short string) (string, error) {
	url, err := service.repository.GetOriginalByShort(short)
	if err != nil {
		return "", err
	}
	service.logger.Infow("URL is finded")
	return url, nil
}

func (service *Service) GenerateShort() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytesId = make([]byte, 8)
	for i := range bytesId {
		bytesId[i] = letters[rand.Intn(len(letters))]
	}
	return string(bytesId)
}

func (service *Service) PingDb() error {
	return service.repository.Ping()
}
