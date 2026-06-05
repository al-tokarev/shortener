package urlrepository

import "github.com/al-tokarev/shortener/internal/model"

var storageUrl = make(map[string]model.Url)
var lastId int

func (repository *Repository) saveLocal(url *model.Url) {
	storageUrl[url.ShortUrl] = *url
	lastId = url.Uuid
}

func (repository *Repository) saveLocalBatch(urls *[]model.Url) {
	for _, url := range *urls {
		storageUrl[url.ShortUrl] = url
		lastId = url.Uuid
	}
}
