package urlrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/al-tokarev/shortener/internal/model"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

var ErrShortURLAlreadyExists = errors.New("Short URL already exists")
var ErrOriginalURLAlreadyExists = errors.New("Original URL already exists")
var ErrURLNotFound = errors.New("URL is not found")

type DbRepository struct {
	logger *zap.SugaredLogger
	conn   *sql.DB
}

func NewDbRepository(conn *sql.DB, logger *zap.SugaredLogger) *DbRepository {
	return &DbRepository{
		logger: logger.With(zap.String("component", "repository")),
		conn:   conn,
	}
}

func (repository *DbRepository) InitializeStorage() error {
	return nil
}

// ЗАПИСЬ

func (repository *DbRepository) Save(url *model.Url) error {
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

func (repository *DbRepository) SaveBatch(urls *[]model.Url) error {
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

// ПОЛУЧЕНИЕ

func (repository *DbRepository) GetOriginalByShort(short string) (string, error) {
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
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		repository.logger.Warn("SQL error by select", err.Error())
		return "", err
	}
	return url.OriginalUrl, nil
}

func (repository *DbRepository) GetByOriginal(original string) (*model.Url, error) {
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

func (repository *DbRepository) GetLastId() int {
	return 0
}

// БД

func (repository *DbRepository) Ping() error {
	repository.logger.Infow("Try db connection...")
	return repository.conn.Ping()
}
