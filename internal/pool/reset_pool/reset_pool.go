package resetpool

import "sync"

type Resetable interface {
	Reset()
}

type Pool[T Resetable] struct {
	p       sync.Pool
	factory func() T
}

func New[T Resetable](factory func() T) *Pool[T] {
	return &Pool[T]{
		factory: factory,
		p: sync.Pool{
			New: func() interface{} {
				return factory()
			},
		},
	}
}

func (p *Pool[T]) Get() T {
	obj := p.p.Get()
	return obj.(T)
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.p.Put(obj)
}
