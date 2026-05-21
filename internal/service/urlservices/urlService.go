package urlservices

import (
	"math/rand"

	"github.com/al-tokarev/shortener/internal/repository/urlrepository"
)

func SetUrl(short string, url string) {
	urlrepository.StorageURL[short] = url
}

func GetFullUrl(short string) (string, bool) {
	url, ok := urlrepository.StorageURL[short]
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
