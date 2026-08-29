package observer

import (
	"sync"

	"github.com/al-tokarev/shortener/internal/observer/events"
	"github.com/al-tokarev/shortener/internal/observer/listeners"
	"go.uber.org/zap"
)

type Dispatcher struct {
	listeners  map[string][]listeners.Listener
	logger     *zap.SugaredLogger
	workerPool chan struct{}
	wg         sync.WaitGroup
}

func NewDispatcher(l *zap.SugaredLogger) *Dispatcher {
	return &Dispatcher{
		listeners:  make(map[string][]listeners.Listener),
		logger:     l.With(zap.String("component", "event dispatcher")),
		workerPool: make(chan struct{}, 10),
		wg:         sync.WaitGroup{},
	}
}

func (d *Dispatcher) Subscribe(eventName string, listener listeners.Listener) {
	d.listeners[eventName] = append(d.listeners[eventName], listener)
}

func (d *Dispatcher) Dispatch(event events.Event) {
	eName := event.GetName()
	if ls, ok := d.listeners[eName]; ok {
		for _, l := range ls {
			d.wg.Add(1)

			go func(listener listeners.Listener) {
				defer d.wg.Done()

				d.workerPool <- struct{}{}
				defer func() { <-d.workerPool }()

				if err := listener.Update(event); err != nil {
					d.logger.Warnw("Error by listener update", "err", err)
				}
			}(l)
		}
	}
}

func (d *Dispatcher) Close() error {
	d.wg.Wait()

	for _, listeners := range d.listeners {
		for _, l := range listeners {
			err := l.Close()
			if err != nil {
				d.logger.Warnw("Error by listener close", "err", err)
			}
		}
	}
	return nil
}
