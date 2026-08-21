// Пакет urlservices содержит бизнес-логику для работы с короткими ссылками.
// Обеспечивает взаимодействие между HTTP обработчиками и репозиториями данных.
package urlservices

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/al-tokarev/shortener/internal/model"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
	"go.uber.org/zap"
)

// deleteTask представляет задачу на удаление ссылок пользователя.
type deleteTask struct {
	shortIDs []string
	userID   string
}

// Service содержит бизнес-логику для работы с короткими ссылками.
// Управляет созданием, получением и удалением URL через репозиторий.
type Service struct {
	repository urlrepository.RepositoryInterface
	logger     *zap.SugaredLogger
	taskCh     chan deleteTask
	wg         sync.WaitGroup
}

// NewService создает новый экземпляр Service.
// Принимает репозиторий для работы с данными и логгер.
// Запускает 5 воркеров для обработки задач на удаление.
func NewService(repository urlrepository.RepositoryInterface, logger *zap.SugaredLogger) *Service {
	s := &Service{
		repository: repository,
		logger:     logger.With(zap.String("component", "service")),
		taskCh:     make(chan deleteTask, 100),
	}

	s.startWorkers(5)
	return s
}

// startWorkers запускает указанное количество воркеров для обработки задач на удаление.
func (s *Service) startWorkers(count int) {
	for i := 0; i < count; i++ {
		s.wg.Add(1)
		go func(workerID int) {
			defer s.wg.Done()
			s.workerLoop(workerID)
		}(i)
	}
	s.logger.Infow("Delete workers started", "count", count)
}

// workerLoop обрабатывает задачи на удаление из канала taskCh.
// Каждый воркер обрабатывает задачи последовательно.
func (s *Service) workerLoop(workerID int) {
	for task := range s.taskCh {
		if err := s.repository.BatchDelete(task.shortIDs, task.userID); err != nil {
			s.logger.Warn("Batch delete failed",
				"worker_id", workerID,
				"error", err)
		}
	}
	s.logger.Infow("Worker stopped", "worker_id", workerID)
}

// Stop останавливает все воркеры и ожидает их завершения.
// Закрывает канал задач и ждет окончания обработки.
func (s *Service) Stop() {
	close(s.taskCh)
	s.wg.Wait()
	s.logger.Info("All delete workers stopped")
}

// SetUrl создает короткую ссылку для оригинального URL.
// Принимает оригинальный URL и идентификатор пользователя.
// Возвращает созданную модель Url или ошибку.
// Если оригинальный URL уже существует, возвращает существующую ссылку.
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

// GetUserURLs возвращает все ссылки, созданные пользователем.
// Принимает идентификатор пользователя.
// Возвращает слайс URL или ошибку.
func (service *Service) GetUserURLs(userID string) (*[]model.Url, error) {
	service.logger.Infow("Getting user URLs", "user_id", userID)
	return service.repository.GetUserURLs(userID)
}

// SetBatch создает несколько коротких ссылок одновременно.
// Принимает слайс запросов на создание и идентификатор пользователя.
// Возвращает слайс созданных ссылок с идентификаторами корреляции.
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

// GetFullUrl возвращает оригинальный URL по короткому идентификатору.
// Принимает короткий идентификатор ссылки.
// Возвращает оригинальный URL или ошибку.
func (service *Service) GetFullUrl(short string) (string, error) {
	url, err := service.repository.GetOriginalByShort(short)
	if err != nil {
		return "", err
	}
	service.logger.Infow("URL is finded")
	return url, nil
}

// GenerateShort генерирует случайный короткий идентификатор.
// Использует криптографически безопасный генератор случайных чисел.
// Возвращает строку из 8 символов.
func (service *Service) GenerateShort() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

// DeleteUserURLs добавляет задачу на удаление ссылок в очередь.
// Принимает список коротких идентификаторов и идентификатор пользователя.
// Задача будет обработана асинхронно одним из воркеров.
func (s *Service) DeleteUserURLs(shortIDs []string, userID string) {
	if len(shortIDs) == 0 {
		return
	}

	go func() {
		s.taskCh <- deleteTask{
			shortIDs: shortIDs,
			userID:   userID,
		}
	}()

	s.logger.Infow("Delete task queued", "count", len(shortIDs), "user_id", userID)
}

// PingDb проверяет доступность базы данных.
// Возвращает ошибку если база данных недоступна.
func (service *Service) PingDb() error {
	return service.repository.Ping()
}
