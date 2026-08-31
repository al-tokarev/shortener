// Пакет urlservices содержит бизнес-логику для работы с короткими ссылками.
// Обеспечивает взаимодействие между HTTP обработчиками и репозиториями данных.
package urlservices

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
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

// URLService содержит бизнес-логику для работы с короткими ссылками.
// Управляет созданием, получением и удалением URL через репозиторий.
type URLService struct {
	repository urlrepository.RepositoryInterface
	logger     *zap.SugaredLogger
	taskCh     chan deleteTask
	wg         sync.WaitGroup
}

// NewService создает новый экземпляр Service.
// Принимает репозиторий для работы с данными и логгер.
// Запускает 5 воркеров для обработки задач на удаление.
func NewService(repository urlrepository.RepositoryInterface, logger *zap.SugaredLogger) URLServiceInterface {
	s := &URLService{
		repository: repository,
		logger:     logger.With(zap.String("component", "service")),
		taskCh:     make(chan deleteTask, 100),
	}

	s.startWorkers(5)
	return s
}

// startWorkers запускает указанное количество воркеров для обработки задач на удаление.
func (s *URLService) startWorkers(count int) {
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
func (s *URLService) workerLoop(workerID int) {
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
func (s *URLService) Stop() {
	close(s.taskCh)
	s.wg.Wait()
	s.logger.Info("All delete workers stopped")
}

// SetURL создает короткую ссылку для оригинального URL.
// Принимает оригинальный URL и идентификатор пользователя.
// Возвращает созданную модель URL или ошибку.
// Если оригинальный URL уже существует, возвращает существующую ссылку.
func (service *URLService) SetURL(original string, userID string) (*model.URL, error) {
	const maxAttempts = 10
	currentAttempt := 1
	var url model.URL
	var errorCreate error
	for currentAttempt < maxAttempts {
		url = model.URL{
			UUID:        service.repository.GetLastID() + 1,
			ShortURL:    service.GenerateShort(),
			OriginalURL: original,
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

	service.logger.Infow("New url", "Original", url.OriginalURL, "Short", url.ShortURL)
	return &url, nil
}

// GetUserURLs возвращает все ссылки, созданные пользователем.
// Принимает идентификатор пользователя.
// Возвращает слайс URL или ошибку.
func (service *URLService) GetUserURLs(userID string) (*[]model.URL, error) {
	service.logger.Infow("Getting user URLs", "user_id", userID)
	return service.repository.GetUserURLs(userID)
}

// SetBatch создает несколько коротких ссылок одновременно.
// Принимает слайс запросов на создание и идентификатор пользователя.
// Возвращает слайс созданных ссылок с идентификаторами корреляции.
func (service *URLService) SetBatch(batchUrls *[]model.RequestBatchURL, userID string) (*[]model.URLBatch, error) {
	urls := []model.URL{}
	urlsBatch := []model.URLBatch{}

	lastID := service.repository.GetLastID() + 1
	for _, bURL := range *batchUrls {
		url := model.URL{
			UUID:        lastID,
			ShortURL:    service.GenerateShort(),
			OriginalURL: bURL.OriginalURL,
			UserID:      userID,
		}
		urls = append(urls, url)
		urlsBatch = append(urlsBatch, model.URLBatch{
			CorrelationID: bURL.CorrelationID,
			URL:           &url,
		})

		lastID++
	}

	errorCreate := service.repository.SaveBatch(&urls)
	if errorCreate != nil {
		return nil, errorCreate
	}

	return &urlsBatch, nil
}

// GetFullURL возвращает оригинальный URL по короткому идентификатору.
// Принимает короткий идентификатор ссылки.
// Возвращает оригинальный URL или ошибку.
func (service *URLService) GetFullURL(short string) (string, error) {
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
func (service *URLService) GenerateShort() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(b)[:8]
}

// DeleteUserURLs добавляет задачу на удаление ссылок в очередь.
// Принимает список коротких идентификаторов и идентификатор пользователя.
// Задача будет обработана асинхронно одним из воркеров.
func (s *URLService) DeleteUserURLs(shortIDs []string, userID string) {
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
func (service *URLService) PingDB() error {
	return service.repository.Ping()
}
