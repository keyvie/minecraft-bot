package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"

	"github.com/keyvie/minecraft-bot/internal/core"
	"github.com/keyvie/minecraft-bot/internal/state"
)

var Online = core.NewApplicationCommand(
	&discord.SlashCommandCreate{
		Name:                     "online",
		Description:              "Show online players",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionsNone),
	},
	func(
		c *core.Container[state.Public],
		event *events.ApplicationCommandInteractionCreate,
	) error {
		resp, err := c.State.RPC.Call(context.Background(), "online")
		if err != nil {
			return err
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, resp.Result, "", "  "); err != nil {
			return fmt.Errorf("failed to format RPC response: %w", err)
		}

		if err := event.CreateMessage(discord.MessageCreate{
			Content: pretty.String(),
		}); err != nil {
			return err
		}
		return nil
	},
)
