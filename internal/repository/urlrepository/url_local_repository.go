package urlrepository

import (
	"bufio"
	"encoding/json"
	"os"
	"sync"

	"github.com/al-tokarev/shortener/internal/config"
	"github.com/al-tokarev/shortener/internal/model"
	"go.uber.org/zap"
)

type LocalRepository struct {
	logger     *zap.SugaredLogger
	lastID     int
	storageURL map[string]*model.URL
	creator    *urlFileCreator
	mutex      sync.RWMutex
}

func NewLocalRepository(logger *zap.SugaredLogger) *LocalRepository {
	return &LocalRepository{
		logger: logger.With(zap.String("component", "repository")),
	}
}

// ИНИЦИАЛИЗАЦИЯ

func (repository *LocalRepository) InitializeStorage() error {
	reader, err := repository.newURLReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	creator, err := repository.newFileURLCreator()
	if err != nil {
		return err
	}
	repository.creator = creator

	tmpLastID := 0
	tmpStorage := make(map[string]*model.URL)
	for {
		url, err := reader.read()
		if url == nil {
			break
		}
		if err != nil {
			return err
		}

		tmpStorage[url.ShortURL] = url
		if tmpLastID < url.UUID {
			tmpLastID = url.UUID
		}
	}

	repository.mutex.Lock()
	repository.storageURL = tmpStorage
	repository.lastID = tmpLastID
	repository.mutex.Unlock()

	repository.logger.Info("Local repository is initialize")
	return nil
}

// ЗАПИСЬ

func (repository *LocalRepository) Save(url *model.URL) error {
	data, err := json.Marshal(url)
	if err != nil {
		repository.logger.Warn("Error by add url", err)
		return err
	}

	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	if _, ok := repository.storageURL[url.ShortURL]; ok {
		return ErrShortURLAlreadyExists
	}

	err = repository.creator.add(data)
	if err != nil {
		repository.logger.Warn("Err by write url to file")
	}

	// добавление в память
	repository.storageURL[url.ShortURL] = url
	repository.lastID = url.UUID
	return nil
}

func (repository *LocalRepository) SaveBatch(urls *[]model.URL) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for i := range *urls {
		url := &(*urls)[i]

		data, err := json.Marshal(url)
		if err != nil {
			return err
		}

		if err := repository.creator.add(data); err != nil {
			return err
		}

		repository.storageURL[url.ShortURL] = url
		repository.lastID = url.UUID
	}

	return nil
}

// ПОЛУЧЕНИЕ

func (repository *LocalRepository) GetOriginalByShort(short string) (string, error) {
	repository.mutex.RLock()
	url, ok := repository.storageURL[short]
	repository.mutex.RUnlock()
	if !ok {
		return "", ErrURLNotFound
	}
	return url.OriginalURL, nil
}

func (repository *LocalRepository) GetByOriginal(original string) (*model.URL, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	for _, url := range repository.storageURL {
		if url.OriginalURL == original {
			return url, nil
		}
	}
	return nil, ErrURLNotFound
}

func (repository *LocalRepository) GetUserURLs(userID string) (*[]model.URL, error) {
	repository.logger.Infow("Get user URLs from local storage", "user_id", userID)

	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	var urls []model.URL
	for _, url := range repository.storageURL {
		if url.UserID == userID {
			urls = append(urls, *url)
		}
	}

	repository.logger.Infow("User URLs found", "count", len(urls))
	return &urls, nil
}

func (repository *LocalRepository) BatchDelete(shortIDs []string, userID string) error {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for _, shortID := range shortIDs {
		if url, ok := repository.storageURL[shortID]; ok && url.UserID == userID {
			url.IsDeleted = true
			repository.storageURL[shortID] = url
		}
	}
	return nil
}

func (repository *LocalRepository) GetLastID() int {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	return repository.lastID
}

func (repository *LocalRepository) Ping() error {
	return nil
}

type urlFileCreator struct {
	f      *os.File
	w      *bufio.Writer
	logger *zap.SugaredLogger
}

type urlFileReader struct {
	f      *os.File
	s      *bufio.Scanner
	logger *zap.SugaredLogger
}

// ЗАПИСЬ В ФАЙЛ

func (repository *LocalRepository) newFileURLCreator() (*urlFileCreator, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &urlFileCreator{
		f:      file,
		w:      bufio.NewWriter(file),
		logger: repository.logger.With(zap.String("component", "url creator")),
	}, nil
}

func (creator *urlFileCreator) add(data []byte) error {
	if _, err := creator.w.Write(data); err != nil {
		return err
	}
	if err := creator.w.WriteByte('\n'); err != nil {
		return err
	}

	return creator.w.Flush()
}

func (creator *urlFileCreator) Close() error {
	return creator.f.Close()
}

// ЧТЕНИЕ ИЗ ФАЙЛА

func (repository *LocalRepository) newURLReader() (*urlFileReader, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &urlFileReader{
		f:      file,
		s:      bufio.NewScanner(file),
		logger: repository.logger.With(zap.String("component", "url reader")),
	}, nil
}

func (reader *urlFileReader) read() (*model.URL, error) {
	if !reader.s.Scan() {
		return nil, reader.s.Err()
	}

	data := reader.s.Bytes()

	url := model.URL{}
	err := json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (reader *urlFileReader) Close() error {
	return reader.f.Close()
}
