package core

import (
	"fmt"
	"reflect"

	"github.com/disgoorg/disgo/bot"
)

func AddHandlers(container *Container) {
	container.Client.AddEventListeners(bot.NewListenerFunc(func(event bot.Event) {
		name := reflect.TypeOf(event).Elem().Name()
		// Add logic which handles application commands

		for _, listener := range container.Listeners {
			if listener.Name != name {
				continue
			}

			if err := listener.Execute(container, event); err != nil {
				fmt.Println(err.Error())
			}
		}
	}))	
}