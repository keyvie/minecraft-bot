package core

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type ApplicationCommandKey struct {
	Type discord.ApplicationCommandType
	Name string
}

type ApplicationCommand[S any] struct {
	Data    discord.ApplicationCommandCreate
	Execute func(
		container *Container[S],
		event *events.ApplicationCommandInteractionCreate,
	) error
}

func (c *ApplicationCommand[S]) Key() ApplicationCommandKey {
	return ApplicationCommandKey{Type: c.Data.Type(), Name: c.Data.CommandName()}
}

func NewApplicationCommand[S any](
	data discord.ApplicationCommandCreate,
	execute func(
		container *Container[S],
		event *events.ApplicationCommandInteractionCreate,
	) error,
) *ApplicationCommand[S] {
	return &ApplicationCommand[S]{
		Data:    data,
		Execute: execute,
	}
}
