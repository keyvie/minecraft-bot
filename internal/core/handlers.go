package core

import (
	"log/slog"
	"reflect"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func AddHandlers[S any](container *Container[S]) {
	container.Client.AddEventListeners(bot.NewListenerFunc(func(event bot.Event) {
		switch e := event.(type) {
		case *events.ApplicationCommandInteractionCreate:
			key := ApplicationCommandKey{
				Name: e.ApplicationCommandInteraction.Data.CommandName(),
				Type: e.ApplicationCommandInteraction.Data.Type(),
			}
			if applicationCommand, exists := container.ApplicationCommands[key]; exists {
				if err := applicationCommand.Execute(container, e); err != nil {
					message := discord.MessageCreate{
						Content: "Error: " + err.Error(),
						Flags:   discord.MessageFlagEphemeral,
					}

					if err := e.CreateMessage(message); err != nil {
						_, followupErr := event.Client().Rest.UpdateInteractionResponse(
							e.ApplicationID(),
							e.Token(),
							discord.MessageUpdate{
								Content: &message.Content,
							},
						)
						if followupErr != nil {
							slog.Error("Failed to send error",
								"initial", err,
								"followup", followupErr,
							)
						}
					}
				}
				return
			}
		}

		name := reflect.TypeOf(event).Elem().Name()
		if name == "" {
			return
		}

		for _, listener := range container.Listeners {
			if listener.Name != name {
				continue
			}

			if err := listener.Execute(container, event); err != nil {
				slog.Error(err.Error())
			}
		}
	}))
}
