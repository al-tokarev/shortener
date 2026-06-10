package urlrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/al-tokarev/shortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

var ErrShortURLAlreadyExists = errors.New("Short URL already exists")
var ErrOriginalURLAlreadyExists = errors.New("Original URL already exists")
var ErrURLNotFound = errors.New("URL is not found")

type Repository struct {
	logger     *zap.SugaredLogger
	conn       *sql.DB
	lastId     int
	storageUrl map[string]model.Url
	mutex      sync.RWMutex
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

	repository.logger.Info("Repository is initialize")
	return nil
}

// ЗАПИСЬ

func (repository *Repository) Save(url *model.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		repository.logger.Warn("Error by add url", err)
		return err
	}

	if _, ok := repository.storageUrl[url.ShortUrl]; ok {
		return ErrShortURLAlreadyExists
	}

	// добавление в бд
	if repository.conn != nil {
		repository.logger.Info("Add url to database ...")

		dbCtx, dbCancel := context.WithCancel(context.Background())
		defer dbCancel()

		stmt, err := repository.conn.PrepareContext(dbCtx, "INSERT INTO urls (short, original) VALUES ($1,$2)")
		if err != nil {
			repository.logger.Warn("SQL error by prepare insert query", err.Error())
			return err
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(dbCtx, url.ShortUrl, url.OriginalUrl)
		if err != nil {
			repository.logger.Warn("SQL error by insert", err.Error())
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return ErrOriginalURLAlreadyExists
			}
			return err
		}
		return nil
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

func (repository *Repository) SaveBatch(urls *[]model.Url) error {
	// начинаем транзакцию
	if repository.conn != nil {
		tx, err := repository.conn.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		dbCtx, dbCancel := context.WithCancel(context.Background())
		defer dbCancel()

		valueStrings := make([]string, 0, len(*urls))
		valueArgs := make([]interface{}, 0, len(*urls)*3)
		i := 0
		for _, url := range *urls {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
			valueArgs = append(valueArgs, url.ShortUrl)
			valueArgs = append(valueArgs, url.OriginalUrl)
			i++
		}
		query := fmt.Sprintf("INSERT INTO urls (short, original) VALUES %s", strings.Join(valueStrings, ","))

		stmt, err := tx.PrepareContext(dbCtx, query)
		if err != nil {
			repository.logger.Warn("SQL error by prepare insert patch", err.Error())
			return err
		}
		defer stmt.Close()

		_, err = stmt.ExecContext(dbCtx, valueArgs...)
		if err != nil {
			repository.logger.Warn("SQL error by insert patch", err.Error())
			return err
		}
		return tx.Commit()
	}

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

// ПОЛУЧЕНИЕ

func (repository *Repository) GetOriginalByShort(short string) (string, error) {
	if repository.conn != nil {
		repository.logger.Info("Find in database ...")

		dbCtx, dbCancel := context.WithCancel(context.Background())
		defer dbCancel()

		stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT id,short,original FROM urls WHERE short = $1")
		if err != nil {
			repository.logger.Warn("SQL error by prepare select query", err.Error())
			return "", err
		}
		defer stmt.Close()

		row := stmt.QueryRowContext(dbCtx, short)
		var url model.Url
		err = row.Scan(&url.Uuid, &url.ShortUrl, &url.OriginalUrl)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			repository.logger.Warn("SQL error by select", err.Error())
			return "", err
		} else if err == nil {
			return url.OriginalUrl, nil
		}
	}
	repository.logger.Info("URL was not finded in database")
	repository.logger.Info("Find in local storage ...")

	repository.mutex.RLock()
	url, ok := repository.storageUrl[short]
	repository.mutex.RUnlock()
	if !ok {
		return "", ErrURLNotFound
	}
	return url.OriginalUrl, nil
}

func (repository *Repository) GetByOriginal(original string) (*model.Url, error) {
	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT id,short,original FROM urls WHERE original = $1")
	if err != nil {
		repository.logger.Warn("SQL error by prepare select query", err.Error())
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(dbCtx, original)
	var url model.Url
	err = row.Scan(&url.Uuid, &url.ShortUrl, &url.OriginalUrl)
	if err != nil {
		repository.logger.Warn("SQL error by select", err.Error())
		return nil, err
	}
	return &url, nil
}

func (repository *Repository) GetLastId() int {
	repository.mutex.RLock()
	defer repository.mutex.RUnlock()
	return repository.lastId
}

// БД

func (repository *Repository) Ping() error {
	repository.logger.Infow("Try db connection...")
	return repository.conn.Ping()
}
