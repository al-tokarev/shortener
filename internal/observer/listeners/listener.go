package listeners

type IListener interface {
	Update(event interface{}) error
}
