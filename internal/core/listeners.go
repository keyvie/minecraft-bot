package core

import (
	"reflect"

	"github.com/disgoorg/disgo/bot"
)

type ListenerKey struct {
	Name  string
	Index int
}

type Listener[S any] struct {
	Name    string
	Index   int
	Execute func(container *Container[S], event bot.Event) error
}

func (l *Listener[S]) Key() ListenerKey {
	return ListenerKey{Name: l.Name, Index: l.Index}
}

func NewListener[E bot.Event, S any](fn func(container *Container[S], event E) error, index ...int) *Listener[S] {
	idx := 0
	if len(index) > 0 {
		idx = index[0]
	}

	return &Listener[S]{
		Name:  reflect.TypeFor[E]().Elem().Name(),
		Index: idx,
		Execute: func(container *Container[S], event bot.Event) error {
			return fn(container, event.(E))
		},
	}
}
