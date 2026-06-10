package urlrepository

import "github.com/al-tokarev/shortener/internal/model"

func (repository *Repository) saveLocal(url *model.Url) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	repository.storageUrl[url.ShortUrl] = *url
	repository.lastId = url.Uuid
}

func (repository *Repository) saveLocalBatch(urls *[]model.Url) {
	repository.mutex.Lock()
	defer repository.mutex.Unlock()

	for _, url := range *urls {
		repository.storageUrl[url.ShortUrl] = url
		repository.lastId = url.Uuid
	}
}
