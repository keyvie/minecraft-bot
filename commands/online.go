package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"

	"github.com/keyvie/minecraft-bot/internal/core"
	"github.com/keyvie/minecraft-bot/internal/rpc"
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
		resp, err := c.State.RPC.Call(context.Background(), "minecraft:players")
		if err != nil {
			return err
		}

		var players []rpc.Player
		if err := json.Unmarshal(resp.Result, &players); err != nil {
			return fmt.Errorf("invalid response: %w", err)
		}

		responseText := []string{"**Online Players**"}
		for _, player := range players {
			responseText = append(responseText, fmt.Sprintf("> **%s** (`%s`)", player.Name, player.ID))
		}

		if err := event.CreateMessage(discord.MessageCreate{
			Content: strings.Join(responseText, "\n"),
		}); err != nil {
			return err
		}
		return nil
	},
)
