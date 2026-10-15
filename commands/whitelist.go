package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/omit"

	"github.com/keyvie/minecraft-bot/internal/core"
	"github.com/keyvie/minecraft-bot/internal/database"
	"github.com/keyvie/minecraft-bot/internal/database/models"
	"github.com/keyvie/minecraft-bot/internal/mojang"
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
			username := data.String("username")
			user := data.User("user")

			// Get the UUID
			profile, err := c.State.Mojang.Profiles.GetByUsername(context.Background(), username)
			if err != nil {
				return err
			}
			uuid := mojang.FormatUUID(profile.ID)

			// Check if someone already owns that minecraft account?
			result, err := c.State.Repos.Accounts.FindOne(context.Background(), database.QueryFindOneOptions{
				Where: map[string]any{
					"uuid": uuid,
				},
			})
			if err != nil {
				return err
			}
			if result.Found {
				return fmt.Errorf("An account with that username has already been linked")
			}

			_, err = c.State.RPC.Call(context.Background(), "minecraft:allowlist/add", map[string]any{
				"add": []map[string]any{
					{
						"name": username,
						"id":   uuid,
					},
				},
			})
			if err != nil {
				return err
			}

			// Add it to the database
			_, err = c.State.Repos.Accounts.Create(context.Background(), models.Account{
				UUID:      uuid,
				DiscordID: user.ID.String(),
				CreatedAt: time.Now().UTC(),
			})
			if err != nil {
				return err
			}

			return event.CreateMessage(discord.MessageCreate{
				Flags:   discord.MessageFlagEphemeral,
				Content: fmt.Sprintf("Added %s to the whitelist", username),
			})

		case "remove":
			username := data.String("username")
			user := data.User("user")

			if username != "" {
				// Get the UUID
				profile, err := c.State.Mojang.Profiles.GetByUsername(context.Background(), username)
				if err != nil {
					return err
				}
				uuid := mojang.FormatUUID(profile.ID)

				// Check if the account exists in the database
				result, err := c.State.Repos.Accounts.FindOne(context.Background(), database.QueryFindOneOptions{
					Where: map[string]any{
						"uuid": uuid,
					},
				})
				if err != nil {
					return err
				}
				if !result.Found {
					return fmt.Errorf("No account with that username is linked")
				}

				_, err = c.State.RPC.Call(context.Background(), "minecraft:allowlist/remove", map[string]any{
					"remove": []map[string]any{
						{
							"name": username,
							"id":   uuid,
						},
					},
				})
				if err != nil {
					return err
				}

				// Remove it from the database
				_, err = c.State.Repos.Accounts.Delete(context.Background(), database.QueryDeleteOptions{
					Where: map[string]any{
						"uuid": uuid,
					},
				})
				if err != nil {
					return err
				}

				return event.CreateMessage(discord.MessageCreate{
					Flags:   discord.MessageFlagEphemeral,
					Content: fmt.Sprintf("Removed %s from the whitelist", username),
				})
			}

			if user.ID != 0 {
				// Lookup in database
				result, err := c.State.Repos.Accounts.Find(context.Background(), database.QueryFindOptions{
					Where: map[string]any{
						"discord_id": user.ID.String(),
					},
				})
				if err != nil {
					return err
				}
				if result.Total <= 0 {
					return fmt.Errorf("No account is linked to that Discord user")
				}

				successful := []string{}
				failed := []string{}
				for _, item := range result.Items {
					uuid := item.UUID

					// Get Minecraft username
					profile, err := c.State.Mojang.Profiles.GetByUUID(context.Background(), uuid)
					if err != nil {
						failed = append(failed, uuid)
						continue
					}
					username := profile.Name

					_, err = c.State.RPC.Call(context.Background(), "minecraft:allowlist/remove", map[string]any{
						"remove": []map[string]any{
							{
								"name": username,
								"id":   uuid,
							},
						},
					})
					if err != nil {
						failed = append(failed, username)
						continue
					}

					_, err = c.State.Repos.Accounts.Delete(context.Background(), database.QueryDeleteOptions{
						Where: map[string]any{
							"uuid": uuid,
						},
					})
					if err != nil {
						failed = append(failed, username)
						continue
					}
					successful = append(successful, username)
				}

				var responseText []string

				if len(successful) > 0 {
					responseText = append(responseText, fmt.Sprintf(
						"**Successful**\n%s",
						strings.Join(successful, "\n"),
					))
				}

				if len(failed) > 0 {
					responseText = append(responseText, fmt.Sprintf(
						"**Failed**\n%s",
						strings.Join(failed, "\n"),
					))
				}

				return event.CreateMessage(discord.MessageCreate{
					Flags:   discord.MessageFlagEphemeral,
					Content: strings.Join(responseText, "\n\n"),
				})
			}

			return fmt.Errorf("You need to select a user or username")

		case "list":
			resp, err := c.State.RPC.Call(context.Background(), "minecraft:allowlist")
			if err != nil {
				return err
			}

			var pretty bytes.Buffer
			if err := json.Indent(&pretty, resp.Result, "", "  "); err != nil {
				return fmt.Errorf("failed to format RPC response: %w", err)
			}

			return event.CreateMessage(discord.MessageCreate{
				Flags: discord.MessageFlagEphemeral,
				Files: []*discord.File{
					{
						Name:   "response.json",
						Reader: bytes.NewReader(pretty.Bytes()),
					},
				},
			})
		}
		return fmt.Errorf("unknown subcommand: %s", *data.SubCommandName)
	},
)
