package urlrepository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/al-tokarev/shortener/internal/config"
	"go.uber.org/zap"
)

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

func (repository *Repository) newFileUrlCreator() (*urlFileCreator, error) {
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

func (repository *Repository) newUrlReader() (*urlFileReader, error) {
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

func (reader *urlFileReader) read() (*Url, error) {
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

func (reader *urlFileReader) Close() error {
	return reader.f.Close()
}
