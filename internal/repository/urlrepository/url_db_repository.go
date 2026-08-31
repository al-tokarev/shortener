// Package urlrepository содержит реализации репозиториев для работы с URL.
// Предоставляет интерфейсы и реализации для хранения ссылок в памяти и базе данных.
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

// ErrShortURLAlreadyExists возвращается при попытке создать ссылку с уже существующим коротким идентификатором.
var ErrShortURLAlreadyExists = errors.New("Short URL already exists")

// ErrOriginalURLAlreadyExists возвращается при попытке создать ссылку с уже существующим оригинальным URL.
var ErrOriginalURLAlreadyExists = errors.New("Original URL already exists")

// ErrURLNotFound возвращается когда запрашиваемая ссылка не найдена.
var ErrURLNotFound = errors.New("URL is not found")

// ErrURLDeleted возвращается при попытке получить удаленную ссылку.
var ErrURLDeleted = errors.New("URL is deleted")

// DBRepository реализует RepositoryInterface для работы с PostgreSQL.
// Хранит подключение к базе данных и логгер.
type DBRepository struct {
	logger *zap.SugaredLogger
	conn   *sql.DB
}

// NewDBRepository создает новый экземпляр DBRepository.
// Принимает подключение к базе данных и логгер.
func NewDBRepository(conn *sql.DB, logger *zap.SugaredLogger) *DBRepository {
	return &DBRepository{
		logger: logger.With(zap.String("component", "repository")),
		conn:   conn,
	}
}

// InitializeStorage инициализирует хранилище.
// Для базы данных не требует действий, всегда возвращает nil.
func (repository *DBRepository) InitializeStorage() error {
	return nil
}

// ЗАПИСЬ

// Save сохраняет новую короткую ссылку в базе данных.
// Принимает модель Url для сохранения.
// Возвращает ErrOriginalURLAlreadyExists если оригинальный URL уже существует.
func (repository *DBRepository) Save(url *model.URL) error {
	repository.logger.Info("Add url to database ...")

	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "INSERT INTO urls (short, original, user_id) VALUES ($1,$2,$3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(dbCtx, url.ShortURL, url.OriginalURL, url.UserID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrOriginalURLAlreadyExists
		}
		return err
	}
	return nil
}

// SaveBatch сохраняет несколько коротких ссылок в базе данных одной транзакцией.
// Принимает слайс URL для сохранения.
// Возвращает ошибку если транзакция не удалась.
func (repository *DBRepository) SaveBatch(urls *[]model.URL) error {
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
		valueArgs = append(valueArgs, url.ShortURL)
		valueArgs = append(valueArgs, url.OriginalURL)
		valueArgs = append(valueArgs, url.UserID)
		i++
	}
	query := fmt.Sprintf("INSERT INTO urls (short, original, user_id) VALUES %s", strings.Join(valueStrings, ","))

	stmt, err := tx.PrepareContext(dbCtx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(dbCtx, valueArgs...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// ПОЛУЧЕНИЕ

// GetOriginalByShort возвращает оригинальный URL по короткому идентификатору.
// Принимает короткий идентификатор ссылки.
// Возвращает оригинальный URL или ошибку:
//   - ErrURLNotFound если ссылка не найдена
//   - ErrURLDeleted если ссылка удалена
func (repository *DBRepository) GetOriginalByShort(short string) (string, error) {
	repository.logger.Info("Find in database ...")

	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT id,short,original,is_deleted FROM urls WHERE short = $1")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(dbCtx, short)
	var url model.URL
	err = row.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL, &url.IsDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrURLNotFound
		}
		return "", err
	}

	if url.IsDeleted {
		return "", ErrURLDeleted
	}

	return url.OriginalURL, nil
}

// GetByOriginal возвращает модель Url по оригинальному URL.
// Принимает оригинальный URL для поиска.
// Возвращает найденную модель или ошибку.
func (repository *DBRepository) GetByOriginal(original string) (*model.URL, error) {
	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT id,short,original FROM urls WHERE original = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(dbCtx, original)
	var url model.URL
	err = row.Scan(&url.UUID, &url.ShortURL, &url.OriginalURL)
	if err != nil {
		return nil, err
	}
	return &url, nil
}

// GetUserURLs возвращает все ссылки, созданные пользователем.
// Принимает идентификатор пользователя.
// Возвращает слайс URL или ошибку.
func (repository *DBRepository) GetUserURLs(userID string) (*[]model.URL, error) {
	dbCtx, dbCancel := context.WithCancel(context.Background())
	defer dbCancel()

	stmt, err := repository.conn.PrepareContext(dbCtx, "SELECT short, original FROM urls WHERE user_id = $1 ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(dbCtx, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []model.URL
	for rows.Next() {
		var userURL model.URL
		err := rows.Scan(&userURL.ShortURL, &userURL.OriginalURL)
		if err != nil {
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

// GetLastID возвращает последний использованный идентификатор.
// Для базы данных всегда возвращает 0, так как ID генерируется автоматически.
func (repository *DBRepository) GetLastID() int {
	return 0
}

// BatchDelete помечает несколько ссылок как удаленные.
// Принимает список коротких идентификаторов и идентификатор пользователя.
// Возвращает ошибку если обновление не удалось.
func (repository *DBRepository) BatchDelete(shortIDs []string, userID string) error {
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
		return err
	}

	result, err := stmt.ExecContext(dbCtx, args...)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rowsAffected, _ := result.RowsAffected()
	repository.logger.Infow("Batch delete completed", "rows_affected", rowsAffected)

	return nil
}

// Ping проверяет доступность базы данных.
// Возвращает ошибку если база данных недоступна.
func (repository *DBRepository) Ping() error {
	repository.logger.Infow("Try db connection...")
	return repository.conn.Ping()
}
