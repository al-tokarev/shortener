package listeners

type Listener interface {
	Update(event interface{}) error
	Close() error
}
