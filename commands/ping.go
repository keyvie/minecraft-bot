package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	
	"github.com/keyvie/minecraft-bot/internal/core"
)

var Ping = core.NewApplicationCommand(
	&discord.SlashCommandCreate{
		Name:        "ping",
		Description: "Get the latency",
	},
	func(container *core.Container, event *events.ApplicationCommandInteractionCreate) error {
		return nil
	},
)