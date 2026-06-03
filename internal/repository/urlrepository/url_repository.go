package urlrepository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/al-tokarev/shortener/internal/config"
	"go.uber.org/zap"
)

var ErrShortURLAlreadyExists = errors.New("short URL already exists")

type Repository struct {
	logger *zap.SugaredLogger
	conn   *sql.DB
}

type Url struct {
	Uuid        int    `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

func NewRepository(conn *sql.DB, logger *zap.SugaredLogger) *Repository {
	return &Repository{
		logger: logger.With(zap.String("component", "repository")),
		conn:   conn,
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

func (repository *Repository) Save(url *Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		repository.logger.Warn("Error by add url", err)
		return err
	}

	if _, ok := storageUrl[url.ShortUrl]; ok {
		return ErrShortURLAlreadyExists
	}

	creator, err := repository.newFileUrlCreator()
	if err != nil {
		repository.logger.Warn("Error by create creator", err)
		return err
	}
	defer creator.Close()
	err = creator.add(data)
	if err != nil {
		repository.logger.Warn("Err by write url to file")
	}

	repository.saveLocal(url)
	return nil
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

// БД

func (repository *Repository) Ping() error {
	repository.logger.Infow("Try db connection", "db", config.Options.DatabaseDSN)
	return repository.conn.Ping()
}
