package core

import (
	"github.com/disgoorg/disgo/bot"
)

type Container struct {
	Client   				*bot.Client
	Listeners 				map[ListenerKey]*Listener
	ApplicationCommands  	map[ApplicationCommandKey]*ApplicationCommand
}

type ContainerOption func(*Container)

func WithListeners(listeners map[ListenerKey]*Listener) ContainerOption {
	return func(c *Container) {
		c.Listeners = listeners
	}
}

func WithApplicationCommands(commands map[ApplicationCommandKey]*ApplicationCommand) ContainerOption {
	return func(c *Container) {
		c.ApplicationCommands = commands
	}
}

func NewContainer(client *bot.Client, options ...ContainerOption) *Container {
	c := &Container{
		Client:              client,
		Listeners:           make(map[ListenerKey]*Listener),
		ApplicationCommands: make(map[ApplicationCommandKey]*ApplicationCommand),
	}

	for _, option := range options {
		option(c)
	}

	return c
}