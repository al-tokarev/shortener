package urlservices

import (
	"math/rand"

	"github.com/al-tokarev/shortener/internal/logger"
	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
)

func SetUrl(short string, original string) {
	logger.Sugar.Infow("Create new url", "short", short, "original", original)

	reader, err := urlrepository.NewUrlReader()
	if err != nil {
		logger.Sugar.Debug("Error by create reader", err)
		return
	}
	defer reader.Close()

	creator, err := urlrepository.NewUrlCreator()
	if err != nil {
		logger.Sugar.Debug("Error by create creator", err)
		return
	}
	defer creator.Close()

	url := urlrepository.Url{
		Uuid:        reader.GetLastId() + 1,
		ShortUrl:    short,
		OriginalUrl: original,
	}

	creator.Add(&url)
}

func GetFullUrl(short string) (string, bool) {
	ok := false
	urlReader, err := urlrepository.NewUrlReader()
	if err != nil {
		logger.Sugar.Debug("Error by create reader", err)
		return "", false
	}
	defer urlReader.Close()

	url := urlReader.GetUrlByShort(&short)
	if url != "" {
		ok = true
	}

	return url, ok
}

func GenerateShort() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var bytesId = make([]byte, 8)
	for i := range bytesId {
		bytesId[i] = letters[rand.Intn(len(letters))]
	}

	if _, ok := GetFullUrl(string(bytesId)); ok {
		return GenerateShort()
	}
	return string(bytesId)
}
