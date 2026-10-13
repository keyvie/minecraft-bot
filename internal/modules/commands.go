package modules

import (
	"fmt"

	"github.com/keyvie/minecraft-bot/internal/core"
)

type commands struct {}
var Commands = &commands{}

func (c *commands) Add(container *core.Container, command *core.ApplicationCommand) error {
	key := command.Key()
	if _, exists := container.ApplicationCommands[key]; exists {
		return fmt.Errorf("command %s already exists as type %s", key.Name, key.Type)
	}
	container.ApplicationCommands[key] = command
	return nil
}

func (c *commands) Remove(container *core.Container, key core.ApplicationCommandKey) error {
	if _, exists := container.ApplicationCommands[key]; !exists {
		return fmt.Errorf("command %s does not exist as type %s", key.Name, key.Type)
	}
	delete(container.ApplicationCommands, key)
	return nil
}

func (c *commands) Get(container *core.Container, key core.ApplicationCommandKey) (*core.ApplicationCommand, error) {
	if command, exists := container.ApplicationCommands[key]; exists {
		return command, nil
	}
	return nil, fmt.Errorf("command %s does not exist as type %s", key.Name, key.Type)
}

func (c *commands) List(container *core.Container) []*core.ApplicationCommand {
	var commands []*core.ApplicationCommand
	for _, command := range container.ApplicationCommands {
		commands = append(commands, command)
	}
	return commands
}