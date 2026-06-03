package urlrepository

var storageUrl = make(map[string]Url)
var lastId int

func (repository *Repository) saveLocal(url *Url) {
	storageUrl[url.ShortUrl] = *url
}
