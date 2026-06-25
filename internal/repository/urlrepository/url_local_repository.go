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
	lastId     int
	storageUrl map[string]model.Url
	mutex      sync.RWMutex
}

func NewLocalRepository(logger *zap.SugaredLogger) *LocalRepository {
	return &LocalRepository{
		logger: logger.With(zap.String("component", "repository")),
	}
}

// ИНИЦИАЛИЗАЦИЯ

func (repository *LocalRepository) InitializeStorage() error {
	reader, err := repository.newUrlReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	tmpLastId := 0
	tmpStorage := make(map[string]model.Url)
	for {
		url, err := reader.read()
		if url == nil {
			break
		}
		if err != nil {
			return err
		}

		tmpStorage[url.ShortUrl] = *url
		if tmpLastId < url.Uuid {
			tmpLastId = url.Uuid
		}
	}

	repository.mutex.Lock()
	repository.storageUrl = tmpStorage
	repository.lastId = tmpLastId
	repository.mutex.Unlock()

	repository.logger.Info("Local repository is initialize")
	return nil
}

// ЗАПИСЬ

func (repository *LocalRepository) Save(url *model.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		repository.logger.Warn("Error by add url", err)
		return err
	}

	if _, ok := repository.storageUrl[url.ShortUrl]; ok {
		return ErrShortURLAlreadyExists
	}

	// добавление в файл
	creator, err := repository.newFileUrlCreator()
	if err != nil {
		repository.logger.Warn("Error by create creator", err.Error())
		return err
	}
	defer creator.Close()
	err = creator.add(data)
	if err != nil {
		repository.logger.Warn("Err by write url to file")
	}

	// добавление в память
	repository.saveLocal(url)
	return nil
}

func (repository *LocalRepository) SaveBatch(urls *[]model.Url) error {
	// добавление в файл
	creator, err := repository.newFileUrlCreator()
	if err != nil {
		repository.logger.Warn("Error by create creator", err.Error())
		return err
	}
	defer creator.Close()

	for _, url := range *urls {
		data, err := json.Marshal(url)
		err = creator.add(data)
		if err != nil {
			repository.logger.Warn("Err by write url to file")
		}
	}
	// добавление в память
	repository.saveLocalBatch(urls)
	return nil
}

func (repository *LocalRepository) saveLocal(url *model.Url) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	repository.storageUrl[url.ShortUrl] = *url
	repository.lastId = url.Uuid
}

func (repository *LocalRepository) saveLocalBatch(urls *[]model.Url) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for _, url := range *urls {
		repository.storageUrl[url.ShortUrl] = url
		repository.lastId = url.Uuid
	}
}

// ПОЛУЧЕНИЕ

func (repository *LocalRepository) GetOriginalByShort(short string) (string, error) {
	repository.mutex.RLock()
	url, ok := repository.storageUrl[short]
	repository.mutex.RUnlock()
	if !ok {
		return "", ErrURLNotFound
	}
	return url.OriginalUrl, nil
}

func (repository *LocalRepository) GetByOriginal(original string) (*model.Url, error) {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	for _, url := range repository.storageUrl {
		if url.OriginalUrl == original {
			return &url, nil
		}
	}
	return nil, ErrURLNotFound
}

func (repository *LocalRepository) GetUserURLs(userID string) (*[]model.Url, error) {
	repository.logger.Infow("Get user URLs from local storage", "user_id", userID)

	repository.mutex.RLock()
	defer repository.mutex.RUnlock()

	var urls []model.Url
	for _, url := range repository.storageUrl {
		if url.UserID == userID {
			urls = append(urls, url)
		}
	}

	repository.logger.Infow("User URLs found", "count", len(urls))
	return &urls, nil
}

func (repository *LocalRepository) GetLastId() int {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	return repository.lastId
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

func (repository *LocalRepository) newFileUrlCreator() (*urlFileCreator, error) {
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

func (repository *LocalRepository) newUrlReader() (*urlFileReader, error) {
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

func (reader *urlFileReader) read() (*model.Url, error) {
	if !reader.s.Scan() {
		return nil, reader.s.Err()
	}

	data := reader.s.Bytes()

	url := model.Url{}
	err := json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (reader *urlFileReader) Close() error {
	return reader.f.Close()
}
