package core

import (
	"reflect"

	"github.com/disgoorg/disgo/bot"
)

type ListenerKey struct {
	Name 	string
	Index 	int
}

type Listener struct {
	Name      	string
	Index 		int
	Execute   	func(container *Container, event bot.Event) error
}

func (l *Listener) Key() ListenerKey {
	return ListenerKey{Name: l.Name, Index: l.Index }
}

func NewListener[E bot.Event](fn func(container *Container, event E) error, index ...int) *Listener {
	idx := 0
	if len(index) > 0 {
		idx = index[0]
	}

	return &Listener{
		Name:  reflect.TypeFor[E]().Elem().Name(),
		Index: idx,
		Execute: func(container *Container, event bot.Event) error {
			return fn(container, event.(E))
		},
	}
}