package gop2p

import (
	"errors"
	"sync"
)

// Handler type involves function to events handling.
type Handler func(d []byte, p Peer)

// eventBus struct contains a list handlers associated with its trigger
// definition, and mutex to control handlers access.
type eventBus struct {
	h map[string]Handler
	m *sync.Mutex
}

// newEventBus function intializes eventBus struct.
func newEventBus() *eventBus {
	return &eventBus{
		h: make(map[string]Handler),
		m: &sync.Mutex{},
	}
}

// on function add new handler to its trigger checking if already exists
// any handler for that trigger. If exists, rais an error, else append
// trigger-handler pair to eventBus handlers list.
func (eb *eventBus) on(t string, f Handler) error {
	eb.m.Lock()
	defer eb.m.Unlock()

	if _, exists := eb.h[t]; exists {
		return errors.New("event handler already defined")
	}

	eb.h[t] = f
	return nil
}

// emit function fires event trigger calling its associated handler,
// if that handler exists.
func (eb *eventBus) emit(t string, d []byte, p Peer) {
	eb.m.Lock()
	f, ok := eb.h[t]
	eb.m.Unlock()

	if ok {
		f(d, p)
	}
}
