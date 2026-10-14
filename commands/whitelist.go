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

var Whitelist = core.NewApplicationCommand(
	&discord.SlashCommandCreate{
		Name:                     "whitelist",
		Description:              "whitelist command",
		DefaultMemberPermissions: omit.NewPtr(discord.PermissionAdministrator),
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "add",
				Description: "Add a player to the whitelist",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "username",
						Description: "The minecraft username to add to the whitelist",
						Required:    true,
					},
					discord.ApplicationCommandOptionUser{
						Name:        "user",
						Description: "That user's Discord account",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "remove",
				Description: "Remove a player from the whitelist",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "username",
						Description: "The minecraft username to remove from the whitelist",
						Required:    false,
					},
					discord.ApplicationCommandOptionUser{
						Name:        "user",
						Description: "That user's Discord account",
						Required:    false,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "list",
				Description: "List all players on the whitelist",
			},
		},
	},
	func(
		c *core.Container[state.Public],
		event *events.ApplicationCommandInteractionCreate,
	) error {
		data := event.SlashCommandInteractionData()

		switch *data.SubCommandName {
		case "add":
			// Handle add subcommand
		case "remove":
			// Handle remove subcommand
		case "list":
			resp, err := c.State.RPC.Call(context.Background(), "minecraft:allowlist")
			if err != nil {
				return err
			}

			var pretty bytes.Buffer
			if err := json.Indent(&pretty, resp.Result, "", "  "); err != nil {
				return fmt.Errorf("failed to format RPC response: %w", err)
			}

			if err := event.CreateMessage(discord.MessageCreate{
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
		}
		return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
	},
)
