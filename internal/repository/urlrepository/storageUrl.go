package urlrepository

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/al-tokarev/shortener/internal/config"
)

type urlCreator struct {
	f *os.File
	w *bufio.Writer
}

type urlReader struct {
	f *os.File
	s *bufio.Scanner
}

type Url struct {
	Uuid        int    `json:"uuid"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

// ЗАПИСЬ

func NewUrlCreator() (*urlCreator, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &urlCreator{
		f: file,
		w: bufio.NewWriter(file),
	}, nil
}

func (creator *urlCreator) Add(url *Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		return err
	}
	if _, err := creator.w.Write(data); err != nil {
		return err
	}
	if err := creator.w.WriteByte('\n'); err != nil {
		return err
	}
	return creator.w.Flush()
}

func (creator *urlCreator) Close() error {
	return creator.f.Close()
}

// ЧТЕНИЕ

func NewUrlReader() (*urlReader, error) {
	file, err := os.OpenFile(config.Options.StoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &urlReader{
		f: file,
		s: bufio.NewScanner(file),
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

func (r *urlReader) GetUrlByShort(shortId *string) string {
	url, err := r.read()
	for url != nil && err == nil {
		if url.ShortUrl == *shortId {
			return url.OriginalUrl
		}
		url, err = r.read()
	}
	return ""
}

func (r *urlReader) GetLastId() int {
	var lastUrl *Url
	url, err := r.read()
	for url != nil && err == nil {
		lastUrl = url
		url, err = r.read()
	}
	if lastUrl == nil {
		return 0
	}
	return lastUrl.Uuid
}
