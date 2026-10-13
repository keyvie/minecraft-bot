package core

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

type ApplicationCommandKey struct {
	Type discord.ApplicationCommandType
	Name string
}

type ApplicationCommand struct {
	Data   	discord.ApplicationCommandCreate
	Execute func(container *Container, event *events.ApplicationCommandInteractionCreate) error
}

func (c *ApplicationCommand) Key() ApplicationCommandKey {
	return ApplicationCommandKey{Type: c.Data.Type(), Name: c.Data.CommandName()}
}

func NewApplicationCommand(
	data discord.ApplicationCommandCreate,
	execute func(container *Container, event *events.ApplicationCommandInteractionCreate) error,
) *ApplicationCommand {
	return &ApplicationCommand{
		Data:    data,
		Execute: execute,
	}
}