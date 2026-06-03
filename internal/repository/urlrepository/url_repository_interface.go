package urlrepository

type RepositoryInterface interface {
	InitializeStorage() error
	Save(url *Url) error
	GetOriginalByShort(short string) string
	GetLastId() int
	Ping() error
}
