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
var ErrURLDeleted = errors.New("URL is deleted")

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

	stmt, err := repository.conn.PrepareContext(dbCtx, "INSERT INTO urls (short, original, user_id) VALUES ($1,$2,$3)")
	if err != nil {
		repository.logger.Warn("SQL error by prepare insert query", err.Error())
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(dbCtx, url.ShortUrl, url.OriginalUrl, url.UserID)
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
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3))
		valueArgs = append(valueArgs, url.ShortUrl)
		valueArgs = append(valueArgs, url.OriginalUrl)
		valueArgs = append(valueArgs, url.UserID)
		i++
	}
	query := fmt.Sprintf("INSERT INTO urls (short, original, user_id) VALUES %s", strings.Join(valueStrings, ","))

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

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT id,short,original,is_deleted FROM urls WHERE short = $1")
	if err != nil {
		repository.logger.Warn("SQL error by prepare select query", err.Error())
		return "", err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(dbCtx, short)
	var url model.Url
	err = row.Scan(&url.Uuid, &url.ShortUrl, &url.OriginalUrl, &url.IsDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		repository.logger.Warn("SQL error by select", err.Error())
		return "", err
	}

	if url.IsDeleted {
		return "", ErrURLDeleted
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

func (repository *DbRepository) GetUserURLs(userID string) (*[]model.Url, error) {
	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT short, original FROM urls WHERE user_id = $1 ORDER BY id DESC")
	if err != nil {
		repository.logger.Warn("SQL error by prepare select user urls", zap.Error(err))
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(dbCtx, userID)
	if err != nil {
		repository.logger.Warn("SQL error by select user urls", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var urls []model.Url
	for rows.Next() {
		var userURL model.Url
		err := rows.Scan(&userURL.ShortUrl, &userURL.OriginalUrl)
		if err != nil {
			repository.logger.Warn("SQL error by scan user urls", zap.Error(err))
			return nil, err
		}
		urls = append(urls, userURL)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	repository.logger.Infow("User URLs found", "count", len(urls))
	return &urls, nil
}

func (repository *DbRepository) GetLastId() int {
	return 0
}

func (repository *DbRepository) BatchDelete(shortIDs []string, userID string) error {
	if len(shortIDs) == 0 {
		return nil
	}

	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	placeholders := make([]string, len(shortIDs))
	args := make([]interface{}, 0, len(shortIDs)*2)

	for i, shortID := range shortIDs {
		placeholders[i] = fmt.Sprintf("(short = $%d AND user_id = $%d)", i*2+1, i*2+2)
		args = append(args, shortID, userID)
	}

	query := fmt.Sprintf(`
        UPDATE urls 
        SET is_deleted = true 
        WHERE %s
    `, strings.Join(placeholders, " OR "))

	stmt, err := repository.conn.PrepareContext(dbCtx, query)
	if err != nil {
		repository.logger.Warn("Batch delete failed", zap.Error(err))
		return err
	}

	result, err := stmt.ExecContext(dbCtx, args...)
	if err != nil {
		repository.logger.Warn("Batch delete failed", zap.Error(err))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	repository.logger.Infow("Batch delete completed", "rows_affected", rowsAffected)

	return nil
}

func (repository *DbRepository) Ping() error {
	repository.logger.Infow("Try db connection...")
	return repository.conn.Ping()
}
