package commands

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"

	"github.com/keyvie/minecraft-bot/internal/core"
)

var Ping = core.NewApplicationCommand(
	&discord.SlashCommandCreate{
		Name:                     "ping",
		Description:              "Get the latency",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionAdministrator),
	},
	func(
		c *core.Container,
		event *events.ApplicationCommandInteractionCreate,
	) error {
		gatewayLatency := c.Client.Gateway.Latency()

		apiStart := time.Now()
		if err := event.CreateMessage(discord.MessageCreate{
			Content: "Pinging...",
		}); err != nil {
			return err
		}
		apiLatency := time.Since(apiStart)

		content := fmt.Sprintf(
			"🏓 **Pong!**\n"+
				"**Gateway:** `%dms`\n"+
				"**API:** `%dms`",
			gatewayLatency.Milliseconds(),
			apiLatency.Milliseconds(),
		)

		_, err := event.Client().Rest.UpdateInteractionResponse(
			event.ApplicationID(),
			event.Token(),
			discord.MessageUpdate{
				Content: &content,
			},
		)
		return err
	},
)
