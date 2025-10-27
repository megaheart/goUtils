package event

import "errors"

// EventManager manages a list of listeners and triggers them with data
type DataEventManager[T any] struct {
	listeners []func(data T)
}

// AddListener adds a listener to the list
func (em *DataEventManager[T]) AddListener(listener func(data T)) error {
	if listener == nil {
		return errors.New("Listener is nil")
	}
	em.listeners = append(em.listeners, listener)
	return nil
}

// Trigger triggers all listeners with data
func (em *DataEventManager[T]) Trigger(data T) {
	for _, listener := range em.listeners {
		listener(data)
	}
}

func NewDataEventManager[T any]() *DataEventManager[T] {
	em := new(DataEventManager[T])
	em.listeners = make([]func(data T), 0)

	return em
}

// EventManager manages a list of listeners and triggers them with data
type EventManager struct {
	listeners []func()
}

// AddListener adds a listener to the list
func (em *EventManager) AddListener(listener func()) error {
	if listener == nil {
		return errors.New("Listener is nil")
	}
	em.listeners = append(em.listeners, listener)
	return nil
}

// Trigger triggers all listeners with data
func (em *EventManager) Trigger() {
	for _, listener := range em.listeners {
		listener()
	}
}

func NewEventManager() *EventManager {
	em := new(EventManager)
	em.listeners = make([]func(), 0)

	return em
}
