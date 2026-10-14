package core

import (
	"github.com/disgoorg/disgo/bot"
)

type Container[S any] struct {
	Client              *bot.Client
	State               S
	Listeners           map[ListenerKey]*Listener[S]
	ApplicationCommands map[ApplicationCommandKey]*ApplicationCommand[S]
}

type ContainerOption[S any] func(*Container[S])

func WithState[S any](state S) ContainerOption[S] {
	return func(c *Container[S]) {
		c.State = state
	}
}

func WithListeners[S any](listeners map[ListenerKey]*Listener[S]) ContainerOption[S] {
	return func(c *Container[S]) {
		if listeners != nil {
			c.Listeners = listeners
		}
	}
}

func WithApplicationCommands[S any](commands map[ApplicationCommandKey]*ApplicationCommand[S]) ContainerOption[S] {
	return func(c *Container[S]) {
		if commands != nil {
			c.ApplicationCommands = commands
		}
	}
}

func NewContainer[S any](client *bot.Client, options ...ContainerOption[S]) *Container[S] {
	c := &Container[S]{
		Client:              client,
		Listeners:           make(map[ListenerKey]*Listener[S]),
		ApplicationCommands: make(map[ApplicationCommandKey]*ApplicationCommand[S]),
	}
	for _, option := range options {
		option(c)
	}
	return c
}

func NewStatelessContainer(client *bot.Client) *Container[any] {
	return NewContainer[any](client)
}
