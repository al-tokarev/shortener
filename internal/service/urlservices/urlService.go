package urlservices

import "github.com/al-tokarev/shortener/internal/repository/urlrepository"

func SetUrl(short string, url string) {
	urlrepository.StorageURL[short] = url
}

func GetFullUrl(short string) (string, bool) {
	url, ok := urlrepository.StorageURL[short]
	return url, ok
}

func GenerateShort() {

}
