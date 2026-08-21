package observer

import (
	"github.com/al-tokarev/shortener/internal/observer/events"
	"github.com/al-tokarev/shortener/internal/observer/listeners"
)

type Dispatcher struct {
	listeners map[string][]listeners.IListener
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]listeners.IListener),
	}
}

func (d Dispatcher) Subscribe(eventName string, listener listeners.IListener) {
	d.listeners[eventName] = append(d.listeners[eventName], listener)
}

func (d Dispatcher) Dispatch(event events.IEvent) {
	eName := event.GetName()
	if listeners, ok := d.listeners[eName]; ok {
		for _, l := range listeners {
			l.Update(event)
		}
	}
}
