package observer

import (
	"github.com/al-tokarev/shortener/internal/observer/events"
	"github.com/al-tokarev/shortener/internal/observer/listeners"
)

type DispatcherInterface interface {
	Subscribe(eventName string, listener listeners.Listener)
	Dispatch(event events.Event)
	Close() error
}
