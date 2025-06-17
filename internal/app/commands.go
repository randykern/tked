package app

import "github.com/gdamore/tcell/v2"

type CommandExit struct{}

func (c *CommandExit) Name() string { return "exit" }

func (c *CommandExit) Execute(view View, screen tcell.Screen, ev *tcell.EventKey) (bool, error) {
	return true, nil
}

func RegisterCommands() {
	RegisterCommand("exit", &CommandExit{})
}
