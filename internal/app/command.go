package app

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type Command interface {
	Name() string
	Execute(view View, screen tcell.Screen, ev *tcell.EventKey) (bool, error)
}

func RegisterCommand(name string, command Command) {
	if _, exists := commands[name]; exists {
		panic(fmt.Sprintf("command %s already registered", name))
	}

	commands[name] = command
}

func GetCommand(name string) Command {
	command, ok := commands[name]
	if !ok {
		panic(fmt.Sprintf("command %s not found", name))
	}

	return command
}

var commands = make(map[string]Command)
