package urlrepository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/al-tokarev/shortener/internal/config"
	"go.uber.org/zap"
)

type Repository struct {
	logger *zap.SugaredLogger
}

type urlCreator struct {
	f      *os.File
	w      *bufio.Writer
	logger *zap.SugaredLogger
}

type urlReader struct {
	f      *os.File
	s      *bufio.Scanner
	logger *zap.SugaredLogger
}

type Url struct {
	Uuid        int    `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

var storageUrl = make(map[string]Url)

var lastId int

func NewRepository(logger *zap.SugaredLogger) *Repository {
	return &Repository{
		logger: logger.With(zap.String("component", "repository")),
	}
}

// ИНИЦИАЛИЗАЦИЯ

func (repository *Repository) InitializeStorage() error {
	reader, err := repository.newUrlReader()
	if err != nil {
		return err
	}
	defer reader.Close()

	for {
		url, err := reader.read()
		if url == nil {
			break
		}
		if err != nil {
			return err
		}

		storageUrl[url.ShortUrl] = *url
		lastId = url.Uuid
	}

	repository.logger.Info("Repository is initialize")
	return nil
}

// ЗАПИСЬ

func (repository *Repository) NewUrlCreator() (*urlCreator, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &urlCreator{
		f:      file,
		w:      bufio.NewWriter(file),
		logger: repository.logger.With(zap.String("component", "url creator")),
	}, nil
}

func (creator *urlCreator) Add(url *Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		creator.logger.Warn("Error by add url", err)
		return err
	}
	if _, err := creator.w.Write(data); err != nil {
		return err
	}
	if err := creator.w.WriteByte('\n'); err != nil {
		return err
	}

	storageUrl[url.ShortUrl] = *url

	return creator.w.Flush()
}

func (creator *urlCreator) Close() error {
	return creator.f.Close()
}

// ЧТЕНИЕ

func (repository *Repository) newUrlReader() (*urlReader, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &urlReader{
		f:      file,
		s:      bufio.NewScanner(file),
		logger: repository.logger.With(zap.String("component", "url reader")),
	}, nil
}

func (reader *urlReader) read() (*Url, error) {
	if !reader.s.Scan() {
		return nil, reader.s.Err()
	}

	data := reader.s.Bytes()

	url := Url{}
	err := json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (reader *urlReader) Close() error {
	return reader.f.Close()
}

// ПОЛУЧЕНИЕ

func (repository *Repository) GetOriginalByShort(short string) string {
	url, ok := storageUrl[short]
	if !ok {
		return ""
	}
	return url.OriginalUrl
}

func (repository *Repository) GetLastId() int {
	return lastId
}
