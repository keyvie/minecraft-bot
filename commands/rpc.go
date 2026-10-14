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
	"github.com/keyvie/minecraft-bot/internal/rpc"
	"github.com/keyvie/minecraft-bot/internal/state"
)

var RPC = core.NewApplicationCommand(
	&discord.SlashCommandCreate{
		Name:                     "rpc",
		Description:              "Run RPC",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionAdministrator),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{
				Name:        "method",
				Description: "The RPC method to call, use `rpc.discover` to see all available methods",
				Required:    true,
			},
			discord.ApplicationCommandOptionString{
				Name:        "params",
				Description: `JSON Example: {"max": 20}`,
				Required:    false,
			},
		},
	},
	func(
		c *core.Container[state.Public],
		event *events.ApplicationCommandInteractionCreate,
	) error {
		d := event.SlashCommandInteractionData()
		method := d.String("method")

		var params map[string]any
		if raw := d.String("params"); raw != "" {
			if err := json.Unmarshal([]byte(raw), &params); err != nil {
				return fmt.Errorf("Cant parse params: %w", err)
			}
		}

		var resp rpc.Response
		var err error
		if params != nil {
			resp, err = c.State.RPC.Call(context.Background(), method, params)
		} else {
			resp, err = c.State.RPC.Call(context.Background(), method)
		}
		if err != nil {
			return err
		}

		var pretty bytes.Buffer
		if err := json.Indent(&pretty, resp.Result, "", "  "); err != nil {
			return fmt.Errorf("failed to format RPC response: %w", err)
		}

		if err := event.CreateMessage(discord.MessageCreate{
			Flags: discord.MessageFlagEphemeral,
			Files: []*discord.File{
				{
					Name:   "response.json",
					Reader: bytes.NewReader(pretty.Bytes()),
				},
			},
		}); err != nil {
			return err
		}
		return nil
	},
)
