package listeners

import (
	"fmt"
	"github.com/disgoorg/disgo/events"

	"github.com/keyvie/minecraft-bot/internal/core"
)

var Ready = core.NewListener(func(container *core.Container, event *events.Ready) error {
	fmt.Println("Bot is ready!")
	return nil
})
