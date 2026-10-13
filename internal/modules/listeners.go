package modules

import (
	"fmt"
	
	"github.com/keyvie/minecraft-bot/internal/core"
)

type listeners struct{}
var Listeners = &listeners{}

func (l *listeners) Add(container *core.Container, listener *core.Listener) error {
	key := listener.Key()
	if _, exists := container.Listeners[key]; exists {
		return fmt.Errorf("listener for event %s with index %d already exists", key.Name, key.Index)
	}
	container.Listeners[key] = listener
	return nil
}

func (l *listeners) Remove(container *core.Container, key core.ListenerKey) error {
	if _, exists := container.Listeners[key]; !exists {
		return fmt.Errorf("listener for event %s with index %d does not exist", key.Name, key.Index)
	}
	delete(container.Listeners, key)
	return nil
}

func (l *listeners) Get(container *core.Container, key core.ListenerKey) (*core.Listener, error) {
	if listener, exists := container.Listeners[key]; exists {
		return listener, nil
	}
	return nil, fmt.Errorf("listener for event %s with index %d does not exist", key.Name, key.Index)
}

func (l *listeners) List(container *core.Container) []*core.Listener {
	var listeners []*core.Listener
	for _, listener := range container.Listeners {
		listeners = append(listeners, listener)
	}
	return listeners
}